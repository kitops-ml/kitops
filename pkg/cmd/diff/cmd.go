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

package diff

import (
	"context"
	"fmt"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/kit"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/output"
)

const (
	// Constants for identifying reference types
	remotePrefix = "remote://"
	localPrefix  = "local://"
	//Constants for formatting output tables.
	layerTableHeadings = "Type    | Digest             | Size"
	layerTableFormat   = "%-7s | %-18s | %s\n"
	shortDesc          = "Compare two ModelKits"
	longDesc           = `Compare two ModelKits to see the differences in their layers.
		
ModelKits can be specified from either a local or from a remote registry.
To specify a local ModelKit, prefix the reference with 'local://', e.g. 'local://jozu.ml/foo/bar'.
To specify a remote ModelKit, prefix the reference with 'remote://', e.g. 'remote://jozu.ml/foo/bar'.
If no prefix is specified, the local registry will be checked first.
`
	examples = `# Compare two ModelKits
kit diff jozu.ml/foo:latest jozu.ml/bar:latest

# Compare two ModelKits from a remote registry
kit diff remote://jozu.ml/foo:champion remote://jozu.ml/bar:latest

# Compare local ModelKit with a remote ModelKit
kit diff local://jozu.ml/foo:latest remote://jozu.ml/foo:latest
`
)

type diffOptions struct {
	kit.DiffOptions
	argA string
	argB string
}

func DiffCommand() *cobra.Command {
	opts := &diffOptions{}
	cmd := &cobra.Command{
		Use:     "diff <ModelKit1> <ModelKit2>",
		Short:   shortDesc,
		Args:    cobra.ExactArgs(2),
		Long:    longDesc,
		Example: examples,
		RunE:    runCommand(opts),
	}
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false
	return cmd
}

func runCommand(opts *diffOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		result, err := kit.Diff(cmd.Context(), &opts.DiffOptions)
		if err != nil {
			return output.Fatalf("%s", err)
		}

		if result.Identical {
			output.Infoln("ModelKits are identical")
			return nil
		}

		output.Infoln("Comparing:")
		output.Infof("  ModelKit1: %s\n", opts.RefA.String())
		output.Infof("  ModelKit2: %s\n\n", opts.RefB.String())

		output.Infoln("Configurations:")
		output.Infoln("---------------------------------------")
		if result.SameConfig {
			output.Infof("  Configs are identical\n\n")
		} else {
			output.Infof("Configs differ\n\n")
		}

		output.Infoln("Annotations:")
		output.Infoln("---------------------------------------")
		if result.AnnotationsMatch {
			output.Infof("  Annotations are identical \n\n")
		} else {
			output.Infof("  Annotations does not match\n\n")
		}

		displayLayers("Shared Layers", result.SharedLayers)
		displayLayers(fmt.Sprintf("Unique Layers to ModelKit1 (%s)", opts.RefA.String()), result.UniqueLayersA)
		displayLayers(fmt.Sprintf("Unique Layers to ModelKit2 (%s)", opts.RefB.String()), result.UniqueLayersB)
		return nil
	}
}

func (opts *diffOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.ConfigHome = configHome

	opts.argA = args[0]
	opts.argB = args[1]
	opts.LocalityA = localityFromPrefix(args[0])
	opts.LocalityB = localityFromPrefix(args[1])

	refA, _, err := artifact.ParseReference(removePrefix(args[0]))
	if err != nil {
		return fmt.Errorf("failed to parse reference for ref1: %w", err)
	}
	opts.RefA = refA

	refB, _, err := artifact.ParseReference(removePrefix(args[1]))
	if err != nil {
		return fmt.Errorf("failed to parse reference for ref2: %w", err)
	}
	opts.RefB = refB

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}
	return nil
}

func localityFromPrefix(arg string) string {
	if strings.HasPrefix(arg, remotePrefix) {
		return "remote"
	}
	if strings.HasPrefix(arg, localPrefix) {
		return "local"
	}
	return ""
}

func removePrefix(arg string) string {
	arg = strings.TrimPrefix(arg, remotePrefix)
	arg = strings.TrimPrefix(arg, localPrefix)
	return arg
}

func displayLayers(title string, layers []ocispec.Descriptor) {
	output.Infoln(title)
	output.Infoln("---------------------------------------")
	if len(layers) > 0 {
		output.Infof(layerTableHeadings)
		for _, layer := range layers {
			output.Infof(layerTableFormat,
				mediatype.FormatMediaTypeForUser(layer.MediaType),
				layer.Digest[:17],
				output.FormatBytes(layer.Size))
		}
	} else {
		output.Infoln("<none>")
	}
	output.Infoln("")
}
