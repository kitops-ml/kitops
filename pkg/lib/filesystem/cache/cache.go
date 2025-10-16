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

package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/output"
)

var cacheHomeDir = os.TempDir()

func SetCacheHome(cacheHome string) {
	cacheHomeDir = cacheHome
}

func cacheHome() string {
	return cacheHomeDir
}

type CacheSubDir string

const (
	CachePackSubdir   CacheSubDir = "pack"
	CacheImportSubdir CacheSubDir = "import"
)

// MkCacheDir creates a directory within configHome to be used for temporary storage and returns a function that can
// be called to remove it once it is no longer needed. If cacheKey is not empty, the cache directory will be
// deterministic and can be used to resume operations. Otherwise the directory will be generated with a random,
// non-colliding name.
func MkCacheDir(subDir CacheSubDir, cacheKey string) (cacheDir string, cleanup func(), err error) {
	cacheSubDir := filepath.Join(cacheHome(), string(subDir))
	if err := os.MkdirAll(cacheSubDir, 0700); err != nil {
		return "", nil, fmt.Errorf("failed to create cache directory %s: %w", cacheSubDir, err)
	}
	if cacheKey != "" {
		cacheDir = filepath.Join(cacheSubDir, cacheKey)
		if err := os.Mkdir(cacheDir, 0700); err != nil {
			return "", nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
		}
	} else {
		tmpDir, err := os.MkdirTemp(cacheSubDir, fmt.Sprintf("kitops_%s_", subDir))
		if err != nil {
			return "", nil, fmt.Errorf("failed to create cache directory in %s: %w", cacheSubDir, err)
		}
		cacheDir = tmpDir
	}

	cleanup = func() {
		if err := os.RemoveAll(cacheDir); err != nil {
			output.Logf(output.LogLevelWarn, "Failed to remove temporary directory %s: %s", cacheDir, err)
		}
	}
	return cacheDir, cleanup, nil
}

func MkCacheFile(subDir CacheSubDir, basename string) (tempFile *os.File, cleanup func(), err error) {
	cacheSubDir := filepath.Join(cacheHome(), string(subDir))
	if err := os.MkdirAll(cacheSubDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("failed to create cache directory %s: %w", cacheSubDir, err)
	}
	f, err := os.CreateTemp(cacheSubDir, basename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temporary file in %s: %w", cacheSubDir, err)
	}

	tempFilePath := filepath.Join(cacheSubDir, f.Name())
	cleanup = func() {
		if err := f.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			output.Errorf("Error closing temporary file %s: %s", tempFilePath, err)
		}
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			output.Errorf("Failed to remove temporary file %s: %w", tempFilePath, err)
		}
	}
	return f, cleanup, nil
}

func CleanCacheDir(subDir CacheSubDir) error {
	cacheSubDir := filepath.Join(cacheHome(), string(subDir))
	ds, err := os.ReadDir(cacheSubDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory %s: %w", cacheSubDir, err)
	}
	for _, d := range ds {
		entryPath := filepath.Join(cacheSubDir, d.Name())
		if d.IsDir() {
			if err := os.RemoveAll(entryPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("failed to remove directory %s: %w", entryPath, err)
			}
		} else {
			if err := os.Remove(entryPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("failed to remove file %s: %w", entryPath, err)
			}
		}
	}
	return nil
}

func StatCache() (totalSize int64, subdirsSize map[string]int64, err error) {
	getDirSize := func(dir string) (int64, error) {
		var dirSize int64
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to examine %s: %w", path, err)
			}
			dirSize = dirSize + info.Size()
			return nil
		})
		if err != nil {
			return 0, err
		}
		return dirSize, nil
	}

	ds, err := os.ReadDir(cacheHome())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, map[string]int64{}, nil
		}
		return 0, nil, fmt.Errorf("failed to read cache directory: %w", err)
	}
	subdirsSize = map[string]int64{}
	for _, dirEntry := range ds {
		size, err := getDirSize(filepath.Join(cacheHome(), dirEntry.Name()))
		if err != nil {
			return 0, nil, err
		}
		subdirsSize[dirEntry.Name()] = size
		totalSize = totalSize + size
	}

	return totalSize, subdirsSize, nil
}

func ClearCache() error {
	ds, err := os.ReadDir(cacheHome())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}
	for _, dirEntry := range ds {
		if dirEntry.IsDir() {
			output.Debugf("Removing cache directory %s", dirEntry.Name())
			os.RemoveAll(filepath.Join(cacheHome(), dirEntry.Name()))
		} else {
			output.Debugf("Removing cache file %s", dirEntry.Name())
			os.Remove(filepath.Join(cacheHome(), dirEntry.Name()))
		}
	}
	return nil
}
