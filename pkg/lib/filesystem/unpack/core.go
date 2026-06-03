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
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/external/s3api"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/lib/skill"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
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
	manifestDesc, err := store.Resolve(ctx, ref.Reference)
	if err != nil {
		return fmt.Errorf("failed to resolve reference: %w", err)
	}

	manifest, err := util.GetManifest(ctx, store, manifestDesc)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %s", err)
	}
	config, err := util.GetKitfileForManifest(ctx, store, manifest)

	// A flag to determine whether we should print a warning about remote datasets when opts.IncludeRemote is false
	containsRemoteDatasets := false

	if err != nil {
		if !errors.Is(err, util.ErrNoKitfile) {
			return err
		}
		output.Logf(output.LogLevelWarn, "Could not get Kitfile: %s", err)
		output.Logf(output.LogLevelWarn, "Functionality may be impacted")
		// TODO: we can probably _also_ handle getting the model-spec config and using it here
		genconfig, err := artifact.GenerateKitfileForModelPack(manifest)
		if err != nil {
			return fmt.Errorf("could not process manifest: %w", err)
		}
		config = genconfig
	} else {
		// These steps only make sense if we have a legitimate Kitfile available
		if config.Model != nil && artifact.IsModelKitReference(config.Model.Path) {
			modelRef, _, err := artifact.ParseReference(config.Model.Path)
			if err != nil {
				return fmt.Errorf("failed to parse remote model reference %s: %w", config.Model.Path, err)
			}
			output.Infof("Unpacking referenced ModelKit %s", config.Model.Path)
			if err := unpackRemote(ctx, modelRef, "", kitfile.BaseTypeModel, opts, visitedRefs); err != nil {
				return err
			}
		}
		for _, dataset := range config.DataSets {
			if !artifact.IsModelKitReference(dataset.RemotePath) {
				continue
			}
			if !kitfile.LayerMatchesAnyFilter(dataset, opts.FilterConfs) {
				continue
			}
			// The docstring for the "include remote" option states that it is for remote datasets. This was originally
			// for S3 remote datasets only but it applies to remote datasets referenced by a ModelKit reference too.
			if opts.IncludeRemote {
				datasetRef, _, err := artifact.ParseReference(dataset.RemotePath)
				if err != nil {
					return fmt.Errorf("failed to parse remote dataset reference %s: %w", dataset.RemotePath, err)
				}
				output.Infof("Unpacking referenced dataset ModelKit %s to %s", dataset.RemotePath, dataset.Path)

				if err := unpackRemote(ctx, datasetRef, dataset.Path, kitfile.BaseTypeDatasets, opts, visitedRefs); err != nil {
					return err
				}
			}
			containsRemoteDatasets = true
		}
		if kitfile.LayerMatchesAnyFilter(config, opts.FilterConfs) {
			if err := unpackConfig(config, opts.UnpackDir, opts.Overwrite); err != nil {
				return err
			}
		}
	}

	// Since there might be multiple datasets, etc. we need to synchronously iterate
	// through the config's relevant field to get the correct path for unpacking
	// We need to support older ModelKits (that were packed without diffIDs and digest
	// in the config) for now, so we need to continue using the old structure.
	var modelPartIdx, codeIdx, datasetIdx, docsIdx, promptIdx int
	for _, layerDesc := range manifest.Layers {
		// This variable supports older-format tar layers (that don't include the
		// layer path). For current ModelKits, this will be empty
		var relPath string

		mediaType, err := mediatype.ParseMediaType(layerDesc.MediaType)
		if err != nil {
			// We may encounter unknown media types while unpacking ModelPacks, e.g. we include Kitfiles
			// which are not ModelPack mediatypes
			output.Logf(output.LogLevelWarn, "Unknown media type %s: skipping unpack", layerDesc.MediaType)
			continue
		}

		// Grab path + layer info from the config object corresponding to this layer
		var layerPath string
		var layerInfo *artifact.LayerInfo
		switch mediaType.Base() {
		case mediatype.ModelBaseType:
			entry := config.Model
			if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
				continue
			}
			layerInfo, layerPath = entry.LayerInfo, entry.Path
			output.Infof("Unpacking model %s to %s", config.Model.Name, config.Model.Path)

		case mediatype.ModelPartBaseType:
			entry := config.Model.Parts[modelPartIdx]
			modelPartIdx += 1
			if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
				continue
			}
			layerInfo, layerPath = entry.LayerInfo, entry.Path
			output.Infof("Unpacking model part %s to %s", entry.Name, entry.Path)

		case mediatype.CodeBaseType:
			// Code-type layers may be either regular code or prompts
			if layerDesc.Annotations[constants.LayerSubtypeAnnotation] == constants.LayerSubtypePrompt {
				entry := config.Prompts[promptIdx]
				promptIdx += 1
				if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
					continue
				}
				layerInfo, layerPath = entry.LayerInfo, entry.Path
				output.Infof("Unpacking prompt to %s", entry.Path)
			} else {
				entry := config.Code[codeIdx]
				codeIdx += 1
				if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
					continue
				}
				layerInfo, layerPath = entry.LayerInfo, entry.Path
				output.Infof("Unpacking code to %s", entry.Path)
			}

		case mediatype.DatasetBaseType:
			// Since some datasets may be remote, we need to search the Kitfile for the next non-remote dataset
			var entry *artifact.DataSet
			for idx := datasetIdx; idx < len(config.DataSets); idx++ {
				dataset := config.DataSets[idx]
				if dataset.RemotePath != "" {
					continue
				}
				entry = &dataset
				datasetIdx = idx + 1
				break
			}
			if entry == nil {
				continue
			}
			if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
				continue
			}
			layerInfo, layerPath = entry.LayerInfo, entry.Path
			output.Infof("Unpacking dataset %s to %s", entry.Name, entry.Path)

		case mediatype.DocsBaseType:
			entry := config.Docs[docsIdx]
			docsIdx += 1
			if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
				continue
			}
			layerInfo, layerPath = entry.LayerInfo, entry.Path
			output.Infof("Unpacking docs to %s", entry.Path)

		case mediatype.ConfigBaseType:
			// ModelPacks may contain a Kitfile in their layers, which is unpacked separately
			continue

		case mediatype.UnknownBaseType:
			// Should never happen as we check earlier, but for completeness' sake:
			output.Logf(output.LogLevelWarn, "Unknown media type %s: skipping unpack", layerDesc.MediaType)
		}

		if layerInfo != nil {
			if layerInfo.Digest != layerDesc.Digest.String() {
				return fmt.Errorf("digest in config and manifest do not match in %s", mediaType.UserString())
			}
			relPath = ""
		} else {
			_, relPath, err = filesystem.VerifySubpath(opts.UnpackDir, layerPath)
			if err != nil {
				return fmt.Errorf("error resolving %s path: %w", mediaType.UserString(), err)
			}
		}

		// TODO: handle DiffIDs when unpacking layers
		if err := unpackLayer(ctx, store, layerDesc, relPath, opts.Overwrite, opts.IgnoreExisting, mediaType.Compression()); err != nil {
			return fmt.Errorf("failed to unpack: %w", err)
		}
	}

	// Handle remotely stored files: first build a list so we can show a warning if remote files are skipped
	remoteFiles := map[string]s3api.S3ObjectReference{}
	for _, dataset := range config.DataSets {
		if dataset.RemotePath == "" || !kitfile.LayerMatchesAnyFilter(dataset, opts.FilterConfs) {
			continue
		}
		// ModelKit references are handled separately (via unpackRemote rooted at dataset.Path).
		if artifact.IsModelKitReference(dataset.RemotePath) {
			continue
		}
		ref, err := s3api.ParseS3ObjectReference(dataset.RemotePath, dataset.RemoteHash)
		if err != nil {
			return fmt.Errorf("failed to parse S3 object reference for dataset %s: %w", dataset.Path, err)
		}
		remoteFiles[dataset.Path] = *ref
		containsRemoteDatasets = true
	}

	if len(remoteFiles) > 0 && opts.IncludeRemote {
		client, err := s3api.SetUpClient(ctx)
		if err != nil {
			return err
		}
		for path, s3Ref := range remoteFiles {
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

	output.Debugf("Unpacked %d model part layers", modelPartIdx)
	output.Debugf("Unpacked %d code layers", codeIdx)
	output.Debugf("Unpacked %d dataset layers", datasetIdx)
	output.Debugf("Unpacked %d docs layers", docsIdx)
	output.Debugf("Unpacked %d prompt layers", promptIdx)

	if containsRemoteDatasets && !opts.IncludeRemote {
		output.Logf(output.LogLevelWarn, "ModelKit contains remote datasets. To unpack, specify the --include-remote flag")
	}

	return nil
}

// unpackSkill handles the --as-skill flow: prompt layers containing SKILL.md
// are installed as agent skills and all other layer types are ignored.
// Unlike unpackRecursive, it does not unpack the Kitfile, traverse parent
// modelkits, or touch the filesystem outside the resolved skills directories.
func unpackSkill(ctx context.Context, opts *UnpackOptions) error {
	ref := opts.ModelRef
	store, err := getStoreForRef(ctx, opts)
	if err != nil {
		ref := artifact.FormatRepositoryForDisplay(opts.ModelRef.String())
		return fmt.Errorf("failed to find reference %s: %s", ref, err)
	}
	manifestDesc, err := store.Resolve(ctx, ref.Reference)
	if err != nil {
		return fmt.Errorf("failed to resolve reference: %w", err)
	}
	manifest, err := util.GetManifest(ctx, store, manifestDesc)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %s", err)
	}
	config, err := util.GetKitfileForManifest(ctx, store, manifest)
	if err != nil {
		if errors.Is(err, util.ErrNoKitfile) {
			output.Infof("No Kitfile found in modelkit; no prompt layers to install as skills")
			return nil
		}
		return err
	}

	var promptsFiltered, skillsFound, skillErrorCount, promptIdx int

	for _, layerDesc := range manifest.Layers {
		mediaType, err := mediatype.ParseMediaType(layerDesc.MediaType)
		if err != nil {
			output.Logf(output.LogLevelWarn, "Unknown media type %s: skipping", layerDesc.MediaType)
			continue
		}
		if mediaType.Base() != mediatype.CodeBaseType {
			continue
		}
		if layerDesc.Annotations[constants.LayerSubtypeAnnotation] != constants.LayerSubtypePrompt {
			continue
		}

		entry := config.Prompts[promptIdx]
		promptIdx++
		if !kitfile.LayerMatchesAnyFilter(entry, opts.FilterConfs) {
			continue
		}

		promptsFiltered++
		result, sErr := installPromptAsSkill(ctx, store, layerDesc, entry, mediaType.Compression(), opts)
		if sErr != nil {
			output.Logf(output.LogLevelWarn, "Error reading prompt %q: %s", entry.Path, sErr)
			skillErrorCount++
			continue
		}
		if result == nil {
			continue
		}
		skillsFound++
		for _, ar := range result.Agents {
			if ar.Err != nil {
				output.Infof("Failed to install skill '%s' for %s: %s", result.SkillName, ar.Agent, ar.Err)
				skillErrorCount++
			} else if ar.Skipped {
				output.Infof("Skipped skill '%s' for %s: already exists (use -o to overwrite)", result.SkillName, ar.Agent)
			} else {
				output.Infof("Installed skill '%s' for %s → %s", result.SkillName, ar.Agent, ar.Path)
			}
		}
	}

	if promptsFiltered == 0 {
		output.Infof("No prompt layers matched the specified filters")
		return nil
	}
	if skillsFound == 0 && skillErrorCount == 0 {
		output.Infof("No agent skills found in modelkit. Prompt layers must contain a SKILL.md file to be installed as skills.")
		return nil
	}
	if skillErrorCount > 0 {
		return fmt.Errorf("failed to install %d skill(s)", skillErrorCount)
	}
	return nil
}

