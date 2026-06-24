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
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/external/hf"

	"github.com/stretchr/testify/assert"
)

func TestHFSourceURI(t *testing.T) {
	tests := []struct {
		name     string
		repo     string
		repoType hf.RepositoryType
		want     string
	}{
		{name: "model", repo: "org/llama", repoType: hf.RepoTypeModel, want: "hf://org/llama"},
		{name: "dataset", repo: "org/squad", repoType: hf.RepoTypeDataset, want: "hf://datasets/org/squad"},
		// Same name across kinds must produce different URIs.
		{name: "model_collision_left", repo: "org/imagenet", repoType: hf.RepoTypeModel, want: "hf://org/imagenet"},
		{name: "dataset_collision_right", repo: "org/imagenet", repoType: hf.RepoTypeDataset, want: "hf://datasets/org/imagenet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hfSourceURI(tt.repo, tt.repoType))
		})
	}
}

func TestHFImportCacheKey(t *testing.T) {
	modelKey := hfImportCacheKey("org/repo", hf.RepoTypeModel, "abc123")

	assert.Equal(t, modelKey, hfImportCacheKey("org/repo", hf.RepoTypeModel, "abc123"))
	assert.NotEqual(t, modelKey, hfImportCacheKey("org/repo", hf.RepoTypeDataset, "abc123"))
	assert.NotEqual(t, modelKey, hfImportCacheKey("org/repo", hf.RepoTypeModel, "def456"))
	assert.NotContains(t, modelKey, "/")
}
