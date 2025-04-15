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

package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"oras.land/oras-go/v2/registry"
)

func listRemoteKits(ctx context.Context, opts *listOptions) ([]string, error) {
	remoteRegistry, err := remote.NewRegistry(opts.remoteRef.Registry, &opts.NetworkOptions)
	if err != nil {
		return nil, fmt.Errorf("could not resolve registry %s: %w", opts.remoteRef.Registry, err)
	}

	repo, err := remoteRegistry.Repository(ctx, opts.remoteRef.Repository)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository: %w", err)
	}
	if opts.remoteRef.Reference != "" {
		return listImageTag(ctx, repo, opts.remoteRef)
	}
	return listTags(ctx, repo, opts.remoteRef)
}

func listTags(ctx context.Context, repo registry.Repository, ref *registry.Reference) ([]string, error) {
	var tags []string
	err := repo.Tags(ctx, "", func(tagsPage []string) error {
		tags = append(tags, tagsPage...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags on repostory: %w", err)
	}

	var allLines []string
	for _, tag := range tags {
		tagRef := &registry.Reference{
			Registry:   ref.Registry,
			Repository: ref.Repository,
			Reference:  tag,
		}
		infoLines, err := listImageTag(ctx, repo, tagRef)
		if err != nil && !errors.Is(err, util.ErrNotAModelKit) {
			return nil, err
		}
		allLines = append(allLines, infoLines...)
	}

	return allLines, nil
}

func listImageTag(ctx context.Context, repo registry.Repository, ref *registry.Reference) ([]string, error) {
	manifestDesc, err := repo.Resolve(ctx, ref.Reference)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", ref.Reference, err)
	}
	manifest, config, err := util.GetManifestAndConfig(ctx, repo, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to read modelkit: %w", err)
	}
	if manifest.Config.MediaType != constants.ModelConfigMediaType.String() {
		return nil, nil
	}
	info := modelInfo{
		repo:   ref.Repository,
		digest: string(manifestDesc.Digest),
		tags:   []string{ref.Reference},
	}
	info.fill(manifest, config)

	return info.format(), nil
}
