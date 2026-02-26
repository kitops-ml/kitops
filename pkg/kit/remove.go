// Copyright 2025 The KitOps Authors.
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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

type RemoveOptions struct {
	options.NetworkOptions
	ConfigHome  string
	ForceDelete bool
	RemoveAll   bool
	Remote      bool
	ModelRef    *registry.Reference
	ExtraTags   []string
}

func Remove(ctx context.Context, opts *RemoveOptions) error {
	switch {
	case opts.ModelRef != nil:
		if opts.Remote {
			return removeRemoteModel(ctx, opts)
		}
		return removeModel(ctx, opts)
	case opts.RemoveAll && !opts.ForceDelete:
		return removeUntaggedModels(ctx, opts)
	case opts.RemoveAll && opts.ForceDelete:
		return removeAllModels(ctx, opts)
	default:
		return nil
	}
}

func removeAllModels(ctx context.Context, opts *RemoveOptions) error {
	localRepos, err := local.GetAllLocalRepos(constants.StoragePath(opts.ConfigHome))
	if err != nil {
		return fmt.Errorf("failed to read local storage: %w", err)
	}
	for _, localRepo := range localRepos {
		repository := util.FormatRepositoryForDisplay(localRepo.GetRepoName())

		models := localRepo.GetAllModels()
		skipManifests := map[digest.Digest]bool{}
		for _, manifestDesc := range models {
			if skipManifests[manifestDesc.Digest] {
				continue
			}
			tags := localRepo.GetTags(manifestDesc)
			for _, tag := range tags {
				if err := localRepo.Untag(ctx, tag); err != nil {
					output.Errorf("Failed to untag %s:%s: %w", repository, tag, err)
				}
				output.Infof("Untagged %s:%s", repository, tag)
			}

			if err := localRepo.Delete(ctx, manifestDesc); err != nil {
				output.Errorf("Failed to remove %s@%s: %s", repository, manifestDesc.Digest, err)
				continue
			}
			skipManifests[manifestDesc.Digest] = true
			output.Infof("Removed %s@%s", repository, manifestDesc.Digest)
		}
	}
	return nil
}

func removeUntaggedModels(ctx context.Context, opts *RemoveOptions) error {
	localRepos, err := local.GetAllLocalRepos(constants.StoragePath(opts.ConfigHome))
	if err != nil {
		return fmt.Errorf("failed to read local storage: %w", err)
	}
	for _, localRepo := range localRepos {
		manifests := localRepo.GetAllModels()
		repo := util.FormatRepositoryForDisplay(localRepo.GetRepoName())
		for _, manifestDesc := range manifests {
			tags := localRepo.GetTags(manifestDesc)
			if len(tags) > 0 {
				output.Debugf("Skipping %s (tags: %s)", manifestDesc.Digest, strings.Join(tags, ", "))
				continue
			}
			if err := localRepo.Delete(ctx, manifestDesc); err != nil {
				output.Errorf("Failed to remove %s@%s: %s", repo, manifestDesc.Digest, err)
				continue
			}
			output.Infof("Removed %s@%s", repo, manifestDesc.Digest)
		}
	}
	return nil
}

func removeModel(ctx context.Context, opts *RemoveOptions) error {
	storageRoot := constants.StoragePath(opts.ConfigHome)
	localRepo, err := local.NewLocalRepo(storageRoot, opts.ModelRef)
	if err != nil {
		return fmt.Errorf("failed to read local storage at path %s: %w", storageRoot, err)
	}
	desc, err := removeModelRef(ctx, localRepo, opts.ModelRef, opts.ForceDelete)
	if err != nil {
		return fmt.Errorf("failed to remove: %s", err)
	}
	displayRef := util.FormatRepositoryForDisplay(opts.ModelRef.String())
	output.Infof("Removed %s (digest %s)", displayRef, desc.Digest)

	for _, tag := range opts.ExtraTags {
		ref := *opts.ModelRef
		ref.Reference = tag
		displayRef := util.FormatRepositoryForDisplay(ref.String())
		desc, err := removeModelRef(ctx, localRepo, &ref, opts.ForceDelete)
		if err != nil {
			output.Errorf("Failed to remove tag %s: %s", tag, err)
		} else {
			output.Infof("Removed %s (digest %s)", displayRef, desc.Digest)
		}
	}
	return nil
}

func removeModelRef(ctx context.Context, localRepo local.LocalRepo, ref *registry.Reference, forceDelete bool) (ocispec.Descriptor, error) {
	desc, err := oras.Resolve(ctx, localRepo, ref.Reference, oras.ResolveOptions{})
	if err != nil {
		if err == errdef.ErrNotFound {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("model %s not found", util.FormatRepositoryForDisplay(ref.String()))
		}
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("error resolving model: %s", err)
	}

	if err := ref.ValidateReferenceAsDigest(); err == nil || forceDelete {
		output.Debugf("Deleting manifest with digest %s", ref.Reference)
		if err := localRepo.Delete(ctx, desc); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to delete model: %ws", err)
		}
		return desc, nil
	}

	tags := localRepo.GetTags(desc)
	if len(tags) <= 1 {
		output.Debugf("Deleting manifest tagged %s", ref.Reference)
		if err := localRepo.Delete(ctx, desc); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to delete model: %w", err)
		}
	} else {
		output.Debugf("Found other tags for manifest: [%s]", strings.Join(tags, ", "))
		output.Debugf("Untagging %s", ref.Reference)
		if err := localRepo.Untag(ctx, ref.Reference); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to untag model: %w", err)
		}
	}

	return desc, nil
}

func removeRemoteModel(ctx context.Context, opts *RemoveOptions) error {
	repository, err := remote.NewRepository(ctx, opts.ModelRef.Registry, opts.ModelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return err
	}

	desc, err := repository.Resolve(ctx, opts.ModelRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return fmt.Errorf("model %s not found", util.FormatRepositoryForDisplay(opts.ModelRef.String()))
		}
		return fmt.Errorf("error resolving modelkit: %w", err)
	}

	if !util.ReferenceIsDigest(opts.ModelRef.Reference) && !opts.ForceDelete {
		output.Infof("Untagging remote ModelKit %s", util.FormatRepositoryForDisplay(opts.ModelRef.String()))
		return untagRemoteModel(ctx, opts.ModelRef.Reference, repository)
	}

	deleteRef := *opts.ModelRef
	deleteRef.Reference = desc.Digest.String()
	output.Infof("Deleting remote ModelKit %s", util.FormatRepositoryForDisplay(deleteRef.String()))
	if err := repository.Delete(ctx, desc); err != nil {
		if errResp, ok := err.(*errcode.ErrorResponse); ok && errResp.StatusCode == http.StatusMethodNotAllowed {
			return fmt.Errorf("removing models is unsupported by registry %s", opts.ModelRef.Registry)
		}
		return fmt.Errorf("failed to remove remote model: %w", err)
	}
	return nil
}

func untagRemoteModel(ctx context.Context, tag string, repo registry.Repository) error {
	untaggerRepo, ok := repo.(content.Untagger)
	if !ok {
		return fmt.Errorf("remote repository implementation does not support untagging ModelKits")
	}
	if err := untaggerRepo.Untag(ctx, tag); err != nil {
		return fmt.Errorf("failed to untag ModelKit in remote: %w", err)
	}
	return nil
}

//
