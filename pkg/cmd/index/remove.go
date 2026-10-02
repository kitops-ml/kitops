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
	"errors"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	removeShortDesc = `Remove a ModelKit from a ModelKit index`
	removeLongDesc  = `Remove a ModelKit from a ModelKit index in local storage. The ModelKit itself
is not removed.`

	removeExample = `# Remove a ModelKit from an index
kit index remove my-org/my-model:all my-org/my-model:q4_0

# Remove a ModelKit from an index by digest
kit index remove my-org/my-model:all sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`
)

type removeOptions struct {
	configHome string
	indexRef   *registry.Reference
	modelRef   *registry.Reference
}

func indexRemoveCommand() *cobra.Command {
	opts := &removeOptions{}

	cmd := &cobra.Command{
		Use:     "remove [flags] INDEX MODELKIT",
		Short:   removeShortDesc,
		Long:    removeLongDesc,
		Example: removeExample,
		RunE:    runRemoveCommand(opts),
		Args:    cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completion.GetLocalIndexesCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			case 1:
				return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmd.Flags().SortFlags = false

	return cmd
}

func runRemoveCommand(opts *removeOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runRemove(cmd.Context(), opts); err != nil {
			return output.Fatalf("Failed to remove ModelKit from index: %s", err)
		}
		return nil
	}
}

func runRemove(ctx context.Context, opts *removeOptions) error {
	repo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
	if err != nil {
		return err
	}

	prevIndexDesc, idx, err := resolveIndex(ctx, repo, opts.indexRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return fmt.Errorf("could not find index %s in local storage", displayRef(opts.indexRef))
		}
		return fmt.Errorf("failed to read index %s: %w", displayRef(opts.indexRef), err)
	}

	entryDigest, err := resolveEntryDigest(ctx, opts)
	if err != nil {
		return err
	}
	if !idx.RemoveEntry(entryDigest) {
		return fmt.Errorf("index %s does not contain %s", displayRef(opts.indexRef), entryDigest)
	}

	desc, err := writeIndex(ctx, repo, idx, opts.indexRef, prevIndexDesc)
	if err != nil {
		return err
	}
	removed := entryDigest.String()
	if !artifact.ReferenceIsDigest(opts.modelRef.Reference) {
		removed = displayRef(opts.modelRef)
	}
	output.Infof("Removed %s from index %s (digest %s)", removed, displayRef(opts.indexRef), desc.Digest)
	return nil
}

// resolveEntryDigest returns the digest of the entry to remove. A tag is resolved in the
// repository it names, since an entry may refer to a ModelKit that is not in local storage.
func resolveEntryDigest(ctx context.Context, opts *removeOptions) (digest.Digest, error) {
	if artifact.ReferenceIsDigest(opts.modelRef.Reference) {
		return digest.Parse(opts.modelRef.Reference)
	}
	modelRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.modelRef)
	if err != nil {
		return "", err
	}
	desc, err := modelRepo.Resolve(ctx, opts.modelRef.Reference)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s in local storage: %w; specify the entry by digest instead", displayRef(opts.modelRef), err)
	}
	return desc.Digest, nil
}

func (opts *removeOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	indexRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse index reference %s: %w", args[0], err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid index reference: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if artifact.ReferenceIsDigest(indexRef.Reference) {
		return fmt.Errorf("index reference must be a tag, not a digest: %s", args[0])
	}
	if indexRef.Reference == "" {
		output.Infof("No tag specified for index. Using 'latest' as default ('%s:latest')", args[0])
		indexRef.Reference = "latest"
	}
	opts.indexRef = indexRef

	modelRef, extraTags, err := artifact.ParseReference(args[1])
	if err != nil {
		return fmt.Errorf("failed to parse ModelKit reference %s: %w", args[1], err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid ModelKit reference: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if modelRef.Reference == "" {
		return fmt.Errorf("missing tag or digest from ModelKit reference '%s'", args[1])
	}
	opts.modelRef = modelRef

	return nil
}
