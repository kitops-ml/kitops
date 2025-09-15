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

package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/spf13/cobra"
)

func GetLocalModelKitsCompletion(ctx context.Context, toComplete string) []string {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		cobra.CompErrorln("Failed to get KitOps config directory")
		return nil
	}
	storageRoot := constants.StoragePath(configHome)
	localRepos, err := local.GetAllLocalRepos(storageRoot)
	if err != nil {
		cobra.CompErrorln("Failed to list local ModelKits")
		return nil
	}
	hasColon := strings.Contains(toComplete, ":")
	hasAt := strings.Contains(toComplete, "@")

	var completions []string
	for _, repo := range localRepos {
		repoName := util.FormatRepositoryForDisplay(repo.GetRepoName())
		tags, digests := getTagsAndDigestsForRepo(repo)
		switch {
		// Note: this case _has_ to come first, as digests themselves contain colons
		case hasAt:
			for _, digest := range digests {
				completions = append(completions, fmt.Sprintf("%s@%s", repoName, digest))
			}
		case hasColon:
			for _, tag := range tags {
				completions = append(completions, fmt.Sprintf("%s:%s", repoName, tag))
			}
		default:
			switch len(tags) {
			case 0:
				completions = append(completions, repoName)
			case 1:
				completions = append(completions, fmt.Sprintf("%s:%s", repoName, tags[0]))
			default:
				completions = append(completions, repoName+":")
			}
		}
	}
	return completions
}

func getTagsAndDigestsForRepo(repo local.LocalRepo) ([]string, []string) {
	manifestDescs := repo.GetAllModels()
	var tags []string
	var digests []string
	for _, desc := range manifestDescs {
		repoTags := repo.GetTags(desc)
		tags = append(tags, repoTags...)
		digests = append(digests, desc.Digest.String())
	}
	return tags, digests
}
