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

package unpack

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// UnpackModelKit performs the core unpacking logic for a ModelKit.
func UnpackModelKit(ctx context.Context, opts *UnpackOptions) error {
	// If an unpack directory is provided, temporarily change the working directory
	// so that unpack operations that are relative to CWD behave as expected.
	// This centralizes tar -C semantics inside the unpack library.
	if opts != nil && opts.UnpackDir != "" {
		// Ensure the directory exists
		if err := os.MkdirAll(opts.UnpackDir, 0o755); err != nil {
			return fmt.Errorf("failed to create unpack directory %s: %w", opts.UnpackDir, err)
		}
		originalWd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		if err := os.Chdir(opts.UnpackDir); err != nil {
			return fmt.Errorf("failed to change working directory to %s: %w", opts.UnpackDir, err)
		}
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				output.Debugf("Failed to restore working directory: %v", err)
			}
		}()
	}
	return unpackRecursive(ctx, opts, []string{})
}

func unpackRecursive(ctx context.Context, opts *UnpackOptions, visitedRefs []string) error {
	if len(visitedRefs) > constants.MaxModelRefChain {
		return fmt.Errorf("reached maximum number of model references: [%s]", strings.Join(visitedRefs, "=>"))
	}

	ref := opts.ModelRef
	store, err := getStoreForRef(ctx, opts)
	if err != nil {
		ref := util.FormatRepositoryForDisplay(opts.ModelRef.String())
		return fmt.Errorf("failed to find reference %s: %s", ref, err)
	}
	manifestDesc, err := store.Resolve(ctx, ref.Reference)
	if err != nil {
		return fmt.Errorf("failed to resolve reference: %w", err)
	}
	manifest, config, err := util.GetManifestAndConfig(ctx, store, manifestDesc)
	if err != nil {
		return fmt.Errorf("failed to read model: %s", err)
	}
	if config.Model != nil && util.IsModelKitReference(config.Model.Path) {
		output.Infof("Unpacking referenced modelkit %s", config.Model.Path)
		if err := unpackParent(ctx, config.Model.Path, opts, visitedRefs); err != nil {
			return err
		}
	}

	if shouldUnpackLayer(config, opts.FilterConfs) {
		if err := unpackConfig(config, opts.UnpackDir, opts.Overwrite); err != nil {
			return err
		}
	}

	// Since there might be multiple datasets, etc. we need to synchronously iterate
	// through the config's relevant field to get the correct path for unpacking
	// We need to support older ModelKits (that were packed without diffIDs and digest
	// in the config) for now, so we need to continue using the old structure.
	var modelPartIdx, codeIdx, datasetIdx, docsIdx int
	for _, layerDesc := range manifest.Layers {
		// This variable supports older-format tar layers (that don't include the
		// layer path). For current ModelKits, this will be empty
		var relPath string

		// Grab path + layer info from the config object corresponding to this layer
		var layerPath string
		var layerInfo *artifact.LayerInfo
		mediaType := constants.ParseMediaType(layerDesc.MediaType)
		switch mediaType.BaseType {
		case constants.ModelType:
			if !shouldUnpackLayer(config.Model, opts.FilterConfs) {
				continue
			}
			layerInfo = config.Model.LayerInfo
			layerPath = config.Model.Path
			output.Infof("Unpacking model %s to %s", config.Model.Name, filepath.Join(opts.UnpackDir, config.Model.Path))

		case constants.ModelPartType:
			part := config.Model.Parts[modelPartIdx]
			if !shouldUnpackLayer(part, opts.FilterConfs) {
				modelPartIdx += 1
				continue
			}
			layerInfo = part.LayerInfo
			layerPath = part.Path
			output.Infof("Unpacking model part %s to %s", part.Name, part.Path)
			modelPartIdx += 1

		case constants.CodeType:
			codeEntry := config.Code[codeIdx]
			if !shouldUnpackLayer(codeEntry, opts.FilterConfs) {
				codeIdx += 1
				continue
			}
			layerInfo = codeEntry.LayerInfo
			layerPath = codeEntry.Path
			output.Infof("Unpacking code to %s", codeEntry.Path)
			codeIdx += 1

		case constants.DatasetType:
			datasetEntry := config.DataSets[datasetIdx]
			if !shouldUnpackLayer(datasetEntry, opts.FilterConfs) {
				datasetIdx += 1
				continue
			}
			layerInfo = datasetEntry.LayerInfo
			layerPath = datasetEntry.Path
			output.Infof("Unpacking dataset %s to %s", datasetEntry.Name, datasetEntry.Path)
			datasetIdx += 1

		case constants.DocsType:
			docsEntry := config.Docs[docsIdx]
			if !shouldUnpackLayer(docsEntry, opts.FilterConfs) {
				docsIdx += 1
				continue
			}
			layerInfo = docsEntry.LayerInfo
			layerPath = docsEntry.Path
			output.Infof("Unpacking docs to %s", docsEntry.Path)
			docsIdx += 1
		}

		if layerInfo != nil {
			if layerInfo.Digest != layerDesc.Digest.String() {
				return fmt.Errorf("digest in config and manifest do not match in %s", mediaType.BaseType)
			}
			relPath = ""
		} else {
			_, relPath, err = filesystem.VerifySubpath(opts.UnpackDir, layerPath)
			if err != nil {
				return fmt.Errorf("error resolving %s path: %w", mediaType.BaseType, err)
			}
		}

		if err := unpackLayer(ctx, store, layerDesc, relPath, opts.Overwrite, opts.IgnoreExisting, mediaType.Compression); err != nil {
			return fmt.Errorf("failed to unpack: %w", err)
		}
	}
	output.Debugf("Unpacked %d model part layers", modelPartIdx)
	output.Debugf("Unpacked %d code layers", codeIdx)
	output.Debugf("Unpacked %d dataset layers", datasetIdx)
	output.Debugf("Unpacked %d docs layers", docsIdx)

	return nil
}

