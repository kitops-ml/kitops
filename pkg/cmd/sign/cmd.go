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

package sign

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

const (
	shortDesc = "Sign the supplied container image. Use the same flags as cosign."
	example   = `kit sign --key cosign.key --tlog-upload=false myimage:latest`
)

type signOptions struct {
	configHome string
	cosignArgs []string
}

func SignCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "sign [flags]",
		Short:   shortDesc,
		Example: example,
		RunE:    runCommand(&signOptions{}),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
		DisableFlagParsing: true,
	}

	return cmd
}

func (opts *signOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.cosignArgs = args
	return nil
}

func runCommand(opts *signOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		args = append([]string{"sign"}, args...)
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		err := RunSign(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to sign: %s", err)
		}
		output.Infof("Modelkit signed")
		return nil
	}
}
