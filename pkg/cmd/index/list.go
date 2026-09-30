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
	"encoding/json"
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
kit index list

# List local indexes as JSON
kit index list --format json`
)

const listTableFmt = "%s\t%s\t%s\t%s\n"

type listOptions struct {
	configHome string
	format     string
}

type indexListing struct {
	Repo      string   `json:"repo"`
	Digest    string   `json:"digest"`
	Tags      []string `json:"tags"`
	ModelKits *int     `json:"modelKits"`
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
	cmd.Flags().StringVar(&opts.format, "format", "table", "Output format: table or json")
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
		if opts.format == "json" {
			if err := printIndexListJSON(cmd.OutOrStdout(), listings); err != nil {
				return output.Fatalf("Failed to print indexes: %s", err)
			}
			return nil
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
	listings := []indexListing{}
	for _, repo := range indexRepos {
		repository := artifact.FormatRepositoryForDisplay(repo.GetRepoName())
		for _, desc := range repo.GetAllModels() {
			tags := repo.GetTags(desc)
			if tags == nil {
				tags = []string{}
			}
			listings = append(listings, indexListing{
				Repo:      repository,
				Digest:    desc.Digest.String(),
				Tags:      tags,
				ModelKits: countEntries(ctx, repo, desc),
			})
		}
	}
	sort.Slice(listings, func(i, j int) bool {
		return (listings[i].Repo < listings[j].Repo) ||
			((listings[i].Repo == listings[j].Repo) && (listings[i].Digest < listings[j].Digest))
	})
	return listings, nil
}

// countEntries reports how many ModelKits an index references, or nil when the index cannot
// be read.
func countEntries(ctx context.Context, repo local.LocalRepo, desc ocispec.Descriptor) *int {
	indexBytes, err := content.FetchAll(ctx, repo, desc)
	if err != nil {
		return nil
	}
	idx, err := libindex.ParseIndex(indexBytes)
	if err != nil {
		return nil
	}
	count := len(idx.Manifests)
	return &count
}

func printIndexList(out io.Writer, listings []indexListing) {
	tw := tabwriter.NewWriter(out, 0, 2, 3, ' ', 0)
	fmt.Fprintf(tw, listTableFmt, "REPOSITORY", "TAG", "MODELKITS", "DIGEST")
	for _, listing := range listings {
		tags := listing.Tags
		if len(tags) == 0 {
			tags = []string{noneValue}
		}
		modelKits := noneValue
		if listing.ModelKits != nil {
			modelKits = strconv.Itoa(*listing.ModelKits)
		}
		for _, tag := range tags {
			fmt.Fprintf(tw, listTableFmt, listing.Repo, tag, modelKits, listing.Digest)
		}
	}
	tw.Flush()
}

func printIndexListJSON(out io.Writer, listings []indexListing) error {
	jsonBytes, err := json.MarshalIndent(listings, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(jsonBytes))
	return nil
}

func (opts *listOptions) complete(ctx context.Context) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	if opts.format != "table" && opts.format != "json" {
		return fmt.Errorf("unsupported format %q: must be table or json", opts.format)
	}
	return nil
}
