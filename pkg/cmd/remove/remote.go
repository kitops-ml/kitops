// Copyright 2025 The KitOps Authors.
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

package remove

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

func removeRemoteModel(ctx context.Context, opts *removeOptions) error {
	repository, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return err
	}

	desc, err := repository.Resolve(ctx, opts.modelRef.Reference)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return fmt.Errorf("model %s not found", util.FormatRepositoryForDisplay(opts.modelRef.String()))
		}
		return fmt.Errorf("error resolving modelkit: %w", err)
	}

	// If user supplied a tag instead of a digest, only do an untag; assume the remote will prune untagged ModelKits
	if !util.ReferenceIsDigest(opts.modelRef.Reference) && !opts.forceDelete {
		output.Infof("Untagging remote ModelKit %s", util.FormatRepositoryForDisplay(opts.modelRef.String()))
		return untagRemoteModel(ctx, opts.modelRef.Reference, repository)
	}

	// Otherwise, delete the ModelKit itself, which will delete all tags that point to it
	deleteRef := *opts.modelRef
	deleteRef.Reference = desc.Digest.String()
	output.Infof("Deleting remote ModelKit %s", util.FormatRepositoryForDisplay(deleteRef.String()))
	if err := repository.Delete(ctx, desc); err != nil {
		if errResp, ok := err.(*errcode.ErrorResponse); ok && errResp.StatusCode == http.StatusMethodNotAllowed {
			return fmt.Errorf("removing models is unsupported by registry %s", opts.modelRef.Registry)
		}
		return fmt.Errorf("failed to remove remote model: %w", err)
	}
	return nil
}

func untagRemoteModel(ctx context.Context, tag string, repo registry.Repository) error {
	// Temporary: registry.Repository does not support the Untagger interface, so we need to be explicit about typing
	// This should be removed once the Untagger implementation is moved into the oras-go library.
	untaggerRepo, ok := repo.(content.Untagger)
	if !ok {
		return fmt.Errorf("remote repository implementation does not support untagging ModelKits")
	}
	return untaggerRepo.Untag(ctx, tag)
}
