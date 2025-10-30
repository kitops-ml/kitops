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

package pull

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Retrieve modelkits from a remote registry to your local environment.`
	longDesc  = `Downloads modelkits from a specified registry. The downloaded modelkits
are stored in the local registry.`

	example = `# Pull the latest version of a modelkit from a remote registry
kit pull registry.example.com/my-model:latest`
)

type pullOptions struct {
	options.NetworkOptions
	configHome string
	modelRef   *registry.Reference
}

func (opts *pullOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	modelRef, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	if modelRef.Registry == "localhost" {
		return fmt.Errorf("registry is required when pulling")
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("reference cannot include multiple tags")
	}
	if modelRef.Reference == "" {
		output.Infof("No tag specified for pull. Using 'latest' as default ('%s:latest')", args[0])
		modelRef.Reference = "latest"
	}
	opts.modelRef = modelRef

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func PullCommand() *cobra.Command {
	opts := &pullOptions{}
	cmd := &cobra.Command{
		Use:     "pull [flags] registry/repository[:tag|@digest]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
	}

	cmd.Args = cobra.ExactArgs(1)
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *pullOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		output.Infof("Pulling %s", opts.modelRef.String())
		desc, err := runPull(cmd.Context(), opts)
		if err != nil {
			return output.Fatalln(err)
		}
		output.Infof("Pulled %s", desc.Digest)
		return nil
	}
}
