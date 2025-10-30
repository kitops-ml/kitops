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

package kitimport

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/git"
	repoutils "github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Import a model from HuggingFace`
	longDesc  = `Download a repository from HuggingFace and package it as a ModelKit.

The repository can be specified either via a repository (e.g. myorg/myrepo) or
with a full URL (https://huggingface.co/myorg/myrepo). The repository will be
downloaded to a temporary directory and be packaged using a generated Kitfile.

In interactive settings, this command will read the EDITOR environment variable
to determine which editor should be used for editing the Kitfile.

This command supports multiple ways of downloading files from the remote
repository. The tool used can be specified using the --tool flag with one of the
options below:

  --tool=hf  : Download files using the Huggingface API. Requires REPOSITORY to
	             be a Huggingface repository. This is the default for Huggingface
							 repositories
  --tool=git : Download files using Git and Git LFS. Works for any Git
	             repository but requires that Git and Git LFS are installed.

By default, Kit will automatically select the tool based on the provided
REPOSITORY.`

	example = `# Download repository myorg/myrepo and package it, using the default tag (myorg/myrepo:latest)
kit import myorg/myrepo

# Download repository using a different version (tag) than the default and package it
kit import myorg/myrepo --ref v1.0.0

# Download repository and tag it 'myrepository:mytag'
kit import myorg/myrepo --tag myrepository:mytag

# Download repository and pack it using an existing Kitfile
kit import myorg/myrepo --file ./path/to/Kitfile`
)

type importOptions struct {
	configHome   string
	repo         string
	repoRef      string
	tag          string
	token        string
	kitfilePath  string
	downloadTool string
	concurrency  int
	modelKitRef  *registry.Reference
}

func ImportCommand() *cobra.Command {
	opts := &importOptions{}

	cmd := &cobra.Command{
		Use:     "import [flags] REPOSITORY",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&opts.repoRef, "ref", git.DefaultGitRef, "Version (tag) of repository to import")
	cmd.Flags().StringVar(&opts.token, "token", "", "Token to use for authenticating with repository")
	cmd.Flags().StringVarP(&opts.tag, "tag", "t", "", "Tag for the ModelKit (default is '[repository]:latest')")
	cmd.Flags().StringVarP(&opts.kitfilePath, "file", "f", "", "Path to Kitfile to use for packing (use '-' to read from standard input)")
	cmd.Flags().StringVar(&opts.downloadTool, "tool", "", "Tool to use for downloading files: options are 'git' and 'hf' (default: detect based on repository)")
	cmd.Flags().IntVar(&opts.concurrency, "concurrency", 5, "Maximum number of simultaneous downloads (for huggingface)")
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *importOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalln(err)
		}

		importer, err := getImporter(opts)
		if err != nil {
			return output.Fatalln(err)
		}
		if err := importer(cmd.Context(), opts); err != nil {
			return output.Fatalln(err)
		}

		return nil
	}
}

func (opts *importOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.repo = args[0]

	if opts.tag == "" {
		tag, err := extractRepoFromURL(opts.repo)
		if err != nil {
			output.Errorf("Could not generate tag from URL: %s", err)
			return fmt.Errorf("use flag --tag to set a tag for ModelKit")
		}
		tag = strings.ToLower(tag)
		opts.tag = fmt.Sprintf("%s:latest", tag)
		output.Infof("Using tag %s. Use flag --tag to override", opts.tag)
	}

	ref, _, err := repoutils.ParseReference(opts.tag)
	if err != nil {
		return fmt.Errorf("invalid argument: tag '%s' is invalid: %w", opts.tag, err)
	}
	if ref.Reference == "" {
		output.Infof("No tag or digest provided with --tag flag. Using 'latest' as default ('%s:latest')", opts.tag)
		ref.Reference = "latest"
	}
	opts.modelKitRef = ref

	validTools := []string{"git", "hf"}
	if opts.downloadTool != "" && !slices.Contains(validTools, opts.downloadTool) {
		return fmt.Errorf("invalid value for --tool flag. Valid options are: %s", strings.Join(validTools, ", "))
	}

	if opts.concurrency < 1 {
		return fmt.Errorf("invalid argument for concurrency (%d): must be at least 1", opts.concurrency)
	}
	return nil
}

func getImporter(opts *importOptions) (func(context.Context, *importOptions) error, error) {
	switch opts.downloadTool {
	case "hf":
		return importUsingHF, nil
	case "git":
		return importUsingGit, nil
	default:
		repoUrl, err := url.Parse(opts.repo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", opts.repo, err)
		}

		if repoUrl.Host == "" || strings.Contains(repoUrl.Host, "huggingface") {
			return importUsingHF, nil
		}

		return importUsingGit, nil
	}
}
