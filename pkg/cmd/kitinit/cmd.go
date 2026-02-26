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
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/hf"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
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
	path                string
	configHome          string
	modelkitName        string
	modelkitDescription string
	modelkitAuthor      string
	overwrite           bool
	remote              bool
	repoRef             string
	token               string
	outputPath          string
	// Computed fields (remote only)
	repo     string
	repoType hf.RepositoryType
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
	cmd.Flags().BoolVar(&opts.remote, "remote", false, "Generate Kitfile from a remote HuggingFace repository")
	cmd.Flags().StringVar(&opts.repoRef, "ref", "main", "Branch or tag for remote repository (requires --remote)")
	cmd.Flags().StringVar(&opts.token, "token", "", "Auth token for remote repository (requires --remote)")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "", "Output path for generated Kitfile ('-' writes to stdout; default: Kitfile in directory for local, stdout for remote)")
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *initOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if !opts.remote {
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

		var dirContents *kfgen.DirectoryListing
		var listErr error
		if opts.remote {
			output.SystemInfof("Fetching file listing from remote repository %s (ref: %s)", opts.repo, opts.repoRef)
			dirContents, listErr = hf.ListFiles(cmd.Context(), opts.repo, opts.repoRef, opts.token, opts.repoType)
			if listErr != nil {
				return output.Fatalf("Error fetching remote repository: %s", listErr)
			}
		} else {
			dirContents, listErr = kfgen.DirectoryListingFromFS(opts.path)
			if listErr != nil {
				return output.Fatalf("Error processing directory: %s", listErr)
			}
		}

		return runInit(dirContents, opts)
	}
}

func runInit(dirContents *kfgen.DirectoryListing, opts *initOptions) error {
	modelPackage := buildPackageFromRepo(opts.repo, opts.modelkitName, opts.modelkitDescription, opts.modelkitAuthor)

	kitfile, err := kfgen.GenerateKitfile(dirContents, modelPackage)
	if err != nil {
		return output.Fatalf("Error generating Kitfile: %s", err)
	}
	bytes, err := kitfile.MarshalToYAML()
	if err != nil {
		return output.Fatalf("Error formatting Kitfile: %s", err)
	}

	if opts.outputPath == "-" {
		fmt.Print(string(bytes))
		return nil
	}
	return writeKitfile(bytes, opts)
}

func writeKitfile(bytes []byte, opts *initOptions) error {
	if _, err := os.Stat(opts.outputPath); err == nil {
		if !opts.overwrite {
			return output.Fatalf("Kitfile already exists at %s. Use '--force' to overwrite", opts.outputPath)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return output.Fatalf("Error checking for existing Kitfile: %s", err)
	}
	if err := os.WriteFile(opts.outputPath, bytes, 0644); err != nil {
		return output.Fatalf("Failed to write Kitfile: %s", err)
	}
	output.Infof("Generated Kitfile:\n\n%s", string(bytes))
	output.Infof("Saved to path '%s'", opts.outputPath)
	return nil
}

func buildPackageFromRepo(repo, name, description, author string) *artifact.Package {
	sections := strings.Split(repo, "/")
	modelPackage := &artifact.Package{}

	if name != "" {
		modelPackage.Name = name
	} else if len(sections) >= 2 {
		modelPackage.Name = sections[len(sections)-1]
	}

	if description != "" {
		modelPackage.Description = description
	}

	if author != "" {
		modelPackage.Authors = append(modelPackage.Authors, author)
	} else if len(sections) >= 2 {
		modelPackage.Authors = append(modelPackage.Authors, sections[len(sections)-2])
	}

	return modelPackage
}

func (opts *initOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.path = args[0]

	if opts.remote {
		repo, repoType, err := hf.ParseHuggingFaceRepo(opts.path)
		if err != nil {
			return fmt.Errorf("invalid HuggingFace repository: %w", err)
		}
		opts.repo = repo
		opts.repoType = repoType
		if opts.outputPath == "" {
			opts.outputPath = "-"
		}
	} else {
		if opts.outputPath == "" {
			opts.outputPath = filepath.Join(opts.path, constants.DefaultKitfileName)
		}
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
	}
	return nil
}
