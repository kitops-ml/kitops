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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	infoShortDesc = `Show the ModelKits contained in an index`
	infoLongDesc  = `List the ModelKits in a ModelKit index, with their labels and annotations.`

	infoExample = `# See the contents of a local index:
kit index info my-org/my-model:all

# See the contents of a local index by digest:
kit index info my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# See the contents of a remote index:
kit index info --remote registry.example.com/my-org/my-model:all`
)

const indexTableFmt = "%s\t%s\t%s\t%s\t%s\n"

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
	var indexStore, modelKitStore oras.Target
	if opts.checkRemote {
		remoteRepo, err := remote.NewRepository(ctx, opts.indexRef.Registry, opts.indexRef.Repository, &opts.NetworkOptions)
		if err != nil {
			return err
		}
		indexStore, modelKitStore = remoteRepo, remoteRepo
	} else {
		indexRepo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
		if err != nil {
			return err
		}
		modelKitRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.indexRef)
		if err != nil {
			return err
		}
		indexStore, modelKitStore = indexRepo, modelKitRepo
	}

	desc, idx, err := resolveIndex(ctx, indexStore, opts.indexRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) && !opts.checkRemote {
			return fmt.Errorf("could not find index %s in local storage; use 'kit index pull' to download it, or --remote to read it from its registry", displayRef(opts.indexRef))
		}
		return err
	}
	printIndexInfo(ctx, out, modelKitStore, desc, idx, opts.indexRef)
	return nil
}

func printIndexInfo(ctx context.Context, out io.Writer, store oras.ReadOnlyTarget, desc ocispec.Descriptor, idx *libindex.ModelKitIndex, ref *registry.Reference) {
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

	var table bytes.Buffer
	tw := tabwriter.NewWriter(&table, 0, 2, 3, ' ', 0)
	fmt.Fprintf(tw, indexTableFmt, "ORIGINAL REFERENCE", "SIZE", "LABELS", "ANNOTATIONS", "DIGEST")
	for _, entry := range idx.Manifests {
		labels := libindex.LabelPairs(entry.ModelMeta)
		annotations := libindex.KeyValuePairs(omitRecordedAnnotations(entry.Annotations))
		for row := range max(1, len(labels), len(annotations)) {
			reference, size, dgst := "", "", ""
			if row == 0 {
				reference, size, dgst = entryReference(entry, ref), entrySize(ctx, store, entry), entry.Digest.String()
			}
			fmt.Fprintf(tw, indexTableFmt, reference, size, cellAt(labels, row), cellAt(annotations, row), dgst)
		}
	}
	tw.Flush()

	// tabwriter aligns runs of lines holding the same number of cells, so every row carries
	// all five; trimming happens afterwards to keep continuation rows free of padding.
	for _, line := range strings.Split(strings.TrimRight(table.String(), "\n"), "\n") {
		fmt.Fprintln(out, strings.TrimRight(line, " "))
	}
}

func cellAt(values []string, row int) string {
	if row < len(values) {
		return values[row]
	}
	if row == 0 {
		return noneValue
	}
	return ""
}

// entrySize reports the total size of the layers of the ModelKit an entry refers to, matching
// what 'kit list' reports. It is read from the size recorded on the entry, or for entries
// without one, from the ModelKit itself; if neither is available, the size is unknown.
func entrySize(ctx context.Context, store oras.ReadOnlyTarget, entry libindex.ModelKitIndexDescriptor) string {
	if recorded, err := strconv.ParseInt(entry.Annotations[constants.ModelKitSizeAnnotation], 10, 64); err == nil {
		return output.FormatBytes(recorded)
	}
	manifest, err := util.GetManifest(ctx, store, entry.Descriptor)
	if err != nil {
		return noneValue
	}
	return output.FormatBytes(util.ModelKitSize(manifest))
}

func entryReference(entry libindex.ModelKitIndexDescriptor, indexRef *registry.Reference) string {
	if reference := libindex.EntryReference(entry, indexRef); reference != "" {
		return reference
	}
	return noneValue
}

// omitRecordedAnnotations drops the original tag and size annotations, which have their own
// columns.
func omitRecordedAnnotations(annotations map[string]string) map[string]string {
	_, hasTag := annotations[constants.OriginalTagAnnotation]
	_, hasSize := annotations[constants.ModelKitSizeAnnotation]
	if !hasTag && !hasSize {
		return annotations
	}
	remaining := maps.Clone(annotations)
	delete(remaining, constants.OriginalTagAnnotation)
	delete(remaining, constants.ModelKitSizeAnnotation)
	return remaining
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
