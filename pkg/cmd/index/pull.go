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
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	pullShortDesc = `Download a ModelKit index from a registry`
	pullLongDesc  = `Pull a ModelKit index from a remote registry. Only the index is pulled; use
'kit pull -l' or 'kit unpack -l' to pull one of its ModelKits.`

	pullExample = `# Pull an index from a remote registry
kit index pull registry.example.com/my-org/my-model:all`
)

type pullOptions struct {
	options.NetworkOptions
	configHome string
	indexRef   *registry.Reference
}

func indexPullCommand() *cobra.Command {
	opts := &pullOptions{}

	cmd := &cobra.Command{
		Use:     "pull [flags] INDEX",
		Short:   pullShortDesc,
		Long:    pullLongDesc,
		Example: pullExample,
		RunE:    runPullCommand(opts),
		Args:    cobra.ExactArgs(1),
	}

	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runPullCommand(opts *pullOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		if err := runPull(cmd.Context(), opts); err != nil {
			return output.Fatalf("Failed to pull index: %s", err)
		}
		return nil
	}
}

func runPull(ctx context.Context, opts *pullOptions) error {
	remoteRepo, err := remote.NewRepository(ctx, opts.indexRef.Registry, opts.indexRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return err
	}

	output.Infof("Pulling index %s", opts.indexRef.String())
	desc, indexBytes, err := fetchIndexBytes(ctx, remoteRepo, opts.indexRef.Reference)
	if err != nil {
		return err
	}
	if _, err := libindex.ParseIndex(indexBytes); err != nil {
		return err
	}

	repo, err := local.NewLocalIndexRepo(constants.StoragePath(opts.configHome), opts.indexRef)
	if err != nil {
		return err
	}

	if _, err := writeIndexBytes(ctx, repo, desc, indexBytes, opts.indexRef, ocispec.Descriptor{}); err != nil {
		return err
	}
	output.Infof("Pulled index %s", desc.Digest)
	return nil
}

func (opts *pullOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	indexRef, extraTags, err := artifact.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference %s: %w", args[0], err)
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if artifact.ReferenceIsDigest(indexRef.Reference) {
		return fmt.Errorf("index reference must be a tag, not a digest: %s", args[0])
	}
	if indexRef.Reference == "" {
		output.Infof("No tag specified for pull. Using 'latest' as default ('%s:latest')", args[0])
		indexRef.Reference = "latest"
	}
	opts.indexRef = indexRef

	if opts.indexRef.Registry == artifact.DefaultRegistry {
		return fmt.Errorf("registry is required when pulling")
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}
