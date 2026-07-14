// Copyright 2026 The KitOps Authors.
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

package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMkCacheDirReusesDeterministicDirectory(t *testing.T) {
	originalCacheHome := cacheHome()
	t.Cleanup(func() {
		SetCacheHome(originalCacheHome)
	})
	SetCacheHome(t.TempDir())

	firstDir, cleanup, err := MkCacheDir(CacheImportSubdir, "repo-ref")
	require.NoError(t, err)
	t.Cleanup(cleanup)

	partialPath := filepath.Join(firstDir, "partial-file")
	require.NoError(t, os.WriteFile(partialPath, []byte("partial"), 0600))

	secondDir, secondCleanup, err := MkCacheDir(CacheImportSubdir, "repo-ref")
	require.NoError(t, err)
	t.Cleanup(secondCleanup)

	require.Equal(t, firstDir, secondDir)
	contents, err := os.ReadFile(filepath.Join(secondDir, "partial-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("partial"), contents)
}

func TestMkCacheDirRejectsUnsafeCacheKeys(t *testing.T) {
	originalCacheHome := cacheHome()
	t.Cleanup(func() {
		SetCacheHome(originalCacheHome)
	})
	SetCacheHome(t.TempDir())

	for _, cacheKey := range []string{
		".",
		"..",
		"nested/path",
		`nested\path`,
		filepath.Join("..", "outside"),
	} {
		t.Run(cacheKey, func(t *testing.T) {
			_, _, err := MkCacheDir(CacheImportSubdir, cacheKey)
			require.Error(t, err)
		})
	}
}
