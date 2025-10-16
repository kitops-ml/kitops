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

package ignore

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreMatches(t *testing.T) {
	tests := []struct {
		name         string
		kitIgnore    []string
		layerPaths   []string
		curPath      string
		curLayerPath string
		shouldIgnore bool
	}{
		{
			name:         "Ignores files in directory",
			kitIgnore:    []string{"dir1"},
			layerPaths:   []string{},
			curPath:      "dir1/subdir1/subdir2/file.txt",
			curLayerPath: "",
			shouldIgnore: true,
		},
		{
			name:         "Ignores files with wildcard",
			kitIgnore:    []string{"dir1/*.txt"},
			layerPaths:   []string{},
			curPath:      "dir1/testfile.txt",
			curLayerPath: "",
			shouldIgnore: true,
		},
		{
			name:         "Ignores files with '**' wildcard",
			kitIgnore:    []string{"**/testfile.txt"},
			layerPaths:   []string{},
			curPath:      "dir1/subdir1/subdir2/testfile.txt",
			curLayerPath: "",
			shouldIgnore: true,
		},
		{
			name:         "Can explicitly include files",
			kitIgnore:    []string{"dir1", "!dir1/testfile.txt"},
			layerPaths:   []string{},
			curPath:      "dir1/testfile.txt",
			curLayerPath: "",
			shouldIgnore: false,
		},
		{
			name:         "Test intersecting layers exclusion",
			kitIgnore:    []string{},
			layerPaths:   []string{"main", "main/subdir"},
			curPath:      "main/subdir/testfile.txt",
			curLayerPath: "main",
			shouldIgnore: true,
		},
		{
			name:         "Test intersecting layers inclusion",
			kitIgnore:    []string{},
			layerPaths:   []string{"main", "main/subdir"},
			curPath:      "main/subdir/testfile.txt",
			curLayerPath: "main/subdir",
			shouldIgnore: false,
		},
		{
			name:         "Test intersecting layers inclusion with kitignore",
			kitIgnore:    []string{"**/testfile.txt"},
			layerPaths:   []string{"main", "main/subdir"},
			curPath:      "main/subdir/testfile.txt",
			curLayerPath: "main/subdir",
			shouldIgnore: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testKitfile := &artifact.KitFile{}
			for _, layerPath := range tt.layerPaths {
				testKitfile.Code = append(testKitfile.Code, artifact.Code{Path: layerPath})
			}
			ignore, err := New(tt.kitIgnore, testKitfile)
			assert.NoError(t, err)

			ignored, err := ignore.Matches(tt.curPath, tt.curLayerPath)
			assert.NoError(t, err)
			assert.Equal(t, tt.shouldIgnore, ignored)
		})
	}
}
