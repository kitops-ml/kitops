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
	"testing"

	"github.com/stretchr/testify/assert"
)

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
