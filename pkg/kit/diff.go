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

package kit

import (
	"context"
	"fmt"
	"sync"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

type DiffOptions struct {
	options.NetworkOptions
	ConfigHome string
	RefA       *registry.Reference
	RefB       *registry.Reference
	LocalityA  string // "local", "remote", or "" (auto: try local then remote)
	LocalityB  string
}

type DiffResult struct {
	Identical        bool
	SameConfig       bool
	AnnotationsMatch bool
	SharedLayers     []ocispec.Descriptor
	UniqueLayersA    []ocispec.Descriptor
	UniqueLayersB    []ocispec.Descriptor
}

func Diff(ctx context.Context, opts *DiffOptions) (*DiffResult, error) {
	if opts.NetworkOptions.CredentialsPath == "" {
		opts.NetworkOptions.CredentialsPath = constants.CredentialsPath(opts.ConfigHome)
		if !opts.NetworkOptions.PlainHTTP && !opts.NetworkOptions.TLSVerify {
			opts.NetworkOptions.TLSVerify = true
		}
	}

	type manifestResult struct {
		manifest *ocispec.Manifest
		desc     ocispec.Descriptor
		err      error
	}

	var wg sync.WaitGroup
	wg.Add(2)

	resultA := &manifestResult{}
	resultB := &manifestResult{}

	go func() {
		defer wg.Done()
		resultA.manifest, resultA.desc, resultA.err = resolveManifest(ctx, opts.RefA, opts.LocalityA, opts)
	}()
	go func() {
		defer wg.Done()
		resultB.manifest, resultB.desc, resultB.err = resolveManifest(ctx, opts.RefB, opts.LocalityB, opts)
	}()
	wg.Wait()

	if resultA.err != nil {
		return nil, fmt.Errorf("failed to get manifest for ref A: %w", resultA.err)
	}
	if resultB.err != nil {
		return nil, fmt.Errorf("failed to get manifest for ref B: %w", resultB.err)
	}

	if resultA.desc.Digest == resultB.desc.Digest {
		return &DiffResult{Identical: true}, nil
	}

	return CompareManifests(resultA.manifest, resultB.manifest), nil
}

// CompareManifests is a pure function that compares two OCI manifests.
func CompareManifests(a, b *ocispec.Manifest) *DiffResult {
	result := &DiffResult{}
	result.SameConfig = a.Config.Digest == b.Config.Digest

	numAnnotations := len(a.Annotations)
	if numAnnotations != len(b.Annotations) {
		result.AnnotationsMatch = false
	} else {
		result.AnnotationsMatch = true
		for k, v := range a.Annotations {
			if v2, ok := b.Annotations[k]; !ok || v2 != v {
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
	for _, layer := range a.Layers {
		layerMapA[layer.Digest.String()] = layer
	}
	for _, layer := range b.Layers {
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

	return result
}

func resolveManifest(ctx context.Context, ref *registry.Reference, locality string, opts *DiffOptions) (*ocispec.Manifest, ocispec.Descriptor, error) {
	switch locality {
	case "remote":
		return resolveManifestRemote(ctx, ref, opts)
	case "local":
		return resolveManifestLocal(ctx, ref, opts)
	default:
		manifest, desc, err := resolveManifestLocal(ctx, ref, opts)
		if err != nil {
			return resolveManifestRemote(ctx, ref, opts)
		}
		return manifest, desc, nil
	}
}

func resolveManifestRemote(ctx context.Context, ref *registry.Reference, opts *DiffOptions) (*ocispec.Manifest, ocispec.Descriptor, error) {
	repository, err := remote.NewRepository(ctx, ref.Registry, ref.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, ocispec.DescriptorEmptyJSON, err
	}
	desc, manifest, err := util.ResolveManifest(ctx, repository, ref.Reference)
	if err != nil {
		return nil, ocispec.DescriptorEmptyJSON, err
	}
	return manifest, desc, nil
}

func resolveManifestLocal(ctx context.Context, ref *registry.Reference, opts *DiffOptions) (*ocispec.Manifest, ocispec.Descriptor, error) {
	storageRoot := constants.StoragePath(opts.ConfigHome)
	localRepo, err := local.NewLocalRepo(storageRoot, ref)
	if err != nil {
		return nil, ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to read local storage: %w", err)
	}
	desc, manifest, err := util.ResolveManifest(ctx, localRepo, ref.Reference)
	if err != nil {
		return nil, ocispec.DescriptorEmptyJSON, err
	}
	return manifest, desc, nil
}
