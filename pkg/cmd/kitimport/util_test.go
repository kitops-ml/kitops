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

package kitimport

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
)

func TestExtractRepoFromURL(t *testing.T) {
	testcases := []struct {
		input           string
		expected        string
		expectErrRegexp string
	}{
		{input: "organization/repository", expected: "organization/repository"},
		{input: "https://example.com/org/repo", expected: "org/repo"},
		{input: "https://huggingface.co/org/repo", expected: "org/repo"},
		{input: "https://github.com/org/repo", expected: "org/repo"},
		{input: "organization/repository.with-dots.and_CAPS", expected: "organization/repository.with-dots.and_CAPS"},
		{input: "https://huggingface.co/org/trailing-slash/", expected: "org/trailing-slash"},
		{input: "https://github.com/org/repo.git", expected: "org/repo.git"},
		{input: ":///invalidURL", expectErrRegexp: "failed to parse url.*"},
		{input: "too/many/path/segments", expectErrRegexp: "could not extract organization and repository from.*"},
		{input: "https://github.com/kitops-ml/github.com/kitops-ml/kitops/tree/main", expectErrRegexp: "could not extract organization and repository from.*"},
	}

	for _, tt := range testcases {
		t.Run(fmt.Sprintf("handles %s", tt.input), func(t *testing.T) {
			actual, actualErr := extractRepoFromURL(tt.input)
			if tt.expectErrRegexp != "" {
				if !assert.Error(t, actualErr) {
					return
				}
				assert.Regexp(t, tt.expectErrRegexp, actualErr.Error())
			} else {
				if !assert.NoError(t, actualErr) {
					return
				}
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestFilterDirectoryListing(t *testing.T) {
	listing := &kfgen.DirectoryListing{
		Name: ".",
		Path: ".",
		Files: []kfgen.FileListing{
			{Name: "model.safetensors", Path: "model.safetensors"},
			{Name: "config.json", Path: "config.json"},
			{Name: "README.md", Path: "README.md"},
		},
		Subdirs: []kfgen.DirectoryListing{
			{
				Name: "onnx",
				Path: "onnx",
				Files: []kfgen.FileListing{
					{Name: "model.onnx", Path: "onnx/model.onnx"},
				},
			},
			{
				Name: "gguf",
				Path: "gguf",
				Files: []kfgen.FileListing{
					{Name: "model-q4_0.gguf", Path: "gguf/model-q4_0.gguf"},
				},
			},
		},
	}

	tests := []struct {
		name      string
		filters   []string
		wantFiles []string
		wantDirs  []string
	}{
		{
			name:      "Filter by extension",
			filters:   []string{"*.gguf"},
			wantFiles: []string{"model-q4_0.gguf"},
			wantDirs:  []string{"gguf"},
		},
		{
			name:      "Filter by path",
			filters:   []string{"onnx/*"},
			wantFiles: []string{"model.onnx"},
			wantDirs:  []string{"onnx"},
		},
		{
			name:      "Multiple filters",
			filters:   []string{"*.safetensors", "config.json"},
			wantFiles: []string{"model.safetensors", "config.json"},
			wantDirs:  []string{},
		},
		{
			name:      "No matches",
			filters:   []string{"*.pth"},
			wantFiles: []string{},
			wantDirs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterDirectoryListing(listing, tt.filters)
			assert.Equal(t, tt.wantFiles, collectFiles(got))
			assert.Equal(t, tt.wantDirs, collectDirs(got))
		})
	}
}

func collectFiles(l *kfgen.DirectoryListing) []string {
	files := []string{}
	for _, f := range l.Files {
		files = append(files, f.Name)
	}
	for _, s := range l.Subdirs {
		files = append(files, collectFiles(&s)...)
	}
	return files
}

func collectDirs(l *kfgen.DirectoryListing) []string {
	dirs := []string{}
	for _, s := range l.Subdirs {
		dirs = append(dirs, s.Name)
		dirs = append(dirs, collectDirs(&s)...)
	}
	return dirs
}
