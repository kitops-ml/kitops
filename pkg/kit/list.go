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
	"errors"
	"fmt"
	"sort"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

type ListOptions struct {
	options.NetworkOptions
	ConfigHome  string
	RemoteRef   *registry.Reference
	FilterConfs []kitfile.FilterConf
}

type ModelInfo struct {
	Repo      string   `json:"repo"`
	Digest    string   `json:"digest"`
	Tags      []string `json:"tags"`
	ModelName string   `json:"modelName"`
	Size      string   `json:"size"`
	Author    string   `json:"author"`
}

func List(ctx context.Context, opts *ListOptions) ([]ModelInfo, error) {
	if opts.RemoteRef == nil {
		return listLocalKits(ctx, opts)
	}
	return listRemoteKits(ctx, opts)
}

func listLocalKits(ctx context.Context, opts *ListOptions) ([]ModelInfo, error) {
	storageRoot := constants.StoragePath(opts.ConfigHome)

	localRepos, err := local.GetAllLocalRepos(storageRoot)
	if err != nil {
		return nil, err
	}
	var allInfo []ModelInfo
	for _, repo := range localRepos {
		infos, err := readInfoFromRepo(ctx, repo, opts.FilterConfs)
		if err != nil {
			return nil, err
		}
		allInfo = append(allInfo, infos...)
	}

	return allInfo, nil
}

func readInfoFromRepo(ctx context.Context, repo local.LocalRepo, filterConfs []kitfile.FilterConf) ([]ModelInfo, error) {
	var infos []ModelInfo
	manifestDescs := repo.GetAllModels()
	for _, manifestDesc := range manifestDescs {
		manifest, config, err := util.GetManifestAndKitfile(ctx, repo, manifestDesc)
		if err != nil {
			if errors.Is(err, util.ErrNotAModelKit) {
				// Shouldn't happen since this is a local repo, but either way it's not a supported artifact
				continue
			}
			// Allow artifacts without Kitfiles as all that will be lacking is some metadata; we can still
			// describe them
			if !errors.Is(err, util.ErrNoKitfile) {
				return nil, err
			}
		}
		if !kitfile.KitfileContainsMatchingLayer(config, filterConfs) {
			continue
		}
		tags := repo.GetTags(manifestDesc)
		// Strip localhost from repo if present, since we added it
		repository := artifact.FormatRepositoryForDisplay(repo.GetRepoName())
		if repository == "" {
			repository = "<none>"
		}
		info := ModelInfo{
			Repo:   repository,
			Digest: string(manifestDesc.Digest),
			Tags:   tags,
		}
		info.fill(manifest, config)

		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return (infos[i].Repo < infos[j].Repo) ||
			((infos[i].Repo == infos[j].Repo) && (infos[i].Digest < infos[j].Digest))
	})
	return infos, nil
}

func listRemoteKits(ctx context.Context, opts *ListOptions) ([]ModelInfo, error) {
	repo, err := remote.NewRepository(ctx, opts.RemoteRef.Registry, opts.RemoteRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository: %w", err)
	}
	if opts.RemoteRef.Reference != "" {
		info, err := listImageTag(ctx, repo, opts.RemoteRef, opts.FilterConfs)
		if info == nil || err != nil {
			return nil, err
		}
		return []ModelInfo{*info}, nil
	}
	return listTags(ctx, repo, opts.RemoteRef, opts.FilterConfs)
}

func listTags(ctx context.Context, repo registry.Repository, ref *registry.Reference, filterConfs []kitfile.FilterConf) ([]ModelInfo, error) {
	var tags []string
	err := repo.Tags(ctx, "", func(tagsPage []string) error {
		tags = append(tags, tagsPage...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags on repository: %w", err)
	}

	var allInfos []ModelInfo
	for _, tag := range tags {
		tagRef := &registry.Reference{
			Registry:   ref.Registry,
			Repository: ref.Repository,
			Reference:  tag,
		}
		info, err := listImageTag(ctx, repo, tagRef, filterConfs)
		if err != nil && !errors.Is(err, util.ErrNotAModelKit) {
			return nil, err
		}
		if info != nil {
			allInfos = append(allInfos, *info)
		}
	}

	return allInfos, nil
}

func listImageTag(ctx context.Context, repo registry.Repository, ref *registry.Reference, filterConfs []kitfile.FilterConf) (*ModelInfo, error) {
	manifestDesc, err := repo.Resolve(ctx, ref.Reference)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", ref.Reference, err)
	}
	manifest, config, err := util.GetManifestAndKitfile(ctx, repo, manifestDesc)
	if err != nil && !errors.Is(err, util.ErrNoKitfile) {
		return nil, fmt.Errorf("failed to read modelkit: %w", err)
	}
	if _, err := mediatype.ModelFormatForManifest(manifest); err != nil {
		return nil, nil
	}
	if !kitfile.KitfileContainsMatchingLayer(config, filterConfs) {
		return nil, nil
	}
	info := &ModelInfo{
		Repo:   ref.Repository,
		Digest: string(manifestDesc.Digest),
		Tags:   []string{ref.Reference},
	}
	info.fill(manifest, config)

	return info, nil
}

func (m *ModelInfo) fill(manifest *ocispec.Manifest, kitfile *artifact.KitFile) {
	m.Size = getModelSize(manifest)
	m.Author = getModelAuthor(kitfile)
	m.ModelName = getModelName(kitfile)
}

func getModelSize(manifest *ocispec.Manifest) string {
	var size int64
	for _, layer := range manifest.Layers {
		size += layer.Size
	}
	return output.FormatBytes(size)
}

func getModelAuthor(kitfile *artifact.KitFile) string {
	if kitfile != nil && len(kitfile.Package.Authors) > 0 {
		return kitfile.Package.Authors[0]
	}
	return "<none>"
}

func getModelName(kitfile *artifact.KitFile) string {
	name := ""
	if kitfile != nil {
		name = kitfile.Package.Name
	}
	if name == "" {
		name = "<none>"
	}
	return name
}
