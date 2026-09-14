// Copyright 2026 The KitOps Authors.
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

package index

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	infoShortDesc = `Show the ModelKits contained in an index`
	infoLongDesc  = `List the ModelKits in a ModelKit index.`

	infoExample = `# See the contents of a local index:
kit index info my-org/my-model:all

# See the contents of a local index by digest:
kit index info my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# See the contents of a remote index:
kit index info --remote registry.example.com/my-org/my-model:all`
)

type infoOptions struct {
	options.NetworkOptions
	configHome  string
	checkRemote bool
	indexRef    *registry.Reference
}

func indexInfoCommand() *cobra.Command {
	opts := &infoOptions{}

	cmd := &cobra.Command{
		Use:     "info [flags] INDEX",
		Short:   infoShortDesc,
		Long:    infoLongDesc,
		Example: infoExample,
		RunE:    runInfoCommand(opts),
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalIndexesCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	cmd.Flags().BoolVarP(&opts.checkRemote, "remote", "r", false, "Check remote registry instead of local storage")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runInfoCommand(opts *infoOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runInfo(cmd.Context(), cmd.OutOrStdout(), opts); err != nil {
			return output.Fatalf("Failed to read index: %s", err)
		}
		return nil
	}
}

func runInfo(ctx context.Context, out io.Writer, opts *infoOptions) error {
	var store oras.Target
	var localRepo local.LocalRepo
	if opts.checkRemote {
		remoteRepo, err := remote.NewRepository(ctx, opts.indexRef.Registry, opts.indexRef.Repository, &opts.NetworkOptions)
		if err != nil {
			return err
		}
		store = remoteRepo
	} else {
		indexRepo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
		if err != nil {
			return err
		}
		modelKitRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.indexRef)
		if err != nil {
			return err
		}
		store, localRepo = indexRepo, modelKitRepo
	}

	desc, idx, err := resolveIndex(ctx, store, opts.indexRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) && !opts.checkRemote {
			return fmt.Errorf("could not find index %s in local storage; use 'kit index pull' to download it, or --remote to read it from its registry", displayRef(opts.indexRef))
		}
		return err
	}
	printIndexInfo(ctx, out, desc, idx, localRepo, opts.indexRef)
	return nil
}

func printIndexInfo(ctx context.Context, out io.Writer, desc ocispec.Descriptor, idx *libindex.ModelKitIndex, localRepo local.LocalRepo, ref *registry.Reference) {
	fmt.Fprintf(out, "Index %s (%s)\n", displayRef(ref), desc.Digest)
	if len(idx.Manifests) == 0 {
		fmt.Fprintln(out, "Index contains no ModelKits")
		return
	}
	if len(idx.Manifests) == 1 {
		fmt.Fprintln(out, "1 ModelKit:")
	} else {
		fmt.Fprintf(out, "%d ModelKits:\n", len(idx.Manifests))
	}

	tw := tabwriter.NewWriter(out, 0, 2, 4, ' ', 0)
	for _, entry := range idx.Manifests {
		status := ""
		if localRepo != nil {
			if exists, err := localRepo.Exists(ctx, entry.Descriptor); err == nil && !exists {
				status = "missing from local storage"
			}
		}
		fmt.Fprintf(tw, "  %s\t%s\t%s\n", entry.Digest, output.FormatBytes(entry.Size), status)
	}
	tw.Flush()

	for _, entry := range idx.Manifests {
		if len(entry.ModelMeta) == 0 {
			continue
		}
		fmt.Fprintf(out, "\n  %s\n", entry.Digest)
		printModelMetadata(out, entry.ModelMeta)
	}
}

func printModelMetadata(out io.Writer, meta libindex.ModelMetadata) {
	tw := tabwriter.NewWriter(out, 0, 2, 4, ' ', 0)
	for _, key := range slices.Sorted(maps.Keys(meta)) {
		fmt.Fprintf(tw, "      %s:\t%s\n", key, formatMetadataValue(meta[key]))
	}
	tw.Flush()
}

// formatMetadataValue prints string values unquoted, leaving other JSON as-is.
func formatMetadataValue(value json.RawMessage) string {
	var str string
	if err := json.Unmarshal(value, &str); err == nil {
		return str
	}
	return string(value)
}

func (opts *infoOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	indexRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return err
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if indexRef.Reference == "" {
		return fmt.Errorf("missing tag or digest from index reference '%s'", args[0])
	}
	opts.indexRef = indexRef

	if opts.indexRef.Registry == artifact.DefaultRegistry && opts.checkRemote {
		return fmt.Errorf("can not check remote: %s does not contain registry", artifact.FormatRepositoryForDisplay(opts.indexRef.String()))
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}
