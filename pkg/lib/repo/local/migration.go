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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

func NeedsMigrate(baseStoragePath string) (bool, error) {
	paths, err := findStoragePaths(baseStoragePath)
	if err != nil {
		return false, err
	}
	return len(paths) > 0, nil
}

func MigrateStorage(ctx context.Context, baseStoragePath string) error {
	localStores, err := GetAllLocalStores(baseStoragePath)
	if err != nil {
		return fmt.Errorf("failed to migrate local storage: %w", err)
	}
	pb := output.GenericProgressBar("Migrating", "Migration done!", int64(len(localStores)))
	for _, localStore := range localStores {
		repoName := localStore.GetRepo()
		localRepo, err := newLocalRepoForName(baseStoragePath, repoName)
		if err != nil {
			return fmt.Errorf("failed to migrate local storage: %w", err)
		}
		toMigrateManifests, err := localStore.GetIndex()
		if err != nil {
			return fmt.Errorf("failed to migrate local storage: %w", err)
		}
		for _, desc := range toMigrateManifests.Manifests {
			tagOrDigest := string(desc.Digest)
			if tag := desc.Annotations[ocispec.AnnotationRefName]; tag != "" {
				tagOrDigest = tag
			}

			output.Debugf("Migrating model %s with reference %s to new storage", repoName, tagOrDigest)
			_, err := oras.Copy(ctx, localStore, tagOrDigest, localRepo, tagOrDigest, oras.DefaultCopyOptions)
			if err != nil {
				return fmt.Errorf("failed to migrate modelkit %s:%s: %w", repoName, tagOrDigest, err)
			}

			// Sanity check that copied objects exist in new store
			if exists, err := localRepo.Exists(ctx, desc); err != nil {
				return fmt.Errorf("error checking for successful migration: %w", err)
			} else if !exists {
				return fmt.Errorf("migrating modelkit %s:%s failed", repoName, tagOrDigest)
			}
		}
		// Clean up this repos blobs; we'll clean up the directories later
		repoDir := localStore.getStorePath()
		if err := os.RemoveAll(repoDir); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to clean up directory %s after migration: %s", repoDir, err)
		}
		pb.Increment()
	}
	pb.Done()

	// Remove old storage directories
	for _, localStore := range localStores {
		storeRepo := localStore.GetRepo()
		baseSubDir := strings.Split(storeRepo, "/")[0]
		rmDir := filepath.Join(baseStoragePath, baseSubDir)
		output.Debugf("Removing storage directory %s", rmDir)
		if err := os.RemoveAll(rmDir); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to clean up directory %s after migration: %s", rmDir, err)
		}
	}
	output.Debugf("Migration done!")
	return nil
}
