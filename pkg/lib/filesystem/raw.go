// Copyright 2026 The KitOps Authors.
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
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func saveContentLayerAsRaw(ctx context.Context, localRepo local.LocalRepo, targetPath string, mediaType mediatype.MediaType, ignore ignore.Paths) (ocispec.Descriptor, *artifact.LayerInfo, error) {
	targetPath = filepath.Clean(targetPath)

	fi, err := os.Lstat(targetPath)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to read file %s: %w", targetPath, err)
	}
	if !fi.Mode().IsRegular() {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("file %s is not a regular file", targetPath)
	}

	tempFile, tempFileCleanup, err := cache.MkCacheFile(cache.CachePackSubdir, "kitops_layer_")
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFileCleanup()
	tempFileName := tempFile.Name()
	output.Debugf("Saving layer to temporary file %s", tempFileName)

	toClose := []io.Closer{tempFile}

	digester := digest.Canonical.Digester()
	fileWriter := io.MultiWriter(tempFile, digester.Hash())

	var diffIdDigester digest.Digester
	var finalWriter io.Writer
	switch mediaType.Compression() {
	case mediatype.GzipCompression:
		compressedWriter := gzip.NewWriter(fileWriter)
		diffIdDigester = digest.Canonical.Digester()
		finalWriter = io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		toClose = append(toClose, compressedWriter)
	case mediatype.GzipFastestCompression:
		compressedWriter, err := gzip.NewWriterLevel(fileWriter, gzip.BestSpeed)
		if err != nil {
			_ = closeAll(toClose)
			return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to set up gzip compression: %w", err)
		}
		diffIdDigester = digest.Canonical.Digester()
		finalWriter = io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		toClose = append(toClose, compressedWriter)
	case mediatype.ZstdCompression:
		compressedWriter, err := zstd.NewWriter(fileWriter)
		if err != nil {
			_ = closeAll(toClose)
			return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to set up zstd compression: %w", err)
		}
		diffIdDigester = digest.Canonical.Digester()
		finalWriter = io.MultiWriter(compressedWriter, diffIdDigester.Hash())
		toClose = append(toClose, compressedWriter)
	case mediatype.NoneCompression:
		finalWriter = fileWriter
		diffIdDigester = digester
	default:
		_ = closeAll(toClose)
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("unsupported compression format: %s", mediaType.Compression())
	}

	f, err := os.Open(targetPath)
	if err != nil {
		_ = closeAll(toClose)
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to open file %s: %w", targetPath, err)
	}
	pReader, pLogger := output.WrapReadCloser("Saving", fi.Size(), f)
	toClose = append(toClose, pReader)

	written, err := io.Copy(finalWriter, pReader)
	if err != nil {
		_ = closeAll(toClose)
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to save file %s: %w", targetPath, err)
	} else if written != fi.Size() {
		_ = closeAll(toClose)
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to read entire file %s", targetPath)
	}

	pLogger.Wait()
	// We want to make sure all writes are flushed to ensure we get the correct digest
	if err := closeAll(toClose); err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}

	tempFileInfo, err := os.Stat(tempFileName)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to stat temporary file: %w", err)
	}
	desc := ocispec.Descriptor{
		MediaType: mediaType.String(),
		Digest:    digester.Digest(),
		Size:      tempFileInfo.Size(),
	}
	if err := fillDescAnnotations(&desc, targetPath, fi); err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}
	layerInfo := &artifact.LayerInfo{
		Digest: digester.Digest().String(),
		DiffId: diffIdDigester.Digest().String(),
	}

	if err := saveFileToRepo(ctx, tempFileName, desc, localRepo); err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}

	return desc, layerInfo, nil
}
