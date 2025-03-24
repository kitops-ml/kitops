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

package login

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	shortDesc = `Log in to an OCI registry`
	longDesc  = `Log in to a specified OCI-compatible registry. Credentials are saved and used
automatically for future CLI operations`

	example = `# Login to ghcr.io
kit login ghcr.io -u github_user -p personal_token

# Login to docker.io with password from stdin
kit login docker.io --password-stdin -u docker_user`
)

type loginOptions struct {
	options.NetworkOptions
	registry          string
	configHome        string
	credential        auth.Credential
	username          string
	password          string
	passwordFromStdIn bool
}

func LoginCommand() *cobra.Command {
	opts := &loginOptions{}

	cmd := &cobra.Command{
		Use:     "login [flags] [REGISTRY]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
	}

	cmd.Args = cobra.ExactArgs(1)
	cmd.Flags().StringVarP(&opts.username, "username", "u", "", "registry username")
	cmd.Flags().StringVarP(&opts.password, "password", "p", "", "registry password or token")
	cmd.Flags().BoolVar(&opts.passwordFromStdIn, "password-stdin", false, "read password from stdin")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *loginOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		err := login(cmd.Context(), opts)
		if err != nil {
			return output.Fatalln(err)
		}
		return nil
	}
}

func (opts *loginOptions) complete(ctx context.Context, args []string) error {
	opts.registry = args[0]
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	if opts.password != "" && opts.username != "" {
		output.Infof("Warning: Using --password via CLI is insecure. Consider using --password-stdin instead")
	}

	username := opts.username
	password := opts.password
	if opts.passwordFromStdIn {
		if password != "" {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}
		readPass, err := readPasswordFromStdin()
		if err != nil {
			return err
		} else if readPass == "" {
			return fmt.Errorf("failed to read password from stdin: got empty string")
		}
		password = readPass
	}

	if password == "" {
		// Prompt for password (and username, if necessary)
		var err error
		if username == "" {
			username, err = util.PromptForInput("Username: ", false)
			if err != nil {
				return err
			}
		}
		password, err = util.PromptForInput("Password: ", true)
		if err != nil {
			return err
		}
		opts.credential = auth.Credential{
			Username: username,
			Password: password,
		}
	} else {
		// If username is empty, assume password is an OAuth token
		if username == "" {
			opts.credential = auth.Credential{
				RefreshToken: password,
			}
		} else {
			opts.credential = auth.Credential{
				Username: username,
				Password: password,
			}
		}
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func readPasswordFromStdin() (string, error) {
	passwd, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read password from standard input")
	}
	return strings.TrimSpace(string(passwd)), err
}
