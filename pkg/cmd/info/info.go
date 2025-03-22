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

package info

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
)

func getInfo(ctx context.Context, opts *infoOptions) (*artifact.KitFile, error) {
	if opts.checkRemote {
		return getRemoteConfig(ctx, opts)
	} else {
		return getLocalConfig(ctx, opts)
	}
}

func getLocalConfig(ctx context.Context, opts *infoOptions) (*artifact.KitFile, error) {
	storageRoot := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageRoot, opts.modelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}
	_, _, config, err := util.ResolveManifestAndConfig(ctx, localRepo, opts.modelRef.Reference)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getRemoteConfig(ctx context.Context, opts *infoOptions) (*artifact.KitFile, error) {
	repository, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, err
	}
	_, _, config, err := util.ResolveManifestAndConfig(ctx, repository, opts.modelRef.Reference)
	if err != nil {
		return nil, err
	}
	return config, nil
}
