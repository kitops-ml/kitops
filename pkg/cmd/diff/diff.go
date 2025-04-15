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

package diff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
)

// Helper struct diffInfo holds the manifest and its descriptor for a ModelKit.
type diffInfo struct {
	Manifest   *ocispec.Manifest
	Descriptor ocispec.Descriptor
}

// Helper struct DiffResult contains the comparison results between two ModelKits.
type DiffResult struct {
	SameConfig       bool
	AnnotationsMatch bool
	SharedLayers     []ocispec.Descriptor
	UniqueLayersA    []ocispec.Descriptor
	UniqueLayersB    []ocispec.Descriptor
}

// compareManifests compares two OCI manifests and returns the shared and unique layers.
func CompareManifests(manifestA *ocispec.Manifest, manifestB *ocispec.Manifest) *DiffResult {
	result := &DiffResult{}

	// Compare the config digests
	result.SameConfig = manifestA.Config.Digest == manifestB.Config.Digest

	// Compare the annotations
	numAnnotations := len(manifestA.Annotations)
	if numAnnotations != len(manifestB.Annotations) {
		result.AnnotationsMatch = false
	} else {
		result.AnnotationsMatch = true
		for k, v := range manifestA.Annotations {
			if v2, ok := manifestB.Annotations[k]; !ok || v2 != v {
				result.AnnotationsMatch = false
				break
			}
			numAnnotations--
		}
		if numAnnotations != 0 {
			result.AnnotationsMatch = false
		}
	}

	layerMapA := make(map[string]ocispec.Descriptor)
	for _, layer := range manifestA.Layers {
		layerMapA[layer.Digest.String()] = layer
	}

	for _, layer := range manifestB.Layers {
		if _, ok := layerMapA[layer.Digest.String()]; ok {
			result.SharedLayers = append(result.SharedLayers, layer)
			delete(layerMapA, layer.Digest.String())
		} else {
			result.UniqueLayersB = append(result.UniqueLayersB, layer)
		}
	}

	result.UniqueLayersA = make([]ocispec.Descriptor, 0, len(layerMapA))
	for _, layer := range layerMapA {
		result.UniqueLayersA = append(result.UniqueLayersA, layer)
	}

	// Sort the slices by layer type
	sort.Slice(result.SharedLayers, func(i, j int) bool {
		return result.SharedLayers[i].MediaType < result.SharedLayers[j].MediaType
	})
	sort.Slice(result.UniqueLayersA, func(i, j int) bool {
		return result.UniqueLayersA[i].MediaType < result.UniqueLayersA[j].MediaType
	})
	sort.Slice(result.UniqueLayersB, func(i, j int) bool {
		return result.UniqueLayersB[i].MediaType < result.UniqueLayersB[j].MediaType
	})

	return result
}

func getManifest(ctx context.Context, arg string, ref *registry.Reference, opts *diffOptions) (*diffInfo, error) {
	if strings.HasPrefix(arg, remotePrefix) {
		return getManifestFromRemote(ctx, ref, opts)
	} else if strings.HasPrefix(arg, localPrefix) {
		return getManifestFromLocal(ctx, ref, opts)
	} else {
		manifest, err := getManifestFromLocal(ctx, ref, opts)
		if err != nil {
			return getManifestFromRemote(ctx, ref, opts)
		}
		return manifest, nil
	}
}

func getManifestFromRemote(ctx context.Context, ref *registry.Reference, opts *diffOptions) (*diffInfo, error) {
	repository, err := remote.NewRepository(ctx, ref.Registry, ref.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, err
	}
	desc, manifest, err := util.ResolveManifest(ctx, repository, ref.Reference)
	if err != nil {
		return nil, err
	}
	return &diffInfo{
		Manifest:   manifest,
		Descriptor: desc,
	}, nil
}

func getManifestFromLocal(ctx context.Context, ref *registry.Reference, opts *diffOptions) (*diffInfo, error) {
	storageRoot := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageRoot, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}
	desc, manifest, err := util.ResolveManifest(ctx, localRepo, ref.Reference)
	if err != nil {
		return nil, err
	}
	return &diffInfo{
		Manifest:   manifest,
		Descriptor: desc,
	}, nil
}
