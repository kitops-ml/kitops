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

package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

const (
	shortDesc = "Verify the ModelKit signature and attestation. Runs both verify and verify-attestation. Use --verify.* and --verify-attestation.* for flags specific to each step."
	example   = `kit verify --key cosign.pub --verify.insecure-ignore-tlog=true DIGEST`
)

type verifyOptions struct {
	configHome string
	cosignArgs []string
}

func (opts *verifyOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.cosignArgs = args

	return nil
}

func VerifyCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "verify [FLAGS]",
		Short:   shortDesc,
		Example: example,
		RunE:    runCommand([]verifyOptions{}),
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

func runCommand(opts []verifyOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		commands := []string{"verify", "verify-attestation"}
		argsnew := [][]string{{}, {}}
		for _, val := range args {
			if val, ok := strings.CutPrefix(val, "--verify."); ok {
				argsnew[0] = append(argsnew[0], "--"+val)
			} else if val, ok := strings.CutPrefix(val, "--verify-attestation."); ok {
				argsnew[1] = append(argsnew[1], "--"+val)
			} else {
				argsnew[0] = append(argsnew[0], val)
				argsnew[1] = append(argsnew[1], val)
			}
		}

		for i := range 2 {
			opts = append(opts, verifyOptions{})
			argsnew[i] = append([]string{commands[i]}, argsnew[i]...)
			if err := opts[i].complete(cmd.Context(), argsnew[i]); err != nil {
				return output.Fatalf("Invalid arguments: %s", err)
			}
		}

		for i := range len(opts) {
			err := RunVerify(cmd.Context(), opts[i])
			if err != nil {
				return output.Fatalf("Failed to %s: %s", commands[i], err)
			}
		}
		output.Infof("Modelkit signed")
		return nil
	}
}
