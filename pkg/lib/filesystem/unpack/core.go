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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/external/s3api"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/klauspost/compress/zstd"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

// UnpackModelKit performs the core unpacking logic for a ModelKit.
func UnpackModelKit(ctx context.Context, opts *UnpackOptions) error {
	if opts == nil {
		return fmt.Errorf("unpack options must not be nil")
	}
	// If an unpack directory is provided, temporarily change the working directory
	// so that unpack operations that are relative to CWD behave as expected.
	// This centralizes tar -C semantics inside the unpack library.
	if opts.UnpackDir != "" {
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
	if opts.SkillOptions != nil {
		return unpackSkill(ctx, opts)
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
		ref := artifact.FormatRepositoryForDisplay(opts.ModelRef.String())
		return fmt.Errorf("failed to find reference %s: %s", ref, err)
	}

	_, manifest, kitfile, err := util.ResolveManifestAndConfig(ctx, store, ref.Reference)
	// If error is ErrNoKitfile, the manifest has been retrieved
	if err != nil && !errors.Is(err, util.ErrNoKitfile) {
		return err
	}

	if kitfile == nil {
		output.Logf(output.LogLevelWarn, "Artifact does not include Kitfile: functionality may be impacted")
		modelConfig, err := util.GetModelPackConfig(ctx, store, manifest.Config)
		if err != nil {
			return err
		}
		genconfig, err := artifact.GenerateKitfileForModelPack(manifest, modelConfig)
		if err != nil {
			return fmt.Errorf("could not process manifest: %w", err)
		}
		kitfile = genconfig
	} else {
		if kfutils.LayerMatchesAnyFilter(kitfile, opts.FilterConfs) {
			if err := unpackConfig(kitfile, opts.UnpackDir, opts.Overwrite); err != nil {
				return err
			}
		}
	}

	if err := handleRemoteData(ctx, kitfile, opts, visitedRefs); err != nil {
		return err
	}

	hasDigest, _, err := artifact.KitfileHasLayerInfo(kitfile)
	if err != nil {
		return err
	}
	if !hasDigest {
		return legacyUnpackLayers(ctx, store, manifest, kitfile, opts)
	}

	steps, err := generateUnpackPlan(manifest, kitfile, opts.FilterConfs)
	if err != nil {
		return fmt.Errorf("failed to plan unpack: %w", err)
	}
	for _, step := range steps {
		output.Infoln(step.userMessage)
		if err := unpackLayer(ctx, store, step.desc, "", opts.Overwrite, opts.IgnoreExisting, step.mediatype); err != nil {
			return fmt.Errorf("failed to unpack: %w", err)
		}
	}
	return nil
}

func handleRemoteData(ctx context.Context, config *artifact.KitFile, opts *UnpackOptions, visitedRefs []string) error {
	if config.Model != nil && artifact.IsModelKitReference(config.Model.Path) {
		modelRef, _, err := artifact.ParseReference(config.Model.Path)
		if err != nil {
			return fmt.Errorf("failed to parse remote model reference %s: %w", config.Model.Path, err)
		}
		output.Infof("Unpacking referenced ModelKit %s", config.Model.Path)
		if err := unpackRemote(ctx, modelRef, "", kfutils.BaseTypeModel, opts, visitedRefs); err != nil {
			return err
		}
	}

	// Build lists of remote references first to allow is to print a warning if remotes are present but skipped in options.
	remoteS3Datasets := map[string]s3api.S3ObjectReference{}
	remoteModelKitDatasets := map[string]*registry.Reference{}
	for _, dataset := range config.DataSets {
		if dataset.RemotePath == "" || !kfutils.LayerMatchesAnyFilter(dataset, opts.FilterConfs) {
			continue
		}
		if artifact.IsModelKitReference(dataset.RemotePath) {
			datasetRef, _, err := artifact.ParseReference(dataset.RemotePath)
			if err != nil {
				return fmt.Errorf("failed to parse remote dataset reference %s: %w", dataset.RemotePath, err)
			}
			remoteModelKitDatasets[dataset.Path] = datasetRef
		} else {
			ref, err := s3api.ParseS3ObjectReference(dataset.RemotePath, dataset.RemoteHash)
			if err != nil {
				return fmt.Errorf("failed to parse S3 object reference for dataset %s: %w", dataset.Path, err)
			}
			remoteS3Datasets[dataset.Path] = *ref
		}
	}

	// The docstring for the "include remote" option states that it is for remote datasets. This was originally
	// for S3 remote datasets only but it applies to remote datasets referenced by a ModelKit reference too.
	if !opts.IncludeRemote && (len(remoteModelKitDatasets) > 0 || len(remoteS3Datasets) > 0) {
		output.Logf(output.LogLevelWarn, "ModelKit contains remote datasets. To unpack, specify the --include-remote flag")
		return nil
	}

	for path, remoteRef := range remoteModelKitDatasets {
		output.Infof("Unpacking referenced dataset ModelKit %s to %s", remoteRef, path)
		if err := unpackRemote(ctx, remoteRef, path, kfutils.BaseTypeDatasets, opts, visitedRefs); err != nil {
			return err
		}
	}

	if len(remoteS3Datasets) > 0 {
		client, err := s3api.SetUpClient(ctx)
		if err != nil {
			return err
		}
		for path, s3Ref := range remoteS3Datasets {
			_, relPath, err := filesystem.VerifySubpath(opts.UnpackDir, path)
			if err != nil {
				return fmt.Errorf("error verifying path %s for remote reference: %w", path, err)
			}

			output.Debugf("Downloading remote dataset: Bucket: %s, Key: %s", s3Ref.Bucket, s3Ref.Key)
			if fi, exists := filesystem.PathExists(relPath); exists {
				if opts.IgnoreExisting {
					output.Debugf("File %s already exists; skipping", path)
					continue
				}
				if !opts.Overwrite {
					return fmt.Errorf("failed to unpack remote dataset: path '%s' already exists", path)
				}
				if !fi.Mode().IsRegular() {
					return fmt.Errorf("failed to unpack remote dataset: path '%s' already exists and is not a regular file", path)
				}
			}

			pathDir := filepath.Dir(relPath)
			if err := os.MkdirAll(pathDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", pathDir, err)
			}
			if err := s3api.DownloadObject(ctx, client, &s3Ref, relPath); err != nil {
				return fmt.Errorf("failed to download remote dataset for path %s: %w", path, err)
			}
			output.Infof("Downloaded remote S3 dataset for path %s", path)
		}
	}

	return nil
}

// unpackRemote recursively unpacks the contents of a referenced ModelKit, restricted to a
// single base type (e.g. "model" or "datasets"). If basePath is non-empty, the remote ModelKit
// is unpacked into that subdirectory of the current unpack dir so the referenced ModelKit's
// layer paths land beneath it
func unpackRemote(ctx context.Context, ref *registry.Reference, basePath string, baseType kfutils.BaseType, optsIn *UnpackOptions, visitedRefs []string) error {
	if idx := getIndex(visitedRefs, ref.String()); idx != -1 {
		cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(visitedRefs[idx:], "=>"), ref)
		return fmt.Errorf("found cycle in modelkit references: %s", cycleStr)
	}

	opts := *optsIn
	opts.ModelRef = ref
	// Restrict unpack to the requested base type from the referenced ModelKit.
	if len(opts.FilterConfs) == 0 {
		filter := kfutils.FilterConf{
			BaseTypes: []kfutils.BaseType{baseType},
		}
		opts.FilterConfs = []kfutils.FilterConf{filter}
	} else {
		var filterConfs []kfutils.FilterConf
		for _, conf := range opts.FilterConfs {
			if conf.MatchesBaseType(baseType) {
				// Drop any other base types from this filter
				conf.BaseTypes = []kfutils.BaseType{baseType}
				filterConfs = append(filterConfs, conf)
			}
		}
		// If we've filtered out all confs, we don't want anything from the parent ModelKit.
		// We have to return here, as no filters is interpreted as "unpack everything"
		if len(filterConfs) == 0 {
			output.Debugf("Skipping unpack of referenced ModelKit %s due to provided filters", ref)
			return nil
		}
		opts.FilterConfs = filterConfs
	}

	if basePath != "" {
		targetDir, _, err := filesystem.VerifySubpath(opts.UnpackDir, basePath)
		if err != nil {
			return fmt.Errorf("error resolving dataset path %s: %w", basePath, err)
		}
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}
		originalWd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		if err := os.Chdir(targetDir); err != nil {
			return fmt.Errorf("failed to change working directory to %s: %w", targetDir, err)
		}
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				output.Logf(output.LogLevelWarn, "Failed to restore working directory after unpacking reference ModelKit: %v", err)
			}
		}()
		opts.UnpackDir = targetDir
	}

	return unpackRecursive(ctx, &opts, append(visitedRefs, ref.String()))
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

