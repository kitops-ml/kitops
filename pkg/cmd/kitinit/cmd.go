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

package kitinit

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/kit"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

const (
	shortDesc = `Generate a Kitfile for the contents of a directory or remote repository`
	longDesc  = `Examine the contents of a directory or remote repository and attempt to generate
a basic Kitfile based on common file formats. Any files whose type (i.e. model,
dataset, etc.) cannot be determined will be included in a code layer.

For local directories, the generated Kitfile is saved in the target directory by
default. Use --output to specify a different path, or --output=- for stdout.
For remote repositories (--remote), the Kitfile is printed to stdout by default.

By default the command will prompt for input for a name and description for the Kitfile.`

	example = `# Generate a Kitfile for the current directory:
kit init .

# Generate a Kitfile for files in ./my-model, with name "mymodel" and a description:
kit init ./my-model --name "mymodel" --desc "This is my model's description"

# Generate a Kitfile, overwriting any existing Kitfile:
kit init ./my-model --force

# Generate a Kitfile for a remote HuggingFace model:
kit init https://huggingface.co/myorg/mymodel --remote

# Generate a Kitfile for a HuggingFace dataset:
kit init huggingface.co/datasets/myorg/mydataset --remote

# Generate a Kitfile for a remote repository with a specific ref:
kit init myorg/mymodel --remote --ref v1.0

# Save the generated Kitfile to a specific path:
kit init myorg/mymodel --remote --output ./Kitfile`
)

type initOptions struct {
	kit.InitOptions
}

func InitCommand() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:     "init [flags] PATH",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Name for the ModelKit")
	cmd.Flags().StringVar(&opts.Description, "desc", "", "Description for the ModelKit")
	cmd.Flags().StringVar(&opts.Author, "author", "", "Author for the ModelKit")
	cmd.Flags().BoolVarP(&opts.Overwrite, "force", "f", false, "Overwrite existing Kitfile if present")
	cmd.Flags().BoolVar(&opts.Remote, "remote", false, "Generate Kitfile from a remote HuggingFace repository")
	cmd.Flags().StringVar(&opts.RepoRef, "ref", "main", "Branch or tag for remote repository (requires --remote)")
	cmd.Flags().StringVar(&opts.Token, "token", "", "Auth token for remote repository (requires --remote)")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "Output path for generated Kitfile ('-' writes to stdout; default: Kitfile in directory for local, stdout for remote)")
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *initOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if !opts.Remote {
			if cmd.Flags().Changed("ref") {
				return output.Fatalf("Invalid arguments: --ref requires --remote")
			}
			if cmd.Flags().Changed("token") {
				return output.Fatalf("Invalid arguments: --token requires --remote")
			}
		}

		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		result, err := kit.Init(cmd.Context(), &opts.InitOptions)
		if err != nil {
			return output.Fatalf("%s", err)
		}

		if opts.OutputPath == "-" || opts.OutputPath == "" {
			fmt.Print(string(result.KitfileBytes))
		} else if result.WrittenPath != "" {
			output.Infof("Generated Kitfile:\n\n%s", string(result.KitfileBytes))
			output.Infof("Saved to path '%s'", result.WrittenPath)
		}
		return nil
	}
}

func (opts *initOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.ConfigHome = configHome
	opts.Path = args[0]

	// Interactive prompts for local mode only (library callers provide these directly)
	if !opts.Remote && util.IsInteractiveSession() {
		if opts.Name == "" {
			name, err := util.PromptForInput("Enter a name for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.Name = name
		}
		if opts.Description == "" {
			desc, err := util.PromptForInput("Enter a short description for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.Description = desc
		}
		if opts.Author == "" {
			author, err := util.PromptForInput("Enter an author for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.Author = author
		}
	}
	return nil
}
