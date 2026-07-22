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

package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
)

type DirectoryListing struct {
	Name       string
	Path       string
	ContextDir string // root directory that relative paths are resolved against
	Files      []FileListing
	Subdirs    []DirectoryListing
}

type FileListing struct {
	Name string
	Path string
	Size int64
}

func DirectoryListingFromFS(contextDir string) (*DirectoryListing, error) {
	dl, err := genDirListingFromPath(".", contextDir)
	if err != nil {
		return nil, err
	}
	dl.ContextDir = contextDir
	return dl, nil
}

func genDirListingFromPath(curDir, contextDir string) (*DirectoryListing, error) {
	dirName := filepath.Base(curDir)
	result := &DirectoryListing{
		Name:       dirName,
		Path:       curDir,
		ContextDir: contextDir,
	}

	fullPath := filepath.Join(contextDir, curDir)
	ds, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}
	for _, dirEntry := range ds {
		// Skip OS-specific files that should not be included in the Kitfile
		if ignore.ShouldIgnoreFile(dirEntry.Name()) {
			continue
		}
		relPath := filepath.Join(curDir, dirEntry.Name())
		fullPath := filepath.Join(contextDir, relPath)
		t := dirEntry.Type()
		switch {
		case t.IsDir():
			dirListing, err := genDirListingFromPath(relPath, contextDir)
			if err != nil {
				return nil, err
			}
			result.Subdirs = append(result.Subdirs, *dirListing)
		case t.IsRegular():
			info, err := dirEntry.Info()
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %s: %w", fullPath, err)
			}
			result.Files = append(result.Files, FileListing{
				Name: dirEntry.Name(),
				Path: filepath.ToSlash(relPath),
				Size: info.Size(),
			})
		default:
			continue
		}
	}

	return result, nil
}
