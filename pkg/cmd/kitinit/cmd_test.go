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

package kitinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/hf"
	"github.com/stretchr/testify/assert"
)

func TestDetectRemoteRepo(t *testing.T) {
	// Create a temporary local directory for testing filesystem detection
	tmpDir := t.TempDir()
	localNestedDir := filepath.Join(tmpDir, "models", "my-model")
	err := os.MkdirAll(localNestedDir, 0755)
	assert.NoError(t, err)

	testcases := []struct {
		input            string
		expectedIsRemote bool
		expectedRepo     string
		expectedType     hf.RepositoryType
	}{
		// HuggingFace URLs should be detected as remote
		{input: "https://huggingface.co/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeModel},
		{input: "https://huggingface.co/datasets/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeDataset},
		{input: "huggingface.co/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeModel},
		{input: "huggingface.co/datasets/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeDataset},
		{input: "http://huggingface.co/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeModel},

		// org/repo pattern that doesn't exist locally should be treated as remote
		{input: "nonexistent-org/nonexistent-repo", expectedIsRemote: true, expectedRepo: "nonexistent-org/nonexistent-repo", expectedType: hf.RepoTypeModel},
		{input: "datasets/org/repo", expectedIsRemote: true, expectedRepo: "org/repo", expectedType: hf.RepoTypeDataset},

		// Explicit local path patterns should not be detected as remote
		{input: ".", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},
		{input: "./my-model", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},
		{input: "/absolute/path/to/model", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},
		{input: "../relative/path", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},

		// CRITICAL: Local nested directories that exist should be treated as local, not remote
		{input: localNestedDir, expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},

		// Non-HuggingFace URLs should not be detected as remote (for now)
		{input: "https://github.com/org/repo", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},
		{input: "https://example.com/org/repo", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},

		// Malicious URLs should not be detected as remote
		{input: "https://huggingface.co.evil.com/org/repo", expectedIsRemote: false, expectedRepo: "", expectedType: hf.RepoTypeUnknown},
	}

	for _, tt := range testcases {
		t.Run(fmt.Sprintf("handles %s", tt.input), func(t *testing.T) {
			isRemote, repo, repoType := detectRemoteRepo(tt.input)
			assert.Equal(t, tt.expectedIsRemote, isRemote, "isRemote mismatch")
			assert.Equal(t, tt.expectedRepo, repo, "repo mismatch")
			assert.Equal(t, tt.expectedType, repoType, "repoType mismatch")
		})
	}
}

func TestBuildPackageFromRepo(t *testing.T) {
	testcases := []struct {
		name            string
		repo            string
		inputName       string
		inputDesc       string
		inputAuthor     string
		expectedName    string
		expectedDesc    string
		expectedAuthors []string
	}{
		{
			name:            "extracts name and author from repo",
			repo:            "myorg/mymodel",
			expectedName:    "mymodel",
			expectedAuthors: []string{"myorg"},
		},
		{
			name:            "user-provided name overrides repo name",
			repo:            "myorg/mymodel",
			inputName:       "custom-name",
			expectedName:    "custom-name",
			expectedAuthors: []string{"myorg"},
		},
		{
			name:            "user-provided author overrides repo org",
			repo:            "myorg/mymodel",
			inputAuthor:     "custom-author",
			expectedName:    "mymodel",
			expectedAuthors: []string{"custom-author"},
		},
		{
			name:            "user-provided description is used",
			repo:            "myorg/mymodel",
			inputDesc:       "My model description",
			expectedName:    "mymodel",
			expectedDesc:    "My model description",
			expectedAuthors: []string{"myorg"},
		},
		{
			name:            "all user-provided values override defaults",
			repo:            "myorg/mymodel",
			inputName:       "custom-name",
			inputDesc:       "Custom description",
			inputAuthor:     "custom-author",
			expectedName:    "custom-name",
			expectedDesc:    "Custom description",
			expectedAuthors: []string{"custom-author"},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			pkg := buildPackageFromRepo(tt.repo, tt.inputName, tt.inputDesc, tt.inputAuthor)
			assert.Equal(t, tt.expectedName, pkg.Name)
			assert.Equal(t, tt.expectedDesc, pkg.Description)
			assert.Equal(t, tt.expectedAuthors, pkg.Authors)
		})
	}
}

func TestCompleteExpandsTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	opts := &initOptions{}
	ctx := context.WithValue(context.Background(), constants.ConfigKey{}, "/tmp")
	err = opts.complete(ctx, []string{"~/model"})
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "model"), opts.path)
}
