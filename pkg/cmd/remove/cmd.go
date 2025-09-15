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

package remove

import (
	"context"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Remove a modelkit from local storage`
	longDesc  = `Removes a modelkit from storage on the local disk.

The model to be removed may be specifed either by a tag or by a digest. If
specified by digest, that modelkit will be removed along with any tags that
might refer to it. If specified by tag (and the --force flag is not used),
the modelkit will only be removed if no other tags refer to it; otherwise
it is only untagged.`

	examples = `# Remove modelkit by tag
kit remove my-registry.com/my-org/my-repo:my-tag

# Remove modelkit by digest
kit remove my-registry.com/my-org/my-repo@sha256:<digest>

# Remove multiple tags for a modelkit
kit remove my-registry.com/my-org/my-repo:tag1,tag2,tag3

# Remove all untagged modelkits
kit remove --all

# Remove all locally stored modelkits
kit remove --all --force`
)

type removeOptions struct {
	options.NetworkOptions
	configHome  string
	forceDelete bool
	removeAll   bool
	remote      bool
	modelRef    *registry.Reference
	extraTags   []string
}

func (opts *removeOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	if len(args) > 0 {
		modelRef, extraTags, err := util.ParseReference(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}
		opts.modelRef = modelRef
		opts.extraTags = extraTags
	}

	if opts.remote && (opts.forceDelete || opts.removeAll) {
		return fmt.Errorf("cannot use --all or --force with --remote")
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	printConfig(opts)
	return nil
}

func RemoveCommand() *cobra.Command {
	opts := &removeOptions{}
	cmd := &cobra.Command{
		Use:     "remove [flags] registry/repository[:tag|@digest]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		RunE:    runCommand(opts),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if cmd.Flags().Changed("all") || cmd.Flags().Changed("remote") {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}
	cmd.Flags().BoolVarP(&opts.forceDelete, "force", "f", false, "remove modelkit and all other tags that refer to it")
	cmd.Flags().BoolVarP(&opts.removeAll, "all", "a", false, "remove all untagged modelkits")
	cmd.Flags().BoolVarP(&opts.remote, "remote", "r", false, "remove modelkit from remote registry")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	cmd.Args = func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			if opts.removeAll {
				return nil
			}
			return fmt.Errorf("modelkit is required for remove unless --all is specified")
		case 1:
			if opts.removeAll {
				return fmt.Errorf("modelkit should not be specified when --all flag is used")
			}
			return nil
		default:
			return cobra.MaximumNArgs(1)(cmd, args)
		}
	}

	return cmd
}

func runCommand(opts *removeOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		var err error
		switch {
		case opts.modelRef != nil:
			if opts.remote {
				err = removeRemoteModel(cmd.Context(), opts)
			} else {
				err = removeModel(cmd.Context(), opts)
			}
		case opts.removeAll && !opts.forceDelete:
			err = removeUntaggedModels(cmd.Context(), opts)
		case opts.removeAll && opts.forceDelete:
			err = removeAllModels(cmd.Context(), opts)
		}
		if err != nil {
			return output.Fatalf(err.Error())
		}
		return nil
	}
}

func printConfig(opts *removeOptions) {
	if opts.modelRef != nil {
		displayRef := util.FormatRepositoryForDisplay(opts.modelRef.String())
		output.Debugf("Removing %s and additional tags: [%s]", displayRef, strings.Join(opts.extraTags, ", "))
	}
	if opts.removeAll {
		if opts.forceDelete {
			output.Debugf("Removing all locally-stored modelkits")
		} else {
			output.Debugf("Removing all untagged modelkits")
		}
	}
}
