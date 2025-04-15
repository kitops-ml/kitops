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

package logout

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

const (
	shortDesc = `Log out from an OCI registry`
	longDesc  = `Log out from a specified OCI-compatible registry. Any saved credentials are
removed from storage.`

	example = `# Log out from ghcr.io
kit logout ghcr.io`
)

type logoutOptions struct {
	credentialStoreHome string
	registry            string
}

func LogoutCommand() *cobra.Command {
	opts := &logoutOptions{}

	cmd := &cobra.Command{
		Use:     "logout [flags] REGISTRY",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
	}
	cmd.Args = cobra.ExactArgs(1)
	return cmd
}

func runCommand(opts *logoutOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		err := logout(cmd.Context(), opts.registry, opts.credentialStoreHome)
		if err != nil {
			return output.Fatalln(err)
		}
		return nil
	}
}

func (opts *logoutOptions) complete(ctx context.Context, args []string) error {
	opts.registry = args[0]
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.credentialStoreHome = constants.CredentialsPath(configHome)
	return nil
}
