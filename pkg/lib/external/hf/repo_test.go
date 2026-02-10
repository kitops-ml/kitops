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

package hf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHuggingFaceRepo(t *testing.T) {
	testcases := []struct {
		input           string
		expectedRepo    string
		expectedType    RepositoryType
		expectErrRegexp string
	}{
		{input: "https://huggingface.co/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeModel},
		{input: "https://huggingface.co/datasets/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeDataset},
		{input: "org/repo", expectedRepo: "org/repo", expectedType: RepoTypeModel},
		{input: "datasets/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeDataset},
		{input: "datasets/only-one-segment", expectErrRegexp: "could not extract repository"},
		{input: "https://example.com/org/repo", expectErrRegexp: "not a Hugging Face repository"},
		// Test for malicious URL that contains "huggingface.co" in hostname
		{input: "https://huggingface.co.evil.com/org/repo", expectErrRegexp: "not a Hugging Face repository"},
		// Test for scheme-less dataset URL
		{input: "huggingface.co/datasets/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeDataset},
		// Test for scheme-less model URL
		{input: "huggingface.co/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeModel},
		// Test URL with trailing slashes
		{input: "https://huggingface.co/org/repo/", expectedRepo: "org/repo", expectedType: RepoTypeModel},
		{input: "https://huggingface.co/datasets/org/repo/", expectedRepo: "org/repo", expectedType: RepoTypeDataset},
		// Test with http (not https)
		{input: "http://huggingface.co/org/repo", expectedRepo: "org/repo", expectedType: RepoTypeModel},
	}

	for _, tt := range testcases {
		t.Run(fmt.Sprintf("handles %s", tt.input), func(t *testing.T) {
			actualRepo, actualType, actualErr := ParseHuggingFaceRepo(tt.input)
			if tt.expectErrRegexp != "" {
				if !assert.Error(t, actualErr) {
					return
				}
				assert.Regexp(t, tt.expectErrRegexp, actualErr.Error())
			} else {
				if !assert.NoError(t, actualErr) {
					return
				}
				assert.Equal(t, tt.expectedRepo, actualRepo)
				assert.Equal(t, tt.expectedType, actualType)
			}
		})
	}
}
