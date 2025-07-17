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

package local

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

func (l *localRepo) PushModel(ctx context.Context, dest oras.Target, ref registry.Reference, opts *options.NetworkOptions) (ocispec.Descriptor, error) {
	// Resolve the reference to get the manifest descriptor
	desc, err := l.Resolve(ctx, ref.Reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to resolve reference %s: %w", ref.Reference, err)
	}

	if desc.MediaType != ocispec.MediaTypeImageManifest {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("expected manifest for push but got %s", desc.MediaType)
	}

	progress := output.NewPushProgress(ctx)

	// Use standard oras.Copy API instead of custom implementation
	copyOptions := oras.CopyOptions{}

	// Configure concurrency if specified
	if opts.Concurrency > 0 {
		copyOptions.Concurrency = opts.Concurrency
	}

	// Set up progress tracking by wrapping the destination
	wrappedDest, _ := output.WrapTarget(dest)

	// Use oras.Copy to handle the entire push operation
	_, err = oras.Copy(ctx, l, ref.Reference, wrappedDest, ref.Reference, copyOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to copy model: %w", err)
	}

	progress.Done()

	return desc, nil
}
