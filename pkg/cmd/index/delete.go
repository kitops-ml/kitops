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
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	deleteShortDesc = `Delete a ModelKit index`
	deleteLongDesc  = `Delete a ModelKit index from local storage. The ModelKits it references are
not removed. To delete an index from a remote registry, use
'kit remove --remote'.

If specified by tag, the index is only untagged while other tags refer to it,
unless --force is used.`

	deleteExample = `# Delete an index from local storage
kit index delete my-org/my-model:all

# Delete an index and every other tag that refers to it
kit index delete --force my-org/my-model:all

# Delete an index by digest
kit index delete my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Delete all untagged indexes
kit index delete --all

# Delete all locally stored indexes
kit index delete --all --force`
)

type deleteOptions struct {
	configHome  string
	forceDelete bool
	deleteAll   bool
	indexRef    *registry.Reference
}

func indexDeleteCommand() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete [flags] INDEX",
		Short:   deleteShortDesc,
		Long:    deleteLongDesc,
		Example: deleteExample,
		RunE:    runDeleteCommand(opts),
	}
	cmd.Flags().BoolVarP(&opts.forceDelete, "force", "f", false, "Delete the index and all other tags that refer to it")
	cmd.Flags().BoolVarP(&opts.deleteAll, "all", "a", false, "Delete all untagged indexes")
	cmd.Flags().SortFlags = false

	cmd.Args = func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			if opts.deleteAll {
				return nil
			}
			return fmt.Errorf("index is required for delete unless --all is specified")
		case 1:
			if opts.deleteAll {
				return fmt.Errorf("index should not be specified when --all flag is used")
			}
			return nil
		default:
			return cobra.MaximumNArgs(1)(cmd, args)
		}
	}

	return cmd
}

func runDeleteCommand(opts *deleteOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		return output.Fatalf("Not implemented: kit index delete")
	}
}

func (opts *deleteOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	if opts.deleteAll {
		return nil
	}

	indexRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if indexRef.Reference == "" {
		return fmt.Errorf("tag or digest is required when deleting an index ('%s:<tag>')", args[0])
	}
	opts.indexRef = indexRef

	return nil
}
