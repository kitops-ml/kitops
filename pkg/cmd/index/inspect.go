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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

const (
	inspectShortDesc = `Inspect a ModelKit index's manifest`
	inspectLongDesc  = `Print the contents of a ModelKit index manifest.

By default, kit will check local storage for the specified index. To
inspect an index stored on a remote registry, use the --remote flag.`

	inspectExample = `# Inspect a local index:
kit index inspect my-org/my-model:all

# Inspect a local index by digest:
kit index inspect my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Inspect a remote index:
kit index inspect --remote registry.example.com/my-org/my-model:all`
)

type inspectOptions struct {
	options.NetworkOptions
	configHome  string
	checkRemote bool
	indexRef    *registry.Reference
}

func indexInspectCommand() *cobra.Command {
	opts := &inspectOptions{}

	cmd := &cobra.Command{
		Use:     "inspect [flags] INDEX",
		Short:   inspectShortDesc,
		Long:    inspectLongDesc,
		Example: inspectExample,
		RunE:    runInspectCommand(opts),
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

func runInspectCommand(opts *inspectOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runInspect(cmd.Context(), opts); err != nil {
			return output.Fatalf("Failed to inspect index: %s", err)
		}
		return nil
	}
}

func runInspect(ctx context.Context, opts *inspectOptions) error {
	var store oras.Target
	if opts.checkRemote {
		remoteRepo, err := remote.NewRepository(ctx, opts.indexRef.Registry, opts.indexRef.Repository, &opts.NetworkOptions)
		if err != nil {
			return err
		}
		store = remoteRepo
	} else {
		repo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
		if err != nil {
			return err
		}
		store = repo
	}

	_, indexBytes, err := fetchIndexBytes(ctx, store, opts.indexRef.Reference)
	if err != nil {
		return err
	}
	if _, err := libindex.ParseIndex(indexBytes); err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, indexBytes, "", "  "); err != nil {
		return fmt.Errorf("failed to format index: %w", err)
	}
	output.Infoln(pretty.String())
	return nil
}

func (opts *inspectOptions) complete(ctx context.Context, args []string) error {
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
