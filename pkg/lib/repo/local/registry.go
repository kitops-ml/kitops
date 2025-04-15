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

// GetRepo returns the registry and repository for the current OCI store.
func (r *localRepo) GetRepoName() string {
	return r.nameRef
}

func (r *localRepo) BlobPath(desc ocispec.Descriptor) string {
	return filepath.Join(r.storagePath, ocispec.ImageBlobsDir, desc.Digest.Algorithm().String(), desc.Digest.Encoded())
}

func (l *localRepo) Delete(ctx context.Context, target ocispec.Descriptor) error {
	if target.MediaType != ocispec.MediaTypeImageManifest {
		return l.Store.Delete(ctx, target)
	}

	canDelete, err := canSafelyDeleteManifest(ctx, l.storagePath, target)
	if err != nil {
		return fmt.Errorf("failed to check if manifest can be deleted: %w", err)
	}
	if canDelete {
		if err := l.Store.Delete(ctx, target); err != nil {
			return err
		}
	}
	return l.localIndex.delete(target)
}

func (l *localRepo) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	if target.MediaType == ocispec.MediaTypeImageManifest {
		return l.localIndex.exists(target), nil
	} else {
		return l.Store.Exists(ctx, target)
	}
}

func (l *localRepo) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	if target.MediaType == ocispec.MediaTypeImageManifest {
		if exists := l.localIndex.exists(target); !exists {
			return nil, errdef.ErrNotFound
		}
	}
	return l.Store.Fetch(ctx, target)
}

func (l *localRepo) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	if expected.MediaType == ocispec.MediaTypeImageManifest {
		// Attempting to push a manifest to oci.Store will return an error if it already exists.
		// Normally, clients check before pushing, but in our case, the manifest may exist in the
		// oci.Store but not the local index. As a result, we have to check if it exists before pushing.
		exists, err := l.Store.Exists(ctx, expected)
		if err != nil {
			return err
		}
		if !exists {
			if err := l.Store.Push(ctx, expected, content); err != nil {
				return err
			}
		}
		return l.localIndex.addManifest(expected)
	} else {
		return l.Store.Push(ctx, expected, content)
	}
}

func (l *localRepo) Resolve(_ context.Context, reference string) (ocispec.Descriptor, error) {
	return l.localIndex.resolve(reference)
}

func (l *localRepo) Tag(_ context.Context, desc ocispec.Descriptor, reference string) error {
	// TODO: should we tag it in the general index.json too?
	return l.localIndex.tag(desc, reference)
}

func (l *localRepo) Untag(_ context.Context, reference string) error {
	return l.localIndex.untag(reference)
}

func (l *localRepo) GetAllModels() []ocispec.Descriptor {
	return l.localIndex.Manifests
}

func (l *localRepo) GetTags(desc ocispec.Descriptor) []string {
	return l.localIndex.listTags(desc)
}

var _ LocalRepo = (*localRepo)(nil)
