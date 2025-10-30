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

package inspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Inspect a modelkit's manifest`
	longDesc  = `Print the contents of a modelkit manifest to the screen.

By default, kit will check local storage for the specified modelkit. To
inspect a modelkit stored on a remote registry, use the --remote flag.`
	example = `# Inspect a local modelkit:
kit inspect mymodel:mytag

# Inspect a local modelkit by digest:
kit inspect mymodel@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Inspect a remote modelkit:
kit inspect --remote registry.example.com/my-model:1.0.0`
)

type inspectOptions struct {
	options.NetworkOptions
	configHome  string
	checkRemote bool
	modelRef    *registry.Reference
}

func InspectCommand() *cobra.Command {
	opts := &inspectOptions{}

	cmd := &cobra.Command{
		Use:     "inspect [flags] MODELKIT",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if cmd.Flags().Changed("remote") {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	opts.AddNetworkFlags(cmd)
	cmd.Flags().BoolVarP(&opts.checkRemote, "remote", "r", false, "Check remote registry instead of local storage")
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *inspectOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		inspectInfo, err := inspectReference(cmd.Context(), opts)
		if err != nil {
			if errors.Is(err, errdef.ErrNotFound) {
				return output.Fatalf("Could not find modelkit %s", util.FormatRepositoryForDisplay(opts.modelRef.String()))
			}
			return output.Fatalf("Error resolving modelkit: %s", err)
		}
		jsonBytes, err := json.MarshalIndent(inspectInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("Error formatting manifest: %w", err)
		}
		output.Infoln(string(jsonBytes))
		return nil
	}
}

func (opts *inspectOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	ref, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return err
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if ref.Reference == "" {
		return fmt.Errorf("missing tag or digest from ModelKit reference '%s'", args[0])
	}
	opts.modelRef = ref

	if opts.modelRef.Registry == util.DefaultRegistry && opts.checkRemote {
		return fmt.Errorf("can not check remote: %s does not contain registry", util.FormatRepositoryForDisplay(opts.modelRef.String()))
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}
