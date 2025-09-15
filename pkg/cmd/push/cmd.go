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

package push

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

const (
	shortDesc = `Upload a modelkit to a specified registry`
	longDesc  = `This command pushes modelkits from local storage to a remote registry.

If specified without a destination, the ModelKit must be tagged locally before
pushing.`

	example = `# Push the ModelKit tagged 'latest' to a remote registry
kit push registry.example.com/my-org/my-model:latest

# Push a ModelKit to a remote registry by digest
kit push registry.example.com/my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Push local modelkit 'mymodel:1.0.0' to a remote registry
kit push mymodel:1.0.0 registry.example.com/my-org/my-model:latest`
)

type pushOptions struct {
	options.NetworkOptions
	configHome   string
	srcModelRef  *registry.Reference
	destModelRef *registry.Reference
}

func (opts *pushOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	srcRef, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference %s: %w", args[0], err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("reference cannot include multiple tags")
	}
	opts.srcModelRef = srcRef
	if len(args) == 1 {
		opts.destModelRef = srcRef
	} else {
		destRef, extraTags, err := util.ParseReference(args[1])
		if err != nil {
			return fmt.Errorf("failed to parse target reference %s: %w", args[1], err)
		}
		if len(extraTags) > 0 {
			return fmt.Errorf("target reference cannot include multiple tags")
		}
		opts.destModelRef = destRef
	}

	if opts.destModelRef.Registry == "localhost" {
		return fmt.Errorf("registry is required when pushing")
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func PushCommand() *cobra.Command {
	opts := &pushOptions{}
	cmd := &cobra.Command{
		Use:     "push [flags] SOURCE [DESTINATION]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.RangeArgs(1, 2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 2 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *pushOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		remoteRepo, err := remote.NewRepository(
			cmd.Context(),
			opts.destModelRef.Registry,
			opts.destModelRef.Repository,
			&opts.NetworkOptions,
		)
		if err != nil {
			return output.Fatalln(err)
		}

		localRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.srcModelRef)
		if err != nil {
			return output.Fatalln(err)
		}

		if opts.srcModelRef.String() != opts.destModelRef.String() {
			output.Infof("Pushing %s to %s", opts.srcModelRef.String(), opts.destModelRef.String())
		} else {
			output.Infof("Pushing %s", opts.srcModelRef.String())
		}
		desc, err := PushModel(cmd.Context(), localRepo, remoteRepo, opts)
		respErr := &errcode.ErrorResponse{}
		if ok := errors.As(err, &respErr); ok {
			output.Debugf("Got error pushing: %s", err)
			errMsg := fmt.Sprintf("Failed to push: got response %d (%s) from remote", respErr.StatusCode, http.StatusText(respErr.StatusCode))
			switch respErr.StatusCode {
			case http.StatusUnauthorized:
				errMsg = fmt.Sprintf("%s. Ensure the repository exists and you have push access to it.", errMsg)
			}
			return output.Fatalf(errMsg)
		} else if err != nil {
			return output.Fatalf("Failed to push: %s.", err)
		}
		output.Infof("Pushed %s", desc.Digest)
		return nil
	}
}