func unpackLayer(ctx context.Context, store oras.ReadOnlyTarget, desc ocispec.Descriptor, unpackPath string, overwrite, ignoreExisting bool, mt mediatype.MediaType) error {
	rc, err := store.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("failed get layer %s: %w", desc.Digest, err)
	}
	var logger *output.ProgressLogger
	rc, logger = output.WrapReadCloser("Unpacking", desc.Size, rc)
	defer rc.Close()

	var cr io.Reader
	switch mt.Compression() {
	case mediatype.GzipCompression, mediatype.GzipFastestCompression:
		gzipReader, err := gzip.NewReader(rc)
		if err != nil {
			return fmt.Errorf("error setting up decompression: %w", err)
		}
		defer gzipReader.Close()
		cr = gzipReader
	case mediatype.ZstdCompression:
		// Note zstd.NewReader is not an io.ReadCloser by default; the Close() method does not return an error.
		zstdReader, err := zstd.NewReader(rc)
		if err != nil {
			return fmt.Errorf("error setting up decompression: %w", err)
		}
		defer zstdReader.Close()
		cr = zstdReader
	case mediatype.NoneCompression:
		cr = rc
	default:
		return fmt.Errorf("unrecognized compression format in '%s'", mt.Format())
	}

	switch mt.Format() {
	case mediatype.TarFormat:
		// Legacy behaviour for ModelKits created prior to Kit v0.5.0 -- tar layers might not include
		// all parent directories and so these need to be created. For newer ModelKits, unpackPath is empty.
		if unpackPath != "" {
			unpackPath = filepath.Dir(unpackPath)
			if err := os.MkdirAll(unpackPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", unpackPath, err)
			}
		}

		tr := tar.NewReader(cr)
		if err := extractTar(tr, unpackPath, overwrite, ignoreExisting, logger); err != nil {
			return err
		}
	case mediatype.RawFormat:
		if err := extractRawLayer(cr, desc, overwrite, ignoreExisting, logger); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unrecognized layer format in '%s'", mt.Format())
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
			file, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, header.FileInfo().Mode()&fs.ModePerm)
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

