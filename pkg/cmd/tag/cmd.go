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

package tag

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = "Create a tag that refers to a modelkit"
	longDesc  = `Create or update a tag {target-modelkit} that refers to {source-modelkit}

This command assigns a new tag to an existing modelkit (source-modelkit) or
updates an existing tag, effectively renaming or categorizing modelkits for
better organization and version control. Tags are identifiers linked to
specific modelkit versions within a repository.

A full modelkit reference has the following format:

[HOST[:PORT_NUMBER]/][NAMESPACE/]REPOSITORY[:TAG]

  * HOST: Optional. The registry hostname where the ModelKit is located.
    Defaults to localhost if unspecified. Must follow standard DNS rules
    (excluding underscores).

  * PORT_NUMBER: Optional. Specifies the registry's port number if a hostname
    is provided.

  * NAMESPACE: Represents a user or organization's namespace, consisting of
    slash-separated components that may include lowercase letters, digits, and
    specific separators (periods, underscores, hyphens).

  * REPOSITORY: The name of the repository, typically corresponding to the
    modelkit's name.

  * TAG: A human-readable identifier for the modelkit version or variant.
    Valid ASCII characters include lowercase and uppercase letters, digits,
    underscores, periods, and hyphens. It cannot start with a period or hyphen
    and is limited to 128 characters.

Tagging is a powerful way to manage different versions or configurations of
your modelkits, making it easier to organize, retrieve, and deploy specific
iterations. Ensure tags are meaningful and consistent across your team or
organization to maintain clarity and avoid confusion.`

	example = `kit tag myregistry.com/myrepo/mykit:latest myregistry.com/myrepo/mykit:v1.0.0`
)

type tagOptions struct {
	configHome string
	sourceRef  *registry.Reference
	targetRef  *registry.Reference
}

func (opts *tagOptions) complete(ctx context.Context, args []string) error {

	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	modelRef, _, err := util.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	opts.sourceRef = modelRef

	modelRef, _, err = util.ParseReference(args[1])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	opts.targetRef = modelRef
	return nil
}

func TagCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "tag SOURCE_MODELKIT[:TAG] TARGET_MODELKIT[:TAG]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(&tagOptions{}),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	cmd.Args = cobra.ExactArgs(2)
	return cmd
}

func runCommand(opts *tagOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		err := RunTag(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to tag modelkit: %s", err)
		}
		output.Infof("Modelkit %s tagged as %s", opts.sourceRef, opts.targetRef)
		return nil
	}
}
