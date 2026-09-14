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
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	createShortDesc = `Create a ModelKit index`
	createLongDesc  = `Create an empty ModelKit index in local storage.`

	createExample = `# Create an index in local storage
kit index create my-org/my-model:all

# Create an index using a full reference
kit index create registry.example.com/my-org/my-model:all`
)

type createOptions struct {
	configHome string
	indexRef   *registry.Reference
}

func indexCreateCommand() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:     "create [flags] INDEX",
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: createExample,
		RunE:    runCreateCommand(opts),
		Args:    cobra.ExactArgs(1),
	}
	cmd.Flags().SortFlags = false

	return cmd
}

func runCreateCommand(opts *createOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runCreate(cmd.Context(), opts); err != nil {
			return output.Fatalf("Failed to create index: %s", err)
		}
		return nil
	}
}

func runCreate(ctx context.Context, opts *createOptions) error {
	repo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
	if err != nil {
		return err
	}

	existing, err := repo.Resolve(ctx, opts.indexRef.Reference)
	if err == nil && existing.MediaType == ocispec.MediaTypeImageIndex {
		return fmt.Errorf("index %s already exists", displayRef(opts.indexRef))
	} else if err != nil && !errors.Is(err, errdef.ErrNotFound) {
		return err
	}

	desc, err := writeIndex(ctx, repo, libindex.CreateIndex(nil), opts.indexRef, ocispec.Descriptor{})
	if err != nil {
		return err
	}
	output.Infof("Created index %s (digest %s)", displayRef(opts.indexRef), desc.Digest)
	return nil
}

func (opts *createOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	indexRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if artifact.ReferenceIsDigest(indexRef.Reference) {
		return fmt.Errorf("index reference must be a tag, not a digest: %s", args[0])
	}
	if indexRef.Reference == "" {
		output.Infof("No tag specified for index. Using 'latest' as default ('%s:latest')", args[0])
		indexRef.Reference = "latest"
	}
	opts.indexRef = indexRef

	return nil
}
