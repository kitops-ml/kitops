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

package unpack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/unpack"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Produce the components from a modelkit on the local filesystem`
	longDesc  = `Produces all or selected components of a modelkit on the local filesystem.

This command unpacks a modelkit's components, including models, code,
datasets, and configuration files, to a specified directory on the local
filesystem. By default, it attempts to find the modelkit in local storage; if
not found, it searches the remote registry and retrieves it. This process
ensures that the necessary components are always available for unpacking,
optimizing for efficiency by fetching only specified components from the
remote registry when necessary.

The content that is unpacked can be limited via the --filter (-f) flag. For example,
use
    --filter=model
to unpack only the model, or
    --filter=datasets:my-dataset
to unpack only the dataset named 'my-dataset'.

Valid filters have the format
    [types]:[filters]
where [types] is a comma-separated list of Kitfile fields (kitfile, model, datasets
code, or docs) and [filters] is an optional comma-separated list of additional filters
to apply, which are matched against the Kitfile to further restrict what is extracted.
Additional filters match elements of the Kitfile on either the name (if present) or
the path used.

The filter field can be specified multiple times. A layer will be unpacked if it matches
any of the specified filters`

	example = `# Unpack all components of a modelkit to the current directory
kit unpack myrepo/my-model:latest -d /path/to/unpacked

# Unpack only the model and datasets of a modelkit to a specified directory
kit unpack myrepo/my-model:latest --filter=model,datasets -d /path/to/unpacked

# Unpack only the dataset named "my-dataset" to the current directory
kit unpack myrepo/my-model:latest --filter=datasets:my-dataset

# Unpack only the docs layer with path "./README.md" to the current directory
kit unpack myrepo/my-model:latest --filter=docs:./README.md

# Unpack the model and the dataset named "validation"
kit unpack myrepo/my-model:latest --filter=model --filter=datasets:validation

# Unpack a modelkit from a remote registry with overwrite enabled
kit unpack registry.example.com/myrepo/my-model:latest -o -d /path/to/unpacked`
)

type unpackOptions struct {
	options.NetworkOptions
	configHome     string
	unpackDir      string
	filters        []string
	unpackConf     unpackConf
	modelRef       *registry.Reference
	overwrite      bool
	ignoreExisting bool
}

// unpackConf configures which elements of the modelkit should be unpacked.
// Deprecated: use filterConf instead, which supports advanced filtering
type unpackConf struct {
	unpackKitfile  bool
	unpackModels   bool
	unpackCode     bool
	unpackDatasets bool
	unpackDocs     bool
}

func (opts *unpackOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	modelRef, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	if modelRef.Reference == "" {
		output.Infof("No tag specified for unpack. Using 'latest' as default ('%s:latest')", args[0])
		modelRef.Reference = "latest"
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("can not unpack multiple tags")
	}
	opts.modelRef = modelRef

	absDir, err := filepath.Abs(opts.unpackDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path %s: %w", opts.unpackDir, err)
	}
	opts.unpackDir = absDir

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	printConfig(opts)
	return nil
}

func UnpackCommand() *cobra.Command {
	opts := &unpackOptions{}

	cmd := &cobra.Command{
		Use:     "unpack [flags] [registry/]repository[:tag|@digest]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	cmd.Args = cobra.ExactArgs(1)
	cmd.Flags().StringVarP(&opts.unpackDir, "dir", "d", "", "The target directory to unpack components into. This directory will be created if it does not exist")
	cmd.Flags().BoolVarP(&opts.overwrite, "overwrite", "o", false, "Overwrites existing files and directories in the target unpack directory without prompting")
	cmd.Flags().BoolVarP(&opts.ignoreExisting, "ignore-existing", "i", false, "Skip unpacking files if a file with that name already exists")
	cmd.Flags().StringArrayVarP(&opts.filters, "filter", "f", []string{}, "Filter what is unpacked from the modelkit based on type and name. Can be specified multiple times")
	cmd.Flags().BoolVar(&opts.unpackConf.unpackKitfile, "kitfile", false, "Unpack only Kitfile (deprecated: use --filter=kitfile)")
	cmd.Flags().BoolVar(&opts.unpackConf.unpackModels, "model", false, "Unpack only model (deprecated: use --filter=model)")
	cmd.Flags().BoolVar(&opts.unpackConf.unpackCode, "code", false, "Unpack only code (deprecated: use --filter=code)")
	cmd.Flags().BoolVar(&opts.unpackConf.unpackDatasets, "datasets", false, "Unpack only datasets (deprecated: use --filter=datasets)")
	cmd.Flags().BoolVar(&opts.unpackConf.unpackDocs, "docs", false, "Unpack only docs (deprecated: use --filter=docs)")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	cmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveDefault)
	return cmd
}

func runCommand(opts *unpackOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		if opts.modelRef.Reference == "" {
			return output.Fatalf("Invalid reference: unpacking requires a tag or digest")
		}

		unpackTo := opts.unpackDir
		if unpackTo == "" {
			unpackTo = "current directory"
		}
		// Make sure target directory exists, in case user is using the -d flag
		if err := os.MkdirAll(opts.unpackDir, 0755); err != nil {
			return output.Fatalf("failed to create directory %s: %w", opts.unpackDir, err)
		}
		output.Infof("Unpacking to %s", unpackTo)

		// Convert command options to library options
		libOpts := &unpack.UnpackOptions{
			ModelRef:       opts.modelRef,
			UnpackDir:      opts.unpackDir,
			ConfigHome:     opts.configHome,
			Filters:        opts.filters,
			Overwrite:      opts.overwrite,
			IgnoreExisting: opts.ignoreExisting,
			NetworkOptions: opts.NetworkOptions,
		}

		// Handle deprecated flags by converting to filters
		conf := opts.unpackConf
		if conf.unpackKitfile || conf.unpackModels || conf.unpackCode || conf.unpackDatasets || conf.unpackDocs {
			deprecatedFilters := unpack.FiltersFromUnpackConf(
				conf.unpackKitfile, conf.unpackModels, conf.unpackCode,
				conf.unpackDatasets, conf.unpackDocs)
			libOpts.FilterConfs = deprecatedFilters
		} else if len(opts.filters) > 0 {
			// Parse filters using library functionality
			for _, filter := range opts.filters {
				filterConf, err := unpack.ParseFilter(filter)
				if err != nil {
					return output.Fatalf("Invalid filter %q: %s", filter, err)
				}
				libOpts.FilterConfs = append(libOpts.FilterConfs, *filterConf)
			}
		}

		err := unpack.UnpackModelKit(cmd.Context(), libOpts)
		if err != nil {
			return output.Fatalln(err)
		}
		return nil
	}
}

func printConfig(opts *unpackOptions) {
	output.Debugf("Overwrite: %t", opts.overwrite)
	output.Debugf("Unpacking %s", opts.modelRef.String())
}
