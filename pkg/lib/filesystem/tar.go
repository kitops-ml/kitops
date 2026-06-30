// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/klauspost/compress/zstd"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func saveContentLayerAsTar(ctx context.Context, localRepo local.LocalRepo, path string, mediaType mediatype.MediaType, ignore ignore.Paths) (ocispec.Descriptor, *artifact.LayerInfo, error) {
	// We want to store a gzipped tar file in store, but to do so we need a descriptor, so we have to compress
	// to a temporary file. Ideally, we can also add this to the internal store by moving the file to avoid
	// copying if possible.
	tempPath, desc, info, err := packLayerToTar(path, mediaType, ignore)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}

	defer func() {
		if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			output.Errorf("Failed to remove temporary file %s: %s", tempPath, err)
		}
	}()

	if err := saveFileToRepo(ctx, tempPath, desc, localRepo); err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}

	return desc, info, nil
}

// packLayerToTar compresses an *artifact.ModelLayer to a gzipped tar file. In order to return
// a descriptor (including hash) for the compressed file, the layer is saved to a temporary file
// on disk and must be moved to an appropriate location. It is the responsibility of the caller
// to clean up the temporary file when it is no longer needed.
func packLayerToTar(path string, mediaType mediatype.MediaType, ignore ignore.Paths) (tempFilePath string, desc ocispec.Descriptor, layerInfo *artifact.LayerInfo, err error) {
	// Clean path to ensure consistent format (./path vs path/ vs path)
	path = filepath.Clean(path)

	if layerIgnored, err := ignore.Matches(path, path); err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, err
	} else if layerIgnored {
		output.Errorf("Warning: %s layer path %s ignored by kitignore", mediaType.UserString(), path)
	}

	totalSize, err := getTotalSize(path, ignore)
	if err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("error processing %s: %w", mediaType.UserString(), err)
	}
	if totalSize == 0 {
		output.Logf(output.LogLevelWarn, "No files detected in %s layer with path %s", mediaType.UserString(), path)
	}

	tempFile, tempFileCleanup, err := cache.MkCacheFile(cache.CachePackSubdir, "kitops_layer_")
	if err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFileName := tempFile.Name()
	output.Debugf("Saving layer to temporary file %s", tempFileName)

	toClose := []io.Closer{tempFile}

	digester := digest.Canonical.Digester()
	fileWriter := io.MultiWriter(tempFile, digester.Hash())

	var pWriter *output.ProgressTar
	var pLog *output.ProgressLogger
	var diffIdDigester digest.Digester

	switch mediaType.Compression() {
	case mediatype.GzipCompression:
		compressedWriter := gzip.NewWriter(fileWriter)
		diffIdDigester = digest.Canonical.Digester()
		mw := io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		tarWriter := tar.NewWriter(mw)
		pWriter, pLog = output.TarProgress(totalSize, tarWriter)
		toClose = append(toClose, compressedWriter, pWriter)
	case mediatype.GzipFastestCompression:
		compressedWriter, err := gzip.NewWriterLevel(fileWriter, gzip.BestSpeed)
		if err != nil {
			_ = closeAll(toClose)
			tempFileCleanup()
			return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to set up gzip compression: %w", err)
		}
		diffIdDigester = digest.Canonical.Digester()
		mw := io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		tarWriter := tar.NewWriter(mw)
		pWriter, pLog = output.TarProgress(totalSize, tarWriter)
		toClose = append(toClose, compressedWriter, pWriter)
	case mediatype.ZstdCompression:
		compressedWriter, err := zstd.NewWriter(fileWriter)
		if err != nil {
			_ = closeAll(toClose)
			tempFileCleanup()
			return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to set up zstd compression: %w", err)
		}
		diffIdDigester = digest.Canonical.Digester()
		mw := io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		tarWriter := tar.NewWriter(mw)
		pWriter, pLog = output.TarProgress(totalSize, tarWriter)
		toClose = append(toClose, compressedWriter, pWriter)
	case mediatype.NoneCompression:
		tarWriter := tar.NewWriter(fileWriter)
		diffIdDigester = digester
		pWriter, pLog = output.TarProgress(totalSize, tarWriter)
		toClose = append(toClose, pWriter)
	default:
		_ = closeAll(toClose)
		tempFileCleanup()
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("unsupported compression format: %s", mediaType.Compression())
	}

	if err := writeLayerToTar(path, ignore, pWriter, pLog); err != nil {
		// Don't care about these errors since we'll be deleting the file anyways
		_ = closeAll(toClose)
		tempFileCleanup()
		return "", ocispec.DescriptorEmptyJSON, nil, err
	}
	pLog.Wait()

	// We want to make sure all writes are flushed to ensure we get the correct digest
	if err := closeAll(toClose); err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, err
	}

	tempFileInfo, err := os.Stat(tempFileName)
	if err != nil {
		tempFileCleanup()
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to stat temporary file: %w", err)
	}

	desc = ocispec.Descriptor{
		MediaType: mediaType.String(),
		Digest:    digester.Digest(),
		Size:      tempFileInfo.Size(),
	}
	if err := fillDescAnnotations(&desc, path, nil); err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, err
	}

	layerInfo = &artifact.LayerInfo{
		Digest: digester.Digest().String(),
		DiffId: diffIdDigester.Digest().String(),
	}
	return tempFileName, desc, layerInfo, nil
}

