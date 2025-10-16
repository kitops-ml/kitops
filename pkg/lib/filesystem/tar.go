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
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

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
		output.Errorf("Warning: %s layer path %s ignored by kitignore", mediaType.BaseType, path)
	}

	totalSize, err := getTotalSize(path, ignore)
	if err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("error processing %s: %w", mediaType.BaseType, err)
	}
	if totalSize == 0 {
		output.Logf(output.LogLevelWarn, "No files detected in %s layer with path %s", mediaType.BaseType, path)
	}

	tempFile, tempFileCleanup, err := cache.MkCacheFile(cache.CachePackSubdir, "kitops_layer_")
	if err != nil {
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFileName := tempFile.Name()
	output.Debugf("Compressing layer to temporary file %s", tempFileName)

	digester := digest.Canonical.Digester()
	var diffIdDigester digest.Digester
	fileWriter := io.MultiWriter(tempFile, digester.Hash())

	var compressedWriter io.WriteCloser
	var tarWriter *tar.Writer
	switch mediaType.Compression {
	case mediatype.GzipCompression:
		compressedWriter = gzip.NewWriter(fileWriter)
		diffIdDigester = digest.Canonical.Digester()
		mw := io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		tarWriter = tar.NewWriter(mw)
	case mediatype.GzipFastestCompression:
		compressedWriter, err = gzip.NewWriterLevel(fileWriter, gzip.BestSpeed)
		if err != nil {
			return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to set up gzip compression: %w", err)
		}
		diffIdDigester = digest.Canonical.Digester()
		mw := io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		tarWriter = tar.NewWriter(mw)
	case mediatype.NoneCompression:
		tarWriter = tar.NewWriter(fileWriter)
		diffIdDigester = digester
	}
	progressTarWriter, plog := output.TarProgress(totalSize, tarWriter)

	if err := writeLayerToTar(path, ignore, progressTarWriter, plog); err != nil {
		// Don't care about these errors since we'll be deleting the file anyways
		_ = progressTarWriter.Close()
		_ = tarWriter.Close()
		if compressedWriter != nil {
			_ = compressedWriter.Close()
		}
		tempFileCleanup()
	}
	plog.Wait()

	callAndPrintError(progressTarWriter.Close, "Failed to close writer: %s")
	callAndPrintError(tarWriter.Close, "Failed to close tar writer: %s")
	if compressedWriter != nil {
		callAndPrintError(compressedWriter.Close, "Failed to close compression writer: %s")
	}

	tempFileInfo, err := tempFile.Stat()
	if err != nil {
		tempFileCleanup()
		return "", ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to stat temporary file: %w", err)
	}
	callAndPrintError(tempFile.Close, "Failed to close temporary file: %s")

	desc = ocispec.Descriptor{
		MediaType: mediaType.String(),
		Digest:    digester.Digest(),
		Size:      tempFileInfo.Size(),
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
		// Skip anything that's not a regular file or directory
		if !fi.Mode().IsRegular() && !fi.Mode().IsDir() {
			return nil
		}
		// Since we're walking from the context directory, we want to skip irrelevant files (e.g. sibling directories)
		if !sameDirTree(basePath, file) {
			if fi.IsDir() {
				return filepath.SkipDir
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
		return fmt.Errorf("error writing file: %w", err)
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
