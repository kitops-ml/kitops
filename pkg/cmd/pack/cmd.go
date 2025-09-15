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

package pack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Pack a modelkit`
	longDesc  = `Pack a modelkit from a kitfile using the given context directory.

The packing process involves taking the configuration and resources defined in
your kitfile and using them to create a modelkit. This modelkit is then stored
in your local registry, making it readily available for further actions such
as pushing to a remote registry for collaboration.

Unless a different location is specified, this command looks for the kitfile
at the root of the provided context directory. Any relative paths defined
within the kitfile are interpreted as being relative to this context
directory.`

	examples = `# Pack a modelkit using the kitfile in the current directory
kit pack .

# Pack a modelkit with a specific kitfile and tag
kit pack . -f /path/to/your/Kitfile -t registry/repository:modelv1`
)

type packOptions struct {
	modelFile   string
	contextDir  string
	configHome  string
	storageHome string
	fullTagRef  string
	compression string
	modelRef    *registry.Reference
	extraRefs   []string
}

func PackCommand() *cobra.Command {
	opts := &packOptions{}

	cmd := &cobra.Command{
		Use:     "pack [flags] DIRECTORY",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		RunE:    runCommand(opts),
	}
	cmd.Flags().StringVarP(&opts.modelFile, "file", "f", "", "Specifies the path to the Kitfile explicitly (use \"-\" to read from standard input)")
	cmd.Flags().StringVarP(&opts.fullTagRef, "tag", "t", "", "Assigns one or more tags to the built modelkit. Example: -t registry/repository:tag1,tag2")
	cmd.Flags().StringVar(&opts.compression, "compression", "none", "Compression format to use for layers. Valid options: 'none' (default), 'gzip', 'gzip-fastest'")
	cmd.Flags().SortFlags = false
	cmd.Args = cobra.ExactArgs(1)
	cmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveDefault)
	return cmd
}

func runCommand(opts *packOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := opts.complete(cmd.Context(), args)
		if err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		// Change working directory to context path to make sure relative paths within
		// tarballs are correct. This is the equivalent of using the -C parameter for tar
		if err := os.Chdir(opts.contextDir); err != nil {
			return output.Fatalf("Failed to use context path %s: %s", opts.contextDir, err)
		}

		err = runPack(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to pack model kit: %s", err)
		}
		return nil
	}
}

func (opts *packOptions) complete(ctx context.Context, args []string) error {
	contextDir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("failed to get context dir %s: %w", args[0], err)
	}
	opts.contextDir = contextDir

	if opts.modelFile == "" {
		foundModel, err := filesystem.FindKitfileInPath(opts.contextDir)
		if err != nil {
			return err
		}
		opts.modelFile = foundModel
	}

	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.storageHome = constants.StoragePath(opts.configHome)

	if opts.fullTagRef != "" {
		modelRef, extraRefs, err := util.ParseReference(opts.fullTagRef)
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}
		opts.modelRef = modelRef
		opts.extraRefs = extraRefs
	} else {
		opts.modelRef = util.DefaultReference()
	}

	if err := constants.IsValidCompression(opts.compression); err != nil {
		return err
	}

	printConfig(opts)
	return nil
}

func printConfig(opts *packOptions) {
	output.Debugf("Using storage path: %s", opts.storageHome)
	output.Debugf("Context dir: %s", opts.contextDir)
	output.Debugf("Model file: %s", opts.modelFile)
	if opts.modelRef != nil {
		output.Debugf("Packing %s", opts.modelRef.String())
	} else {
		output.Debugln("No tag or reference specified")
	}
	if len(opts.extraRefs) > 0 {
		output.Debugf("Additional tags: %s", strings.Join(opts.extraRefs, ", "))
	}
}
