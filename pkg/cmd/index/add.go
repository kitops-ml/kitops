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
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	addShortDesc = `Add a ModelKit to a ModelKit index`
	addLongDesc  = `Add a ModelKit to a ModelKit index.`

	addExample = `# Add a ModelKit to an index
kit index add my-org/my-model:all my-org/my-model:q4_0

# Add a ModelKit to an index by digest
kit index add my-org/my-model:all my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`
)

type addOptions struct {
	options.NetworkOptions
	configHome  string
	checkRemote bool
	indexRef    *registry.Reference
	modelRef    *registry.Reference
}

func indexAddCommand() *cobra.Command {
	opts := &addOptions{}

	cmd := &cobra.Command{
		Use:     "add [flags] INDEX MODELKIT",
		Short:   addShortDesc,
		Long:    addLongDesc,
		Example: addExample,
		RunE:    runAddCommand(opts),
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

	cmd.Flags().BoolVarP(&opts.checkRemote, "remote", "r", false, "Resolve the ModelKit in its remote registry instead of local storage")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runAddCommand(opts *addOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runAdd(cmd.Context(), opts); err != nil {
			return output.Fatalf("Failed to add ModelKit to index: %s", err)
		}
		return nil
	}
}

func runAdd(ctx context.Context, opts *addOptions) error {
	repo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
	if err != nil {
		return err
	}

	var src oras.Target
	if opts.checkRemote {
		remoteRepo, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
		if err != nil {
			return err
		}
		src = remoteRepo
	} else {
		modelRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.modelRef)
		if err != nil {
			return err
		}
		src = modelRepo
	}
	modelDesc, _, err := util.ResolveManifest(ctx, src, opts.modelRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) && !opts.checkRemote {
			return fmt.Errorf("could not find ModelKit %s in local storage; use --remote to resolve it in its registry", displayRef(opts.modelRef))
		}
		return err
	}

	prevIndexDesc, idx, err := resolveIndex(ctx, repo, opts.indexRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return fmt.Errorf("could not find index %s in local storage; use 'kit index create' to create it", displayRef(opts.indexRef))
		}
		return fmt.Errorf("failed to read index %s: %w", displayRef(opts.indexRef), err)
	}

	idx.AddEntry(libindex.ModelKitIndexDescriptor{Descriptor: modelDesc})
	desc, err := writeIndex(ctx, repo, idx, opts.indexRef, prevIndexDesc)
	if err != nil {
		return err
	}
	output.Infof("Added %s to index %s (digest %s)", displayRef(opts.modelRef), displayRef(opts.indexRef), desc.Digest)
	return nil
}

func (opts *addOptions) complete(ctx context.Context, args []string) error {
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
	if opts.modelRef.Registry == artifact.DefaultRegistry && opts.checkRemote {
		return fmt.Errorf("can not check remote: %s does not contain registry", artifact.FormatRepositoryForDisplay(opts.modelRef.String()))
	}

	if opts.indexRef.Registry != opts.modelRef.Registry || opts.indexRef.Repository != opts.modelRef.Repository {
		return fmt.Errorf("ModelKit %s is not in the same repository as index %s; an index may only reference ModelKits in its own repository",
			displayRef(opts.modelRef), displayRef(opts.indexRef))
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}
