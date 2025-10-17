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

package unpack

import (
	"context"
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
)

// getStoreForRef returns the appropriate store (local or remote) for a ModelKit reference.
func getStoreForRef(ctx context.Context, opts *UnpackOptions) (oras.Target, error) {
	storageHome := constants.StoragePath(opts.ConfigHome)
	localRepo, err := local.NewLocalRepo(storageHome, opts.ModelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %s\n", err)
	}

	if _, err := localRepo.Resolve(ctx, opts.ModelRef.Reference); err == nil {
		// Reference is present in local storage
		return localRepo, nil
	}

	if opts.ModelRef.Registry == util.DefaultRegistry {
		return nil, fmt.Errorf("not found")
	}
	// Not in local storage, check remote
	repo, err := remote.NewRepository(ctx, opts.ModelRef.Registry, opts.ModelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, fmt.Errorf("could not resolve repository %s in registry %s", opts.ModelRef.Repository, opts.ModelRef.Registry)
	}
	if _, err := repo.Resolve(ctx, opts.ModelRef.Reference); err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, fmt.Errorf("reference %s is not present in local storage and could not be found in remote", opts.ModelRef.String())
		}
		return nil, fmt.Errorf("unexpected error retrieving reference from remote: %w", err)
	}

	return repo, nil
}
