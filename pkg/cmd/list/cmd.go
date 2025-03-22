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
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `List modelkits in a repository`
	longDesc  = `Displays a list of modelkits available in a repository.

This command provides an overview of modelkits stored either in the local
repository or a specified remote repository. It displays each modelkit along
with its associated tags and the cumulative size of its contents. Modelkits
comprise multiple artifacts, including models, datasets, code, and
configuration, designed to enhance reusability and modularity. However, this
command focuses on the aggregate rather than listing individual artifacts.

Each modelkit entry includes its DIGEST, a unique identifier that ensures
distinct versions of a modelkit are easily recognizable, even if they share
the same name or tags. Modelkits with multiple tags or repository names will
appear multiple times in the list, distinguished by their DIGEST.

The SIZE displayed for each modelkit represents the total storage space
occupied by all its components.`

	example = `# List local modelkits
kit list

# List modelkits from a remote repository
kit list registry.example.com/my-namespace/my-model`
)

type listOptions struct {
	options.NetworkOptions
	configHome string
	remoteRef  *registry.Reference
}

func (opts *listOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	if len(args) > 0 {
		remoteRef, extraTags, err := util.ParseReference(args[0])
		if err != nil {
			return fmt.Errorf("invalid reference: %w", err)
		}
		if len(extraTags) > 0 {
			return fmt.Errorf("repository cannot reference multiple tags")
		}
		opts.remoteRef = remoteRef
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	printConfig(opts)
	return nil
}

// ListCommand represents the models command
func ListCommand() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list [flags] [REPOSITORY]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
	}

	cmd.Args = cobra.MaximumNArgs(1)
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *listOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		var allInfoLines []string
		if opts.remoteRef == nil {
			lines, err := listLocalKits(cmd.Context(), opts)
			if err != nil {
				return output.Fatalln(err)
			}
			allInfoLines = lines
		} else {
			lines, err := listRemoteKits(cmd.Context(), opts)
			if err != nil {
				return output.Fatalln(err)
			}
			allInfoLines = lines
		}
		printSummary(cmd.OutOrStdout(), allInfoLines)
		return nil
	}
}

func printSummary(w io.Writer, lines []string) {
	tw := tabwriter.NewWriter(w, 0, 2, 3, ' ', 0)
	fmt.Fprintln(tw, listTableHeader)
	for _, line := range lines {
		fmt.Fprintln(tw, line)
	}
	tw.Flush()
}

func printConfig(opts *listOptions) {
	if opts.remoteRef != nil {
		output.Debugf("Listing remote model kits in %s", opts.remoteRef.String())
	}
}
