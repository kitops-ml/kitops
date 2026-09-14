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

package index

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/content"
)

const (
	listShortDesc = `List ModelKit indexes in local storage`
	listLongDesc  = `List ModelKit indexes in local storage, with the number of ModelKits each one
references.`

	listExample = `# List local indexes
kit index list`
)

const listTableFmt = "%s\t%s\t%s\t%s\n"

type listOptions struct {
	configHome string
}

type indexListing struct {
	repo      string
	tags      []string
	modelKits string
	digest    string
}

func indexListCommand() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   listShortDesc,
		Long:    listLongDesc,
		Example: listExample,
		RunE:    runListCommand(opts),
		Args:    cobra.NoArgs,
	}
	cmd.Flags().SortFlags = false

	return cmd
}

func runListCommand(opts *listOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context()); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		listings, err := listLocalIndexes(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to list indexes: %s", err)
		}
		printIndexList(cmd.OutOrStdout(), listings)
		return nil
	}
}

func listLocalIndexes(ctx context.Context, opts *listOptions) ([]indexListing, error) {
	indexRepos, err := local.GetAllLocalIndexRepos(constants.StoragePath(opts.configHome))
	if err != nil {
		return nil, err
	}
	var listings []indexListing
	for _, repo := range indexRepos {
		repository := artifact.FormatRepositoryForDisplay(repo.GetRepoName())
		for _, desc := range repo.GetAllModels() {
			listings = append(listings, indexListing{
				repo:      repository,
				tags:      repo.GetTags(desc),
				modelKits: countEntries(ctx, repo, desc),
				digest:    desc.Digest.String(),
			})
		}
	}
	sort.Slice(listings, func(i, j int) bool {
		return (listings[i].repo < listings[j].repo) ||
			((listings[i].repo == listings[j].repo) && (listings[i].digest < listings[j].digest))
	})
	return listings, nil
}

// countEntries reports how many ModelKits an index references, or that it is unknown when
// the index cannot be read.
func countEntries(ctx context.Context, repo local.LocalRepo, desc ocispec.Descriptor) string {
	indexBytes, err := content.FetchAll(ctx, repo, desc)
	if err != nil {
		return noneValue
	}
	idx, err := libindex.ParseIndex(indexBytes)
	if err != nil {
		return noneValue
	}
	return strconv.Itoa(len(idx.Manifests))
}

func printIndexList(out io.Writer, listings []indexListing) {
	tw := tabwriter.NewWriter(out, 0, 2, 3, ' ', 0)
	fmt.Fprintf(tw, listTableFmt, "REPOSITORY", "TAG", "MODELKITS", "DIGEST")
	for _, listing := range listings {
		tags := listing.tags
		if len(tags) == 0 {
			tags = []string{noneValue}
		}
		for _, tag := range tags {
			fmt.Fprintf(tw, listTableFmt, listing.repo, tag, listing.modelKits, listing.digest)
		}
	}
	tw.Flush()
}

func (opts *listOptions) complete(ctx context.Context) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	return nil
}
