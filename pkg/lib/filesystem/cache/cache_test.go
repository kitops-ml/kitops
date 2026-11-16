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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanCacheDirKeepsRecentEntries(t *testing.T) {
	origHome := cacheHomeDir
	SetCacheHome(t.TempDir())
	t.Cleanup(func() { SetCacheHome(origHome) })

	subDir := filepath.Join(cacheHome(), string(CacheImportSubdir))
	old := time.Now().Add(-2 * staleCacheAge)

	setAge := func(path string, modTime time.Time) {
		require.NoError(t, os.Chtimes(path, modTime, modTime))
	}
	mkFile := func(path string) {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
		require.NoError(t, os.WriteFile(path, []byte("data"), 0600))
	}

	staleFile := filepath.Join(subDir, "stale-file")
	mkFile(staleFile)
	setAge(staleFile, old)

	recentFile := filepath.Join(subDir, "recent-file")
	mkFile(recentFile)

	staleDir := filepath.Join(subDir, "stale-dir")
	mkFile(filepath.Join(staleDir, "nested", "file"))
	setAge(filepath.Join(staleDir, "nested", "file"), old)
	setAge(filepath.Join(staleDir, "nested"), old)
	setAge(staleDir, old)

	// An old directory that is still being written to, e.g. a long download into a subdirectory
	activeDir := filepath.Join(subDir, "active-dir")
	mkFile(filepath.Join(activeDir, "nested", "file"))
	setAge(filepath.Join(activeDir, "nested"), old)
	setAge(activeDir, old)

	require.NoError(t, CleanCacheDir(CacheImportSubdir))

	assert.NoFileExists(t, staleFile)
	assert.NoDirExists(t, staleDir)
	assert.FileExists(t, recentFile)
	assert.FileExists(t, filepath.Join(activeDir, "nested", "file"))
}

func TestCleanCacheDirMissingSubdir(t *testing.T) {
	origHome := cacheHomeDir
	SetCacheHome(t.TempDir())
	t.Cleanup(func() { SetCacheHome(origHome) })

	assert.NoError(t, CleanCacheDir(CachePackSubdir))
}
