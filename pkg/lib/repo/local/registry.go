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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

type LocalRepo interface {
	GetRepoName() string
	BlobPath(ocispec.Descriptor) string
	GetAllModels() []ocispec.Descriptor
	GetTags(ocispec.Descriptor) []string
	PullModel(context.Context, oras.ReadOnlyTarget, registry.Reference, *options.NetworkOptions) (ocispec.Descriptor, error)
	EnsureDirs(ocispec.Descriptor) error
	oras.Target
	content.Deleter
	content.Untagger
}

type localRepo struct {
	storagePath string
	nameRef     string
	localIndex  *localIndex
	*oci.Store
}

func NewLocalRepo(storagePath string, ref *registry.Reference) (LocalRepo, error) {
	nameRef := path.Join(ref.Registry, ref.Repository)
	return newLocalRepoForName(storagePath, nameRef)
}

func newLocalRepoForName(storagePath, name string) (LocalRepo, error) {
	repo := &localRepo{}
	repo.storagePath = storagePath
	repo.nameRef = name

	store, err := oci.New(storagePath)
	if err != nil {
		return nil, err
	}
	repo.Store = store

	// Initialize repo-specific index.json
	localIndex, err := newLocalIndex(storagePath, name)
	if err != nil {
		return nil, err
	}
	repo.localIndex = localIndex

	return repo, nil
}

func GetAllLocalRepos(storagePath string) ([]LocalRepo, error) {
	entries, err := os.ReadDir(storagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}

	var repos []LocalRepo
	for _, dirEntry := range entries {
		if dirEntry.IsDir() {
			continue
		}
		if !constants.FileIsLocalIndex(dirEntry.Name()) {
			continue
		}
		repoName, err := constants.RepoForIndexJsonPath(dirEntry.Name())
		if err != nil {
			return nil, err
		}
		repo, err := newLocalRepoForName(storagePath, repoName)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	// Sort alphabetically
	slices.SortFunc(repos, func(a, b LocalRepo) int {
		return strings.Compare(a.GetRepoName(), b.GetRepoName())
	})

	return repos, nil
}

// GetRepoName returns the string representation of <registry>/<repository> for the current local repo.
func (lr *localRepo) GetRepoName() string {
	return lr.nameRef
}

func (lr *localRepo) BlobPath(desc ocispec.Descriptor) string {
	return filepath.Join(lr.storagePath, ocispec.ImageBlobsDir, desc.Digest.Algorithm().String(), desc.Digest.Encoded())
}

func (lr *localRepo) Delete(ctx context.Context, target ocispec.Descriptor) error {
	output.SafeLogf(output.LogLevelTrace, "Deleting digest %s in local repository %s", target.Digest.String(), lr.nameRef)
	if target.MediaType != ocispec.MediaTypeImageManifest {
		return lr.Store.Delete(ctx, target)
	}

	canDelete, err := canSafelyDeleteManifest(ctx, lr.storagePath, target)
	if err != nil {
		return fmt.Errorf("failed to check if manifest can be deleted: %w", err)
	}
	if canDelete {
		if err := lr.Store.Delete(ctx, target); err != nil {
			return err
		}
	}
	return lr.localIndex.delete(target)
}

func (lr *localRepo) Exists(ctx context.Context, target ocispec.Descriptor) (exists bool, err error) {
	if target.MediaType == ocispec.MediaTypeImageManifest {
		exists, err = lr.localIndex.exists(target), nil
	} else {
		exists, err = lr.Store.Exists(ctx, target)
	}
	if err != nil {
		return false, err
	}
	if exists {
		output.SafeLogf(output.LogLevelTrace, "Found digest %s in local repository %s", target.Digest.String(), lr.nameRef)
	} else {
		output.SafeLogf(output.LogLevelTrace, "Digest %s does not exist in local repository %s", target.Digest.String(), lr.nameRef)
	}
	return exists, nil
}

func (lr *localRepo) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	output.SafeLogf(output.LogLevelTrace, "Fetching digest %s in local repository %s", target.Digest.String(), lr.nameRef)
	if target.MediaType == ocispec.MediaTypeImageManifest {
		if exists := lr.localIndex.exists(target); !exists {
			return nil, errdef.ErrNotFound
		}
	}
	// Handle empty descriptors and descriptors using the data field
	if target.Data != nil {
		return io.NopCloser(bytes.NewReader(target.Data)), nil
	}

	return lr.Store.Fetch(ctx, target)
}

func (lr *localRepo) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	output.SafeLogf(output.LogLevelTrace, "Pushing digest %s to local repository %s", expected.Digest.String(), lr.nameRef)
	if expected.MediaType == ocispec.MediaTypeImageManifest {
		// Attempting to push a manifest to oci.Store will return an error if it already exists.
		// Normally, clients check before pushing, but in our case, the manifest may exist in the
		// oci.Store but not the local index. As a result, we have to check if it exists before pushing.
		exists, err := lr.Store.Exists(ctx, expected)
		if err != nil {
			return err
		}
		if !exists {
			if err := lr.Store.Push(ctx, expected, content); err != nil {
				return err
			}
		}
		return lr.localIndex.addManifest(expected)
	}
	return lr.Store.Push(ctx, expected, content)
}

func (lr *localRepo) Resolve(_ context.Context, reference string) (ocispec.Descriptor, error) {
	output.SafeLogf(output.LogLevelTrace, "Resolving reference %s in local repository %s", reference, lr.nameRef)
	return lr.localIndex.resolve(reference)
}

func (lr *localRepo) Tag(_ context.Context, desc ocispec.Descriptor, reference string) error {
	output.SafeLogf(output.LogLevelTrace, "Tagging digest %s with %s in local repository %s", desc.Digest.String(), reference, lr.nameRef)
	// TODO: should we tag it in the general index.json too?
	return lr.localIndex.tag(desc, reference)
}

func (lr *localRepo) Untag(_ context.Context, reference string) error {
	output.SafeLogf(output.LogLevelTrace, "Untagging reference %s in local repository %s", reference, lr.nameRef)
	return lr.localIndex.untag(reference)
}

func (lr *localRepo) GetAllModels() []ocispec.Descriptor {
	return lr.localIndex.Manifests
}

func (lr *localRepo) GetTags(desc ocispec.Descriptor) []string {
	return lr.localIndex.listTags(desc)
}

func (lr *localRepo) EnsureDirs(desc ocispec.Descriptor) error {
	path := filepath.Join(lr.storagePath, ocispec.ImageBlobsDir, desc.Digest.Algorithm().String())
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to set up directories for local storage: %w", err)
	}
	return nil
}

var _ LocalRepo = (*localRepo)(nil)
