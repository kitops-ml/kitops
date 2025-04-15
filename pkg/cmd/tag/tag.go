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

package tag

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
)

func RunTag(ctx context.Context, options *tagOptions) error {
	storageHome := constants.StoragePath(options.configHome)
	sourceRepo, err := local.NewLocalRepo(storageHome, options.sourceRef)
	if err != nil {
		return fmt.Errorf("failed to open local storage: %w", err)
	}
	descriptor, err := oras.Resolve(ctx, sourceRepo, options.sourceRef.Reference, oras.ResolveOptions{})
	if err != nil {
		if err == errdef.ErrNotFound {
			return fmt.Errorf("model %s not found", options.sourceRef.String())
		}
		return fmt.Errorf("error resolving model: %s", err)
	}
	if options.sourceRef.Registry == options.targetRef.Registry && options.sourceRef.Repository == options.targetRef.Repository {
		err = sourceRepo.Tag(ctx, descriptor, options.targetRef.Reference)
		if err != nil {
			return fmt.Errorf("failed to tag reference %s: %w", options.targetRef, err)
		}
		return nil
	}

	// Target is under a different repo name (org/repo pair); manifest needs to be pushed to _that_ local store
	// Note that since local repos all share the same blob storage, only the manifest will need to be copied.
	targetRepo, err := local.NewLocalRepo(storageHome, options.targetRef)
	if err != nil {
		return fmt.Errorf("failed to open local storage: %w", err)
	}
	_, err = oras.Copy(ctx, sourceRepo, options.sourceRef.Reference, targetRepo, options.targetRef.Reference, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to tag model: %w", err)
	}
	return nil
}
