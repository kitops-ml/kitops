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

package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// CleanGitMetadata removes all git-related files from a repository
func CleanGitMetadata(dir string) error {
	filesToClean := []string{
		".git",
		".gitattributes",
		".gitignore",
	}

	for _, path := range filesToClean {
		fullPath := filepath.Join(dir, path)
		if err := os.RemoveAll(fullPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("error removing %s: %w", path, err)
		}
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Name() == ".gitignore" || d.Name() == ".gitattributes" {
			return os.Remove(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while cleaning git files from directory: %w", err)
	}

	return nil
}
