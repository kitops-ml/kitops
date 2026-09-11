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

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	pushShortDesc = `Upload a ModelKit index to a specified registry`
	pushLongDesc  = `Push a ModelKit index to a remote registry. Only the index is pushed; the
ModelKits it references must already be in the destination repository.`

	pushExample = `# Push an index to a remote registry
kit index push registry.example.com/my-org/my-model:all

# Push a local index to a remote registry
kit index push my-org/my-model:all registry.example.com/my-org/my-model:all`
)

type pushOptions struct {
	options.NetworkOptions
	configHome   string
	srcIndexRef  *registry.Reference
	destIndexRef *registry.Reference
}

func indexPushCommand() *cobra.Command {
	opts := &pushOptions{}

	cmd := &cobra.Command{
		Use:     "push [flags] SOURCE [DESTINATION]",
		Short:   pushShortDesc,
		Long:    pushLongDesc,
		Example: pushExample,
		RunE:    runPushCommand(opts),
		Args:    cobra.RangeArgs(1, 2),
	}

	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runPushCommand(opts *pushOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		return output.Fatalf("Not implemented: kit index push")
	}
}

func (opts *pushOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	srcRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference %s: %w", args[0], err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("reference cannot include multiple tags")
	}
	if srcRef.Reference == "" {
		output.Infof("No tag specified for push. Using 'latest' as default ('%s:latest')", args[0])
		srcRef.Reference = "latest"
	}
	opts.srcIndexRef = srcRef

	if len(args) == 1 {
		opts.destIndexRef = srcRef
	} else {
		destRef, extraTags, err := artifact.ParseReference(args[1])
		if err != nil {
			return fmt.Errorf("failed to parse target reference %s: %w", args[1], err)
		}
		if len(extraTags) > 0 {
			return fmt.Errorf("target reference cannot include multiple tags")
		}
		if destRef.Reference == "" {
			output.Infof("No tag specified for push target. Using 'latest' as default ('%s:latest')", args[1])
			destRef.Reference = "latest"
		}
		opts.destIndexRef = destRef
	}

	if opts.destIndexRef.Registry == artifact.DefaultRegistry {
		return fmt.Errorf("registry is required when pushing")
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}
