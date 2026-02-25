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
	"syscall"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/hf"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	shortDesc = `Generate a Kitfile for the contents of a directory or remote repository`
	longDesc  = `Examine the contents of a directory or remote repository and attempt to generate
a basic Kitfile based on common file formats. Any files whose type (i.e. model,
dataset, etc.) cannot be determined will be included in a code layer.

For local directories, the generated Kitfile is saved in the target directory.
For remote repositories (e.g. HuggingFace), the Kitfile is printed to stdout
or saved to a path specified with --output.

By default the command will prompt for input for a name and description for the Kitfile.`

	example = `# Generate a Kitfile for the current directory:
kit init .

# Generate a Kitfile for files in ./my-model, with name "mymodel" and a description:
kit init ./my-model --name "mymodel" --desc "This is my model's description"

# Generate a Kitfile, overwriting any existing Kitfile:
kit init ./my-model --force

# Generate a Kitfile for a remote HuggingFace model:
kit init https://huggingface.co/myorg/mymodel

# Generate a Kitfile for a HuggingFace dataset:
kit init huggingface.co/datasets/myorg/mydataset

# Generate a Kitfile for a remote repository with a specific ref:
kit init myorg/mymodel --ref v1.0

# Save the generated Kitfile to a specific path:
kit init myorg/mymodel --output ./Kitfile`
)

type initOptions struct {
	path                string
	configHome          string
	modelkitName        string
	modelkitDescription string
	modelkitAuthor      string
	overwrite           bool
	// Remote repository options
	repoRef    string
	token      string
	outputPath string
	// Computed fields
	isRemote bool
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
	cmd.Flags().StringVar(&opts.repoRef, "ref", "main", "Branch or tag to use for remote repositories")
	cmd.Flags().StringVar(&opts.token, "token", "", "Token for authentication with remote repositories")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "", "Output path for generated Kitfile (default: stdout for remotes, Kitfile in directory for local)")
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *initOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		if opts.isRemote {
			return runRemoteInit(cmd.Context(), opts)
		}
		return runLocalInit(opts)
	}
}

func runLocalInit(opts *initOptions) error {
	var modelPackage *artifact.Package
	if opts.modelkitName != "" || opts.modelkitDescription != "" {
		modelPackage = &artifact.Package{
			Name:        opts.modelkitName,
			Description: opts.modelkitDescription,
		}
	}
	if opts.modelkitAuthor != "" {
		if modelPackage == nil {
			modelPackage = &artifact.Package{}
		}
		modelPackage.Authors = append(modelPackage.Authors, opts.modelkitAuthor)
	}

	kitfilePath := opts.outputPath
	if kitfilePath == "" {
		kitfilePath = filepath.Join(opts.path, constants.DefaultKitfileName)
	}

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

func runRemoteInit(ctx context.Context, opts *initOptions) error {
	if opts.outputPath == "" {
		output.SystemInfof("Fetching file listing from remote repository %s (ref: %s)", opts.repo, opts.repoRef)
	} else {
		output.Infof("Fetching file listing from remote repository %s (ref: %s)", opts.repo, opts.repoRef)
	}

	dirContents, err := hf.ListFiles(ctx, opts.repo, opts.repoRef, opts.token, opts.repoType)
	if err != nil {
		return output.Fatalf("Error fetching remote repository: %s", err)
	}

	modelPackage := buildPackageFromRepo(opts.repo, opts.modelkitName, opts.modelkitDescription, opts.modelkitAuthor)

	kitfile, err := kfgen.GenerateKitfile(dirContents, modelPackage)
	if err != nil {
		return output.Fatalf("Error generating Kitfile: %s", err)
	}
	bytes, err := kitfile.MarshalToYAML()
	if err != nil {
		return output.Fatalf("Error formatting Kitfile: %s", err)
	}

	if opts.outputPath != "" {
		if _, err := os.Stat(opts.outputPath); err == nil {
			if !opts.overwrite {
				return output.Fatalf("File already exists at %s. Use '--force' to overwrite", opts.outputPath)
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return output.Fatalf("Error checking for existing file: %s", err)
		}

		if err := os.WriteFile(opts.outputPath, bytes, 0644); err != nil {
			return output.Fatalf("Failed to write Kitfile: %s", err)
		}
		output.Infof("Generated Kitfile:\n\n%s", string(bytes))
		output.Infof("Saved to path '%s'", opts.outputPath)
	} else {
		fmt.Print(string(bytes))
	}

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
	if opts.path == "~" || strings.HasPrefix(opts.path, "~/") || strings.HasPrefix(opts.path, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		if opts.path == "~" {
			opts.path = home
		} else {
			opts.path = filepath.Join(home, opts.path[2:])
		}
	}

	opts.isRemote, opts.repo, opts.repoType = detectRemoteRepo(opts.path)

	if opts.isRemote {
		return opts.completeRemote()
	}
	return opts.completeLocal()
}

func (opts *initOptions) completeLocal() error {
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

func (opts *initOptions) completeRemote() error {
	// For remote repos, only prompt if:
	// 1. stdin is a terminal (interactive session)
	// 2. stdout is a terminal (not redirected with >)
	// 3. no output path specified (output goes to stdout)
	stdoutIsTerminal := term.IsTerminal(int(syscall.Stdout))
	if util.IsInteractiveSession() && stdoutIsTerminal && opts.outputPath == "" {
		if opts.modelkitDescription == "" {
			desc, err := util.PromptForInput("Enter a short description for the ModelKit: ", false)
			if err != nil {
				return err
			}
			opts.modelkitDescription = desc
		}
	}
	return nil
}

func detectRemoteRepo(path string) (isRemote bool, repo string, repoType hf.RepositoryType) {
	// First, check if this looks like an explicit local filesystem path
	// Local paths include: ".", "..", paths starting with "./", "../", "/", or containing backslashes
	if isLocalPath(path) {
		return false, "", hf.RepoTypeUnknown
	}

	// Check if the path exists on the filesystem
	// This handles cases like "models/my-model" which could be either local or remote
	if _, err := os.Stat(path); err == nil || errors.Is(err, fs.ErrPermission) {
		// Path exists locally, or permission denied (path exists but is unreadable) - treat as local
		return false, "", hf.RepoTypeUnknown
	}

	// Path doesn't exist locally - try to parse as a HuggingFace URL/repo
	if repo, repoType, err := hf.ParseHuggingFaceRepo(path); err == nil {
		return true, repo, repoType
	}

	return false, "", hf.RepoTypeUnknown
}

func isLocalPath(path string) bool {
	// Check for common local path patterns
	if path == "." || path == ".." {
		return true
	}
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return true
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return true
	}
	if strings.HasPrefix(path, "/") {
		return true
	}
	// Windows-style paths
	if strings.HasPrefix(path, "\\") || (len(path) >= 2 && path[1] == ':') {
		return true
	}
	return false
}
