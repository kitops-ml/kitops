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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
)

// RepoPath returns the path that should be used for creating a local OCI index given a
// specific *registry.Reference.
func RepoPath(storagePath string, ref *registry.Reference) string {
	return filepath.Join(storagePath, ref.Registry, ref.Repository)
}

// GetManifestAndKitfile returns the manifest and config (Kitfile) for a manifest Descriptor.
// Calls GetManifest and GetKitfileForManifest. If the manifest is retrieved but no Kitfile
// can be found, returns the manifest and error equal to ErrNoKitfile
func GetManifestAndKitfile(ctx context.Context, store oras.ReadOnlyTarget, manifestDesc ocispec.Descriptor) (*ocispec.Manifest, *artifact.KitFile, error) {
	manifest, err := GetManifest(ctx, store, manifestDesc)
	if err != nil {
		return nil, nil, err
	}
	config, err := GetKitfileForManifest(ctx, store, manifest)
	if err != nil {
		return manifest, nil, err
	}
	return manifest, config, nil
}

// GetManifest returns the Manifest described by a Descriptor. Returns an error if the manifest blob cannot be
// resolved or does not represent a modelkit manifest.
func GetManifest(ctx context.Context, store oras.ReadOnlyTarget, manifestDesc ocispec.Descriptor) (*ocispec.Manifest, error) {
	manifestBytes, err := content.FetchAll(ctx, store, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest %s: %w", manifestDesc.Digest, err)
	}
	manifest := &ocispec.Manifest{}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest %s: %w", manifestDesc.Digest, err)
	}
	if _, err := mediatype.ModelFormatForManifest(manifest); err != nil {
		return nil, ErrNotAModelKit
	}

	return manifest, nil
}

// GetKitfileForManifest returns the Kitfile for a given manifest, either by retrieving it from an
// OCI store or by reading it from manifest annotations. If manifest type is unrecognized, returns
// ErrNotAModelKit. If the manifest is recognized but does not contain a Kitfile (e.g. it was not
// created by Kit), returns ErrNoKitfile.
func GetKitfileForManifest(ctx context.Context, store oras.ReadOnlyTarget, manifest *ocispec.Manifest) (*artifact.KitFile, error) {
	modelFormat, err := mediatype.ModelFormatForManifest(manifest)
	if err != nil {
		return nil, ErrNotAModelKit
	}
	switch modelFormat {
	case mediatype.KitFormat:
		return GetConfig(ctx, store, manifest.Config)
	case mediatype.ModelPackFormat:
		// TODO: can we (try to) generate a Kitfile from a ModelPack manifest?
		if manifest.Annotations == nil || manifest.Annotations[constants.KitfileJsonAnnotation] == "" {
			return nil, ErrNoKitfile
		}
		kfstring := manifest.Annotations[constants.KitfileJsonAnnotation]
		kitfile := &artifact.KitFile{}
		if err := json.Unmarshal([]byte(kfstring), kitfile); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
		return kitfile, nil
	default:
		// Won't happen but necessary for completeness
		return nil, fmt.Errorf("unknown artifact type")
	}
}

// GetConfig returns the config (Kitfile) described by a descriptor. Returns an error if the config blob cannot
// be resolved or if the descriptor does not describe a Kitfile.
func GetConfig(ctx context.Context, store oras.ReadOnlyTarget, configDesc ocispec.Descriptor) (*artifact.KitFile, error) {
	if configDesc.MediaType == "" || configDesc.MediaType == ocispec.MediaTypeEmptyJSON {
		return nil, fmt.Errorf("manifest does not have a config section")
	}
	if configDesc.MediaType != mediatype.KitConfigMediaType.String() {
		return nil, fmt.Errorf("configuration descriptor does not describe a Kitfile")
	}
	configBytes, err := content.FetchAll(ctx, store, configDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	config := &artifact.KitFile{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return config, nil
}

// ResolveManifest returns the manifest for a reference (tag), if present in the target store
func ResolveManifest(ctx context.Context, store oras.Target, reference string) (ocispec.Descriptor, *ocispec.Manifest, error) {
	desc, err := store.Resolve(ctx, reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("reference %s not found in repository: %w", reference, err)
	}
	manifest, err := GetManifest(ctx, store, desc)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}
	return desc, manifest, nil
}

// ResolveManifestAndConfig returns the manifest and config (Kitfile) for a given reference (tag), if present
// in the store. Calls GetManifest and GetKitfileForManifest. If the manifest is retrieved but no Kitfile
// can be found, returns the manifest and error equal to ErrNoKitfile
func ResolveManifestAndConfig(ctx context.Context, store oras.Target, reference string) (ocispec.Descriptor, *ocispec.Manifest, *artifact.KitFile, error) {
	desc, manifest, err := ResolveManifest(ctx, store, reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, nil, err
	}
	config, err := GetKitfileForManifest(ctx, store, manifest)
	if err != nil {
		return desc, manifest, nil, err
	}
	return desc, manifest, config, nil
}
