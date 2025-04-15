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
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/lib/constants"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
)

type LocalStorage interface {
	GetRepo() string
	GetIndex() (*ocispec.Index, error)
	getStorePath() string
	oras.Target
	content.Deleter
	content.Untagger
}

type LocalStore struct {
	storePath string
	repo      string
	*oci.Store
}

// GetAllLocalStores returns all local OCI indexes in the storageRoot path. As OCI
// indexes only support tags and not registry/repository, the storageRoot may
// contain multiple OCI indexes at separate paths (stored in
// <storageRoot>/registry/repository/index.json)
func GetAllLocalStores(storageRoot string) ([]LocalStorage, error) {
	subDirs, err := findStoragePaths(storageRoot)
	if err != nil {
		return nil, err
	}
	var stores []LocalStorage
	for _, subDir := range subDirs {
		// convert to forward slashes for repo
		repo := filepath.ToSlash(subDir)
		storePath := filepath.Join(storageRoot, subDir)
		ociStore, err := oci.New(storePath)
		if err != nil {
			return nil, err
		}
		localStore := &LocalStore{
			storePath: storePath,
			repo:      repo,
			Store:     ociStore,
		}
		stores = append(stores, localStore)
	}
	return stores, nil
}

// NewLocalStore returns a new LocalStorage for the provided *registry.Reference's
// registry and repository. The returned storage will only contain blobs present in
// this index.
func NewLocalStore(storageRoot string, ref *registry.Reference) (LocalStorage, error) {
	storePath := storageRoot
	repo := ""
	if ref != nil {
		repo = path.Join(ref.Registry, ref.Repository)
		storePath = filepath.Join(storePath, ref.Registry, ref.Repository)
	}
	store, err := oci.New(storePath)
	if err != nil {
		return nil, err
	}
	return &LocalStore{
		storePath: storePath,
		repo:      repo,
		Store:     store,
	}, nil
}

// GetIndex is a shortcut to reading the index.json for an OCI index, allowing for
// listing all manfiests stored.
func (s *LocalStore) GetIndex() (*ocispec.Index, error) {
	return parseIndexJson(s.storePath)
}

// GetRepo returns the registry and repository for the current OCI store.
func (s *LocalStore) GetRepo() string {
	return s.repo
}

func (s *LocalStore) getStorePath() string {
	return s.storePath
}

func BlobPathForManifest(store LocalStorage, desc ocispec.Descriptor) string {
	return filepath.Join(store.getStorePath(), "blobs", "sha256", desc.Digest.Encoded())
}

// findStoragePaths walks the filesystem rooted at storageRoot looking for index.json
// files that represent OCI indexes, returning a list of paths relative to storageRoot.
func findStoragePaths(storageRoot string) ([]string, error) {
	var indexPaths []string
	err := filepath.WalkDir(storageRoot, func(file string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "index.json" && !info.IsDir() {
			dir := filepath.Dir(file)
			relDir, err := filepath.Rel(storageRoot, dir)
			if err != nil {
				return err
			}
			if relDir != "." {
				indexPaths = append(indexPaths, relDir)
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}
	return indexPaths, nil
}

// parseIndexJson parses the OCI index.json stored in the OCI index at storageHome
func parseIndexJson(storageHome string) (*ocispec.Index, error) {
	indexBytes, err := os.ReadFile(constants.IndexJsonPath(storageHome))
	if err != nil {
		if os.IsNotExist(err) {
			return &ocispec.Index{}, nil
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	index := &ocispec.Index{}
	if err := json.Unmarshal(indexBytes, index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return index, nil
}