func unpackParent(ctx context.Context, ref string, optsIn *UnpackOptions, visitedRefs []string) error {
	if idx := getIndex(visitedRefs, ref); idx != -1 {
		cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(visitedRefs[idx:], "=>"), ref)
		return fmt.Errorf("found cycle in modelkit references: %s", cycleStr)
	}

	parentRef, _, err := util.ParseReference(ref)
	if err != nil {
		return err
	}
	opts := *optsIn
	opts.ModelRef = parentRef
	// Unpack only model, ignore code/datasets
	if len(opts.FilterConfs) == 0 {
		modelFilter, err := ParseFilter("model")
		if err != nil {
			// Shouldn't happen, ever
			return fmt.Errorf("failed to parse filter for parent modelkit: %w", err)
		}
		opts.FilterConfs = []FilterConf{*modelFilter}
	} else {
		var filterConfs []FilterConf
		for _, conf := range opts.FilterConfs {
			if conf.matchesBaseType(constants.ModelType) {
				// Drop any other base types from this filter
				conf.BaseTypes = []string{constants.ModelType}
				filterConfs = append(filterConfs, conf)
			}
		}
		// If we've filtered out all confs, we don't want anything from the parent ModelKit.
		// We have to return here, as no filters is interpreted as "unpack everything"
		if len(filterConfs) == 0 {
			return nil
		}
		opts.FilterConfs = filterConfs
	}

	return unpackRecursive(ctx, &opts, append(visitedRefs, ref))
}

func unpackConfig(config *artifact.KitFile, unpackDir string, overwrite bool) error {
	configBytes, err := config.MarshalToYAML()
	if err != nil {
		return fmt.Errorf("failed to unpack config: %w", err)
	}

	configPath := filepath.Join(unpackDir, constants.DefaultKitfileName)
	if fi, exists := filesystem.PathExists(configPath); exists {
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("failed to unpack config: path %s exists and is not a regular file", configPath)
		}
		if !overwrite {
			if fi.Size() != int64(len(configBytes)) {
				return fmt.Errorf("failed to unpack config: path %s already exists", configPath)
			}
			existingBytes, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read existing Kitfile: %w", err)
			}
			if slices.Equal(configBytes, existingBytes) {
				output.Infof("Found existing Kitfile at %s", configPath)
				return nil
			}
			return fmt.Errorf("failed to unpack: Kitfile exists and does not match model's Kitfile")
		}
	}

	output.Infof("Unpacking config to %s", configPath)
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func unpackLayer(ctx context.Context, store content.Storage, desc ocispec.Descriptor, unpackPath string, overwrite, ignoreExisting bool, compression string) error {
	rc, err := store.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("failed get layer %s: %w", desc.Digest, err)
	}
	var logger *output.ProgressLogger
	rc, logger = output.WrapUnpackReadCloser(desc.Size, rc)
	defer rc.Close()

	var cr io.ReadCloser
	var cErr error
	switch compression {
	case constants.GzipCompression, constants.GzipFastestCompression:
		cr, cErr = gzip.NewReader(rc)
	case constants.NoneCompression:
		cr = rc
	}
	if cErr != nil {
		return fmt.Errorf("error setting up decompress: %w", cErr)
	}
	defer cr.Close()
	tr := tar.NewReader(cr)

	if unpackPath != "" {
		unpackPath = filepath.Dir(unpackPath)
		if err := os.MkdirAll(unpackPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", unpackPath, err)
		}
	}

	if err := extractTar(tr, unpackPath, overwrite, ignoreExisting, logger); err != nil {
		return err
	}

	logger.Wait()
	return nil
}

func extractTar(tr *tar.Reader, extractDir string, overwrite, ignoreExisting bool, logger *output.ProgressLogger) (err error) {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		outPath := header.Name
		if extractDir != "" {
			outPath = filepath.Join(extractDir, header.Name)
		}
		// Check if the outPath is within the target directory
		_, _, err = filesystem.VerifySubpath(extractDir, outPath)
		if err != nil {
			return fmt.Errorf("illegal file path: %s: %w", outPath, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if fi, exists := filesystem.PathExists(outPath); exists {
				if !fi.IsDir() {
					return fmt.Errorf("path '%s' already exists and is not a directory", outPath)
				}
			} else {
				logger.Debugf("Creating directory %s", outPath)
				if err := os.MkdirAll(outPath, header.FileInfo().Mode()); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", outPath, err)
				}
			}

		case tar.TypeReg:
			if fi, exists := filesystem.PathExists(outPath); exists {
				if ignoreExisting {
					output.Debugf("File %s already exists; skipping", outPath)
					continue
				}
				if !overwrite {
					return fmt.Errorf("path '%s' already exists", outPath)
				}
				if !fi.Mode().IsRegular() {
					return fmt.Errorf("path '%s' already exists and is not a regular file", outPath)
				}
			}
			logger.Debugf("Unpacking file %s", outPath)
			file, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outPath, err)
			}
			defer func() {
				err = errors.Join(err, file.Close())
			}()
			written, err := io.Copy(file, tr)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", outPath, err)
			}
			if written != header.Size {
				return fmt.Errorf("could not unpack file %s", outPath)
			}

		default:
			return fmt.Errorf("unrecognized type in archive: %s", header.Name)
		}
	}
	return nil
}

func getIndex(list []string, s string) int {
	for idx, item := range list {
		if s == item {
			return idx
		}
	}
	return -1
}
