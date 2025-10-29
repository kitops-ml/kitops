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
	"sort"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
)

func listLocalKits(ctx context.Context, opts *listOptions) ([]modelInfo, error) {
	storageRoot := constants.StoragePath(opts.configHome)

	localRepos, err := local.GetAllLocalRepos(storageRoot)
	if err != nil {
		return nil, err
	}
	var allInfo []modelInfo
	for _, repo := range localRepos {
		infos, err := readInfoFromRepo(ctx, repo)
		if err != nil {
			return nil, err
		}
		allInfo = append(allInfo, infos...)
	}

	return allInfo, nil
}

func readInfoFromRepo(ctx context.Context, repo local.LocalRepo) ([]modelInfo, error) {
	var infos []modelInfo
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
		tags := repo.GetTags(manifestDesc)
		// Strip localhost from repo if present, since we added it
		repository := util.FormatRepositoryForDisplay(repo.GetRepoName())
		if repository == "" {
			repository = "<none>"
		}
		info := modelInfo{
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