func writeLayerToTar(basePath string, ignore ignore.Paths, tarWriter *output.ProgressTar, plog *output.ProgressLogger) error {
	// Make sure target path exists; otherwise we'll miss it while walking below
	_, err := os.Stat(basePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path %s does not exist", basePath)
		}
		return err
	}

	// Utility function to decide if two paths are in the same directory tree (i.e. one is a parent of the other)
	sameDirTree := func(a, b string) bool {
		aToB, errA := filepath.Rel(a, b)
		bToA, errB := filepath.Rel(b, a)
		if errA != nil || errB != nil {
			plog.Logf(output.LogLevelWarn, "Cannot compare directories %s and %s, skipping path", a, b)
			return false
		}
		if strings.Contains(aToB, "..") && strings.Contains(bToA, "..") {
			return false
		}
		return true
	}

	err = filepath.Walk(".", func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == "." {
			return nil
		}
		// Since we're walking from the context directory, we want to skip irrelevant files (e.g. sibling directories)
		if !sameDirTree(basePath, file) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip anything that's not a regular file or directory
		if !fi.Mode().IsRegular() && !fi.Mode().IsDir() {
			if fi.Mode()&os.ModeSymlink != 0 {
				plog.Logf(output.LogLevelWarn, "Skipping symlink %s (symlinks are not supported)", file)
			} else {
				plog.Debugf("Skipping file %s (not a regular file)", file)
			}
			return nil
		}

		// Check if file should be ignored by the ignorefile/other Kitfile layers
		if shouldIgnore, err := ignore.Matches(file, basePath); err != nil {
			return fmt.Errorf("failed to match %s against ignore file: %w", file, err)
		} else if shouldIgnore {
			if !ignore.HasExclusions() && fi.IsDir() {
				plog.Debugf("Skipping directory %s: ignored", file)
				return filepath.SkipDir
			}
			plog.Debugf("Skipping file %s: ignored", file)
			return nil
		}

		if err := writeHeaderToTar(file, fi, tarWriter, plog); err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		return writeFileToTar(file, fi, tarWriter, plog)
	})
	if err != nil {
		return err
	}

	return nil
}

func writeHeaderToTar(name string, fi os.FileInfo, ptw *output.ProgressTar, plog *output.ProgressLogger) error {
	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return fmt.Errorf("failed to generate header for %s: %w", name, err)
	}
	header.Name = name
	sanitizeTarHeader(header)
	if err := ptw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	plog.Debugf("Wrote header %s to tar file", header.Name)
	return nil
}

func writeFileToTar(file string, fi os.FileInfo, ptw *output.ProgressTar, plog *output.ProgressLogger) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file for archiving: %w", err)
	}
	defer f.Close()

	if written, err := io.Copy(ptw, f); err != nil {
		return fmt.Errorf("failed to add file to archive: %w", err)
	} else if written != fi.Size() {
		return fmt.Errorf("file written to tar does not match expected size")
	}
	plog.Debugf("Wrote file %s to tar file", file)
	return nil
}

func sanitizeTarHeader(header *tar.Header) {
	// On windows, store paths linux-style (forward slashes). This is a no-op if
	// filepath.Separator is '/'
	header.Name = filepath.ToSlash(header.Name)
	// Clear fields that break reproducible tars
	header.AccessTime = time.Time{}
	header.ModTime = time.Time{}
	header.ChangeTime = time.Time{}
	header.Uid = 1000
	header.Gid = 0
	header.Uname = ""
	header.Gname = ""
}
