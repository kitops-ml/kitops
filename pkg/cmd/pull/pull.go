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

package pull

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func runPull(ctx context.Context, opts *pullOptions) (ocispec.Descriptor, error) {
	storageHome := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageHome, opts.modelRef)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	return runPullRecursive(ctx, localRepo, opts, []string{})
}

func runPullRecursive(ctx context.Context, localRepo local.LocalRepo, opts *pullOptions, pulledRefs []string) (ocispec.Descriptor, error) {
	refStr := artifact.FormatRepositoryForDisplay(opts.modelRef.String())
	if idx := getIndex(pulledRefs, refStr); idx != -1 {
		cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(pulledRefs[idx:], "=>"), refStr)
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("found cycle in modelkit references: %s", cycleStr)
	}
	pulledRefs = append(pulledRefs, refStr)
	if len(pulledRefs) > constants.MaxModelRefChain {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("reached maximum number of model references: [%s]", strings.Join(pulledRefs, "=>"))
	}

	desc, err := pullModel(ctx, localRepo, opts)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}

	if err := pullParents(ctx, localRepo, desc, opts, pulledRefs); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to pull referenced modelkits: %w", err)
	}

	return desc, nil
}

func pullParents(ctx context.Context, localRepo local.LocalRepo, desc ocispec.Descriptor, optsIn *pullOptions, pulledRefs []string) error {
	_, config, err := util.GetManifestAndKitfile(ctx, localRepo, desc)
	if err != nil {
		if errors.Is(err, util.ErrNoKitfile) {
			// If there's no Kitfile but it's otherwise a support artifact type, skip pulling parents as there aren't any
			return nil
		}
		return err
	}

	var parentRefs []string
	if config.Model != nil && artifact.IsModelKitReference(config.Model.Path) {
		parentRefs = append(parentRefs, config.Model.Path)
	}
	for _, dataset := range config.DataSets {
		if artifact.IsModelKitReference(dataset.RemotePath) {
			parentRefs = append(parentRefs, dataset.RemotePath)
		}
	}

	for _, refStr := range parentRefs {
		parentRef, _, err := artifact.ParseReference(refStr)
		if err != nil {
			return err
		}
		output.Infof("Pulling referenced image %s", refStr)
		opts := *optsIn
		opts.modelRef = parentRef
		// Labels select from the index named on the command line; referenced ModelKits are
		// named by the Kitfile and are pulled as they are.
		opts.labels = nil
		if _, err := runPullRecursive(ctx, localRepo, &opts, pulledRefs); err != nil {
			return err
		}
	}
	return nil
}

func pullModel(ctx context.Context, localRepo local.LocalRepo, opts *pullOptions) (ocispec.Descriptor, error) {
	repo, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to read repository: %w", err)
	}
	modelRef, entry, err := resolveModelRef(ctx, opts.modelRef, repo, opts.labels)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	if entry == nil && len(opts.labels) > 0 {
		output.Logf(output.LogLevelWarn, "Labels are ignored: %s is a ModelKit, not a ModelKit index",
			artifact.FormatRepositoryForDisplay(opts.modelRef.String()))
	}

	if entry != nil {
		named := index.EntryReference(*entry, opts.modelRef)
		if named == "" {
			named = entry.Digest.String()
		}
		output.Infof("Index %s refers to ModelKit %s", artifact.FormatRepositoryForDisplay(opts.modelRef.String()), named)
	}

	desc, err := localRepo.PullModel(ctx, repo, *modelRef, &opts.NetworkOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to pull: %w", err)
	}

	// An entry is pulled by digest, so PullModel does not tag it; like pulling a multi-platform
	// image, the selected ModelKit takes the index's tag in local storage.
	if entry != nil && !artifact.ReferenceIsDigest(opts.modelRef.Reference) {
		if err := localRepo.Tag(ctx, desc, opts.modelRef.Reference); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to tag pulled ModelKit: %w", err)
		}
	}

	return desc, nil
}

// resolveModelRef returns the reference that should be pulled. A ModelKit index resolves to
// the entry its labels select, so that pulling always produces a ModelKit; anything else must
// already be a ModelKit and is returned unchanged.
func resolveModelRef(ctx context.Context, ref *registry.Reference, repo registry.Repository, labels map[string]string) (*registry.Reference, *index.ModelKitIndexDescriptor, error) {
	desc, rc, err := repo.FetchReference(ctx, ref.Reference)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", ref.String(), err)
	}
	defer rc.Close()

	manifestBytes, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	if desc.MediaType == ocispec.MediaTypeImageIndex {
		// Registries do not include artifactType in the descriptors they return, so the index
		// itself is the only thing that can say whether this is a ModelKit index.
		idx, err := index.ParseIndex(manifestBytes)
		if errors.Is(err, index.ErrNotAKitIndex) {
			return nil, nil, fmt.Errorf("reference %s is not an image manifest", ref.String())
		} else if err != nil {
			return nil, nil, err
		}
		if len(labels) == 0 {
			return nil, nil, fmt.Errorf("reference %s is a ModelKit index; select the ModelKit to pull with --label (-l). 'kit index info --remote' lists its ModelKits and their labels",
				artifact.FormatRepositoryForDisplay(ref.String()))
		}
		entry, err := index.SelectEntry(idx, labels, ref)
		if err != nil {
			return nil, nil, err
		}
		entryRef := *ref
		entryRef.Reference = entry.Digest.String()
		return &entryRef, &entry, nil
	}
	if desc.MediaType != ocispec.MediaTypeImageManifest {
		return nil, nil, fmt.Errorf("reference %s is not an image manifest", ref.String())
	}
	manifest := &ocispec.Manifest{}
	if err := json.Unmarshal(manifestBytes, manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	if _, err := mediatype.ModelFormatForManifest(manifest); err != nil {
		return nil, nil, fmt.Errorf("reference %s does not refer to a model", ref.String())
	}
	return ref, nil, nil
}

func getIndex(list []string, s string) int {
	for idx, item := range list {
		if s == item {
			return idx
		}
	}
	return -1
}
