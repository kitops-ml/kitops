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

package kit

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

type TagOptions struct {
	ConfigHome string
	SourceRef  *registry.Reference
	TargetRef  *registry.Reference
}

func Tag(ctx context.Context, options *TagOptions) error {
	storageHome := constants.StoragePath(options.ConfigHome)
	sourceRepo, err := local.NewLocalRepo(storageHome, options.SourceRef)
	if err != nil {
		return fmt.Errorf("failed to open local storage: %w", err)
	}
	descriptor, err := oras.Resolve(ctx, sourceRepo, options.SourceRef.Reference, oras.ResolveOptions{})
	if err != nil {
		if err == errdef.ErrNotFound {
			return fmt.Errorf("model %s not found", options.SourceRef.String())
		}
		return fmt.Errorf("error resolving model: %s", err)
	}
	if options.SourceRef.Registry == options.TargetRef.Registry && options.SourceRef.Repository == options.TargetRef.Repository {
		err = sourceRepo.Tag(ctx, descriptor, options.TargetRef.Reference)
		if err != nil {
			return fmt.Errorf("failed to tag reference %s: %w", options.TargetRef, err)
		}
		return nil
	}

	targetRepo, err := local.NewLocalRepo(storageHome, options.TargetRef)
	if err != nil {
		return fmt.Errorf("failed to open local storage: %w", err)
	}
	_, err = oras.Copy(ctx, sourceRepo, options.SourceRef.Reference, targetRepo, options.TargetRef.Reference, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to tag model: %w", err)
	}
	return nil
}
