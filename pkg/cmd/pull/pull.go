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

package pull

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func runPull(ctx context.Context, opts *pullOptions) (ocispec.Descriptor, error) {
	storageHome := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageHome, opts.modelRef)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	return runPullRecursive(ctx, localRepo, opts, []string{})
}

func runPullRecursive(ctx context.Context, localRepo local.LocalRepo, opts *pullOptions, pulledRefs []string) (ocispec.Descriptor, error) {
	refStr := util.FormatRepositoryForDisplay(opts.modelRef.String())
	if idx := getIndex(pulledRefs, refStr); idx != -1 {
		cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(pulledRefs[idx:], "=>"), refStr)
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("found cycle in modelkit references: %s", cycleStr)
	}
	pulledRefs = append(pulledRefs, refStr)
	if len(pulledRefs) > constants.MaxModelRefChain {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("reached maximum number of model references: [%s]", strings.Join(pulledRefs, "=>"))
	}

	desc, err := pullModel(ctx, localRepo, opts)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}

	if err := pullParents(ctx, localRepo, desc, opts, pulledRefs); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to pull referenced modelkits: %w", err)
	}

	return desc, nil
}

func pullParents(ctx context.Context, localRepo local.LocalRepo, desc ocispec.Descriptor, optsIn *pullOptions, pulledRefs []string) error {
	_, config, err := util.GetManifestAndConfig(ctx, localRepo, desc)
	if err != nil {
		return err
	}
	if config.Model == nil || !util.IsModelKitReference(config.Model.Path) {
		return nil
	}
	output.Infof("Pulling referenced image %s", config.Model.Path)
	parentRef, _, err := util.ParseReference(config.Model.Path)
	if err != nil {
		return err
	}
	opts := *optsIn
	opts.modelRef = parentRef
	_, err = runPullRecursive(ctx, localRepo, &opts, pulledRefs)
	return err
}

func pullModel(ctx context.Context, localRepo local.LocalRepo, opts *pullOptions) (ocispec.Descriptor, error) {
	repo, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to read repository: %w", err)
	}
	if err := referenceIsModel(ctx, opts.modelRef, repo); err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}

	desc, err := localRepo.PullModel(ctx, repo, *opts.modelRef, &opts.NetworkOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to pull: %w", err)
	}

	return desc, nil
}

func referenceIsModel(ctx context.Context, ref *registry.Reference, repo registry.Repository) error {
	desc, rc, err := repo.FetchReference(ctx, ref.Reference)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", ref.String(), err)
	}
	defer rc.Close()

	if desc.MediaType != ocispec.MediaTypeImageManifest {
		return fmt.Errorf("reference %s is not an image manifest", ref.String())
	}
	manifestBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	manifest := &ocispec.Manifest{}
	if err := json.Unmarshal(manifestBytes, manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}
	if manifest.Config.MediaType != mediatype.ModelConfigMediaType.String() {
		return fmt.Errorf("reference %s does not refer to a model", ref.String())
	}
	return nil
}

func getIndex(list []string, s string) int {
	for idx, item := range list {
		if s == item {
			return idx
		}
	}
	return -1
}
