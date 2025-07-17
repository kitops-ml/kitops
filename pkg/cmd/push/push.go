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

package push

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func PushModel(ctx context.Context, localRepo local.LocalRepo, repo registry.Repository, opts *pushOptions) (ocispec.Descriptor, error) {
	trackedRepo, logger := output.WrapTarget(repo)

	// Use the optimized PushModel implementation
	desc, err := localRepo.PushModel(ctx, trackedRepo, *opts.destModelRef, &opts.NetworkOptions)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to push model: %w", err)
	}
	logger.Wait()

	return desc, err
}