// unpackRemote recursively unpacks the contents of a referenced ModelKit, restricted to a
// single base type (e.g. "model" or "datasets"). If basePath is non-empty, the remote ModelKit
// is unpacked into that subdirectory of the current unpack dir so the referenced ModelKit's
// layer paths land beneath it
func unpackRemote(ctx context.Context, ref *registry.Reference, basePath string, baseType kitfile.BaseType, optsIn *UnpackOptions, visitedRefs []string) error {
	if idx := getIndex(visitedRefs, ref.String()); idx != -1 {
		cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(visitedRefs[idx:], "=>"), ref)
		return fmt.Errorf("found cycle in modelkit references: %s", cycleStr)
	}

	opts := *optsIn
	opts.ModelRef = ref
	// Restrict unpack to the requested base type from the referenced ModelKit.
	if len(opts.FilterConfs) == 0 {
		filter := kitfile.FilterConf{
			BaseTypes: []kitfile.BaseType{baseType},
		}
		opts.FilterConfs = []kitfile.FilterConf{filter}
	} else {
		var filterConfs []kitfile.FilterConf
		for _, conf := range opts.FilterConfs {
			if conf.MatchesBaseType(baseType) {
				// Drop any other base types from this filter
				conf.BaseTypes = []kitfile.BaseType{baseType}
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

func unpackLayer(ctx context.Context, store content.Storage, desc ocispec.Descriptor, unpackPath string, overwrite, ignoreExisting bool, compression mediatype.CompressionType) error {
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
	case mediatype.GzipCompression, mediatype.GzipFastestCompression:
		cr, cErr = gzip.NewReader(rc)
	case mediatype.NoneCompression:
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

// installPromptAsSkill fetches a prompt layer, reads it via ReadSkillLayer,
// and installs it as a skill if it contains a SKILL.md.
// Returns nil result (and nil error) if the layer is not a skill.
// Returns nil result with an error if ReadSkillLayer fails.
// Returns a result (and nil error) if installation was attempted.
func installPromptAsSkill(ctx context.Context, store content.Storage, desc ocispec.Descriptor, entry artifact.Prompt, compression mediatype.CompressionType, opts *UnpackOptions) (*skill.InstallResult, error) {
	// Fast reject: the OCI descriptor carries the compressed layer size.
	// If it already exceeds the limit, skip the download entirely.
	if desc.Size > skill.MaxSkillLayerSize {
		return nil, fmt.Errorf("prompt layer %q compressed size (%d bytes) exceeds maximum (%d bytes)", entry.Path, desc.Size, skill.MaxSkillLayerSize)
	}

	rc, err := store.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching layer: %w", err)
	}
	defer func() { _ = rc.Close() }()

	var cr io.ReadCloser
	switch compression {
	case mediatype.GzipCompression, mediatype.GzipFastestCompression:
		cr, err = gzip.NewReader(rc)
		if err != nil {
			return nil, fmt.Errorf("decompressing layer: %w", err)
		}
	case mediatype.NoneCompression:
		cr = rc
	default:
		return nil, fmt.Errorf("unsupported compression type %q for prompt layer %q", compression, entry.Path)
	}
	defer func() { _ = cr.Close() }()

	tr := tar.NewReader(cr)
	entries, isSkill, frontmatterName, err := skill.ReadSkillLayer(tr)
	if err != nil {
		return nil, err
	}

	if !isSkill {
		output.Infof("Skipping prompt %q: not a SKILL.md, cannot install as skill", entry.Path)
		return nil, nil
	}

	skillName := skill.DeriveSkillName(frontmatterName, entry, opts.ModelRef)
	result := skill.InstallSkill(entries, skillName, entry, opts.SkillOptions)
	return &result, nil
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
