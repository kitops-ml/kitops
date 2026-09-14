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
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
)

const noneValue = "<none>"

// resolveIndex resolves a reference and reads the ModelKit index it describes.
func resolveIndex(ctx context.Context, store oras.Target, reference string) (ocispec.Descriptor, *libindex.ModelKitIndex, error) {
	desc, indexBytes, err := fetchIndexBytes(ctx, store, reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}
	idx, err := libindex.ParseIndex(indexBytes)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}
	return desc, idx, nil
}

func fetchIndexBytes(ctx context.Context, store oras.Target, reference string) (ocispec.Descriptor, []byte, error) {
	desc, err := store.Resolve(ctx, reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, err
	}
	if desc.MediaType != ocispec.MediaTypeImageIndex {
		return ocispec.DescriptorEmptyJSON, nil, libindex.ErrNotAKitIndex
	}
	indexBytes, err := content.FetchAll(ctx, store, desc)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to read index: %w", err)
	}
	return desc, indexBytes, nil
}

// writeIndex saves an updated index, moves the reference's tag to it, and removes the
// previous version of the index if nothing else refers to it.
func writeIndex(ctx context.Context, repo local.LocalRepo, idx *libindex.ModelKitIndex, ref *registry.Reference, prevIndexDesc ocispec.Descriptor) (ocispec.Descriptor, error) {
	desc, indexBytes, err := idx.Marshal()
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	return writeIndexBytes(ctx, repo, desc, indexBytes, ref, prevIndexDesc)
}

// writeIndexBytes saves index content as it was given, so that content from elsewhere keeps
// the digest it already has.
func writeIndexBytes(ctx context.Context, repo local.LocalRepo, desc ocispec.Descriptor, indexBytes []byte, ref *registry.Reference, prevIndexDesc ocispec.Descriptor) (ocispec.Descriptor, error) {
	if err := repo.Push(ctx, desc, bytes.NewReader(indexBytes)); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to save index: %w", err)
	}
	if err := repo.Tag(ctx, desc, ref.Reference); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to tag index: %w", err)
	}
	if prevIndexDesc.Digest != "" && prevIndexDesc.Digest != desc.Digest {
		if tags := repo.GetTags(prevIndexDesc); len(tags) == 0 {
			if err := repo.Delete(ctx, prevIndexDesc); err != nil {
				output.Logf(output.LogLevelWarn, "Failed to remove previous version of index %s: %s", prevIndexDesc.Digest, err)
			}
		}
	}
	return desc, nil
}

func displayRef(ref *registry.Reference) string {
	return artifact.FormatRepositoryForDisplay(ref.String())
}
