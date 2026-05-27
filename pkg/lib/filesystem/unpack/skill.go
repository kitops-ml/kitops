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

package unpack

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/lib/skill"
	"github.com/kitops-ml/kitops/pkg/output"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

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