func extractRawLayer(r io.Reader, desc ocispec.Descriptor, overwrite, ignoreExisting bool, logger *output.ProgressLogger) (err error) {
	targetPath := desc.Annotations[modelspecv1.AnnotationFilepath]
	if targetPath == "" {
		return fmt.Errorf("failed to unpack raw layer: no %s annotation", modelspecv1.AnnotationFilepath)
	}
	targetPath = filepath.Clean(targetPath)

	if _, _, err := filesystem.VerifySubpath("", targetPath); err != nil {
		return fmt.Errorf("illegal file path: %s: %w", targetPath, err)
	}

	var fileMeta *modelspecv1.FileMetadata
	if fileMetaJson := desc.Annotations[modelspecv1.AnnotationFileMetadata]; fileMetaJson != "" {
		fileMeta = &modelspecv1.FileMetadata{}
		err := json.Unmarshal([]byte(fileMetaJson), fileMeta)
		if err != nil {
			return fmt.Errorf("error reading %s annotation on manifest: %w", modelspecv1.AnnotationFileMetadata, err)
		}
	} else {
		output.Debugf("no %s annotation on manifest, using defaults", modelspecv1.AnnotationFileMetadata)
	}

	if fi, exists := filesystem.PathExists(targetPath); exists {
		if ignoreExisting {
			output.Debugf("File %s already exists; skipping", targetPath)
			return
		}
		if !overwrite {
			return fmt.Errorf("path '%s' already exists", targetPath)
		}
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("path '%s' already exists and is not a regular file", targetPath)
		}
	}

	dir := filepath.Dir(targetPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directories for raw layer: %w", err)
		}
	}

	var fileMode fs.FileMode = 0644 // default: rw-r--r--
	if fileMeta != nil {
		fileMode = fs.FileMode(fileMeta.Mode) & fs.ModePerm
	}
	logger.Debugf("Unpacking file %s", targetPath)
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fileMode)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	written, err := io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetPath, err)
	}
	if fileMeta != nil && written != fileMeta.Size {
		return fmt.Errorf("unpacked size of file %s does not match metadata annotation", targetPath)
	}

	return nil
}
