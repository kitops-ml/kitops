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
	"context"
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
)

type unpackStep struct {
	desc        ocispec.Descriptor
	mediatype   mediatype.MediaType
	userMessage string
}

func generateUnpackPlan(manifest *ocispec.Manifest, kitfile *artifact.KitFile, filters []kfutils.FilterConf) ([]unpackStep, error) {
	var steps []unpackStep

	messageFor := func(layerType, name, path string) string {
		if name != "" {
			return fmt.Sprintf("Unpacking %s %s to %s", layerType, name, path)
		}
		return fmt.Sprintf("Unpacking %s to %s", layerType, path)
	}

	descToDigest := make(map[string]ocispec.Descriptor, len(manifest.Layers))
	for _, layerDesc := range manifest.Layers {
		descToDigest[layerDesc.Digest.String()] = layerDesc
	}

	addStep := func(digest, layerType, name, path string) error {
		desc, ok := descToDigest[digest]
		if !ok {
			return fmt.Errorf("digest %s not found in manifest", digest)
		}
		mt, err := mediatype.ParseMediaType(desc.MediaType)
		if err != nil {
			return err
		}
		steps = append(steps, unpackStep{
			desc:        desc,
			mediatype:   mt,
			userMessage: messageFor(layerType, name, path),
		})
		return nil
	}

	if kitfile.Model != nil {
		if !artifact.IsModelKitReference(kitfile.Model.Path) && kfutils.LayerMatchesAnyFilter(kitfile.Model, filters) {
			if err := addStep(kitfile.Model.Digest, "model", kitfile.Model.Name, kitfile.Model.Path); err != nil {
				return nil, err
			}
		}
		for _, modelpart := range kitfile.Model.Parts {
			if kfutils.LayerMatchesAnyFilter(modelpart, filters) {
				if err := addStep(modelpart.Digest, "modelpart", modelpart.Name, modelpart.Path); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, dataset := range kitfile.DataSets {
		if dataset.RemotePath != "" {
			continue
		}
		if kfutils.LayerMatchesAnyFilter(dataset, filters) {
			if err := addStep(dataset.Digest, "dataset", dataset.Name, dataset.Path); err != nil {
				return nil, err
			}
		}
	}
	for _, code := range kitfile.Code {
		if kfutils.LayerMatchesAnyFilter(code, filters) {
			if err := addStep(code.Digest, "code", "", code.Path); err != nil {
				return nil, err
			}
		}
	}
	for _, doc := range kitfile.Docs {
		if kfutils.LayerMatchesAnyFilter(doc, filters) {
			if err := addStep(doc.Digest, "docs", "", doc.Path); err != nil {
				return nil, err
			}
		}
	}
	for _, prompt := range kitfile.Prompts {
		if kfutils.LayerMatchesAnyFilter(prompt, filters) {
			if err := addStep(prompt.Digest, "prompt", prompt.Name, prompt.Path); err != nil {
				return nil, err
			}
		}
	}
	for _, server := range kitfile.MCPServers {
		if kfutils.LayerMatchesAnyFilter(server, filters) {
			if err := addStep(server.Digest, "MCP server", server.Name, server.Path); err != nil {
				return nil, err
			}
		}
	}

	return steps, nil
}

// getStoreForRef returns the appropriate store (local or remote) for a ModelKit reference.
func getStoreForRef(ctx context.Context, opts *UnpackOptions) (oras.Target, error) {
	storageHome := constants.StoragePath(opts.ConfigHome)
	localRepo, err := local.NewLocalRepo(storageHome, opts.ModelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %s\n", err)
	}

	if _, err := localRepo.Resolve(ctx, opts.ModelRef.Reference); err == nil {
		// Reference is present in local storage
		return localRepo, nil
	}

	if opts.ModelRef.Registry == artifact.DefaultRegistry {
		return nil, fmt.Errorf("not found")
	}
	// Not in local storage, check remote
	repo, err := remote.NewRepository(ctx, opts.ModelRef.Registry, opts.ModelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, fmt.Errorf("could not resolve repository %s in registry %s", opts.ModelRef.Repository, opts.ModelRef.Registry)
	}
	if _, err := repo.Resolve(ctx, opts.ModelRef.Reference); err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, fmt.Errorf("reference %s is not present in local storage and could not be found in remote", opts.ModelRef.String())
		}
		return nil, fmt.Errorf("unexpected error retrieving reference from remote: %w", err)
	}

	return repo, nil
}

func getIndex(list []string, s string) int {
	for idx, item := range list {
		if s == item {
			return idx
		}
	}
	return -1
}
