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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

const (
	shortDesc = `Generate a Kitfile for the contents of a directory`
	longDesc  = `Examine the contents of a directory and attempt to generate a basic Kitfile
based on common file formats. Any files whose type (i.e. model, dataset, etc.)
cannot be determined will be included in a code layer.

By default the command will prompt for input for a name and description for the Kitfile`
	example = `# Generate a Kitfile for the current directory:
kit init .

# Generate a Kitfile for files in ./my-model, with name "mymodel" and a description:
kit init ./my-model --name "mymodel" --desc "This is my model's description"

# Generate a Kitfile, overwriting any existing Kitfile:
kit init ./my-model --force`
)

type initOptions struct {
	path                string
	configHome          string
	modelkitName        string
	modelkitDescription string
	modelkitAuthor      string
	overwrite           bool
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

	cmd.Flags().StringVar(&opts.modelkitName, "name", "", "Name for the ModelKit")
	cmd.Flags().StringVar(&opts.modelkitDescription, "desc", "", "Description for the ModelKit")
	cmd.Flags().StringVar(&opts.modelkitAuthor, "author", "", "Author for the ModelKit")
	cmd.Flags().BoolVarP(&opts.overwrite, "force", "f", false, "Overwrite existing Kitfile if present")
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *initOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		var modelPackage *artifact.Package
		if opts.modelkitName != "" || opts.modelkitDescription != "" {
			modelPackage = &artifact.Package{
				Name:        opts.modelkitName,
				Description: opts.modelkitDescription,
			}
		}
		if opts.modelkitAuthor != "" {
			modelPackage.Authors = append(modelPackage.Authors, opts.modelkitAuthor)
		}

		// Check for existing Kitfile
		kitfilePath := filepath.Join(opts.path, constants.DefaultKitfileName)
		if _, err := os.Stat(kitfilePath); err == nil {
			if !opts.overwrite {
				return output.Fatalf("Kitfile already exists at %s. Use '--force' to overwrite", kitfilePath)
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return output.Fatalf("Error checking for existing Kitfile: %s", err)
		}

		dirContents, err := kfgen.DirectoryListingFromFS(opts.path)
		if err != nil {
			return output.Fatalf("Error processing directory: %s", err)
		}
		kitfile, err := kfgen.GenerateKitfile(dirContents, modelPackage)
		if err != nil {
			return output.Fatalf("Error generating Kitfile: %s", err)
		}
		bytes, err := kitfile.MarshalToYAML()
		if err != nil {
			return output.Fatalf("Error formatting Kitfile: %s", err)
		}
		if err := os.WriteFile(kitfilePath, bytes, 0644); err != nil {
			return output.Fatalf("Failed to write Kitfile: %s", err)
		}
		output.Infof("Generated Kitfile:\n\n%s", string(bytes))
		output.Infof("Saved to path '%s'", kitfilePath)
		return nil
	}
}

func (opts *initOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.path = args[0]

	if util.IsInteractiveSession() {
		if opts.modelkitName == "" {
			name, err := util.PromptForInput("Enter a name for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.modelkitName = name
		}
		if opts.modelkitDescription == "" {
			desc, err := util.PromptForInput("Enter a short description for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.modelkitDescription = desc
		}
		if opts.modelkitAuthor == "" {
			author, err := util.PromptForInput("Enter an author for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.modelkitAuthor = author
		}
	}

	return nil
}
