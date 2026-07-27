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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"

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
directory.

If --push is set along with a --tag pointing to a remote registry, the modelkit
is streamed directly to that registry instead of being stored locally.`

	examples = `# Pack a modelkit using the kitfile in the current directory
kit pack .

# Pack a modelkit with a specific kitfile and tag
kit pack . -f /path/to/your/Kitfile -t registry/repository:modelv1

# Pack a modelkit and push it directly to a remote registry without storing it locally
kit pack . -t registry.example.com/my-org/my-model:latest --push`
)

type packOptions struct {
	options.NetworkOptions
	modelFile    string
	contextDir   string
	configHome   string
	storageHome  string
	fullTagRef   string
	compression  string
	layerFormat  string
	modelRef     *registry.Reference
	extraRefs    []string
	useModelPack bool
	push         bool
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
	cmd.Flags().StringVar(&opts.compression, "compression", "none", "Compression format to use for layers. Valid options: 'none', 'gzip', 'gzip-fastest', 'zstd'")
	cmd.Flags().StringVar(&opts.layerFormat, "layer-format", "tar", "Packaging format to use for layers. Valid options: 'tar', 'raw'")
	cmd.Flags().BoolVar(&opts.useModelPack, "use-model-pack", false, "Pack model in ModelPack format instead of ModelKit")
	cmd.Flags().BoolVar(&opts.push, "push", false, "Stream the packed modelkit directly to the remote registry specified by --tag, without storing it locally")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false
	cmd.Args = cobra.ExactArgs(1)
	cmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveDefault)
	return cmd
}

func runCommand(opts *packOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := opts.complete(cmd, args)
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

func (opts *packOptions) complete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
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
		modelRef, extraRefs, err := artifact.ParseReference(opts.fullTagRef)
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}
		if modelRef.Reference == "" {
			output.Infof("No tag or digest specified with --tag flag. Using 'latest' as default ('%s:latest')", opts.fullTagRef)
			modelRef.Reference = "latest"
		}
		opts.modelRef = modelRef
		opts.extraRefs = extraRefs
	} else {
		opts.modelRef = artifact.DefaultReference()
	}

	if err := mediatype.IsValidCompression(opts.compression); err != nil {
		return err
	}

	if err := mediatype.IsValidLayerFormat(opts.layerFormat); err != nil {
		return err
	}

	if opts.push {
		if opts.modelRef.Registry == artifact.DefaultRegistry {
			return fmt.Errorf("--push requires a remote registry to be specified with --tag")
		}
		if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
			return err
		}
	} else if networkFlagsChanged(cmd) {
		output.Infof("Warning: network-related flags are only applicable with --push flag")
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

func networkFlagsChanged(cmd *cobra.Command) bool {
	flags := []string{"plain-http", "tls-verify", "tls-cert", "cert", "key", "concurrency", "proxy"}
	for _, f := range flags {
		if cmd.Flags().Changed(f) {
			return true
		}
	}
	return false
}

