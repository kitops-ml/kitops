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
