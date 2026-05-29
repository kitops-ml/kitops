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

package artifact

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2/registry"
)

type kitfileGenerateTestCase struct {
	Name        string
	Description string  `yaml:"description"`
	Manifest    string  `yaml:"manifestJson"`
	Config      string  `yaml:"configJson"`
	Kitfile     string  `yaml:"kitfileJson"`
	ErrRegexp   *string `yaml:"errRegexp"`
}

func (tc kitfileGenerateTestCase) withName(name string) kitfileGenerateTestCase {
	tc.Name = name
	return tc
}

func TestGenerateKitfileForModelPack(t *testing.T) {
	tests := loadAllTestCasesOrPanic[kitfileGenerateTestCase](t, filepath.Join("testdata", "kitfile-generation"))
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
			manifest := &ocispec.Manifest{}
			if err := json.Unmarshal([]byte(tt.Manifest), manifest); err != nil {
				t.Fatalf("Error unmarshaling test manifest: %s", err)
			}
			var config *modelspecv1.Model
			if tt.Config != "" {
				config = &modelspecv1.Model{}
				if err := json.Unmarshal([]byte(tt.Config), config); err != nil {
					t.Fatalf("Error unmarshaling test config: %s", err)
				}
			}

			actualKitfile, err := GenerateKitfileForModelPack(manifest, config)
			if tt.ErrRegexp == nil {
				if !assert.NoError(t, err) {
					return
				}
				expectedKitfile := &KitFile{}
				if err := json.Unmarshal([]byte(tt.Kitfile), expectedKitfile); err != nil {
					t.Fatalf("Error unmarshalling test Kitfile: %s", err)
				}
				assert.Equal(t, expectedKitfile, actualKitfile)
			} else {
				if !assert.Error(t, err) {
					return
				}
				assert.Regexp(t, *tt.ErrRegexp, err.Error())
			}
		})
	}
}

func TestParseReference(t *testing.T) {
	tests := []struct {
		input        string
		expectedRef  *registry.Reference
		expectedTags []string
		expectErr    bool
	}{
		{
			input:     "",
			expectErr: true,
		},
		{
			input:        "testregistry.io/test-organization/test-repository:test-tag",
			expectedRef:  reference("testregistry.io", "test-organization/test-repository", "test-tag"),
			expectedTags: []string{},
		},
		{
			input:        "testregistry.io/test-organization/test-repository:test-tag,extraTag1,extraTag2",
			expectedRef:  reference("testregistry.io", "test-organization/test-repository", "test-tag"),
			expectedTags: []string{"extraTag1", "extraTag2"},
		},
		{
			input:        "test-repository:test-tag,extraTag1,extraTag2",
			expectedRef:  reference(DefaultRegistry, "test-repository", "test-tag"),
			expectedTags: []string{"extraTag1", "extraTag2"},
		},
		{
			input:        "localhost:5000/test-organization/test-repository:test-tag,extraTag1,extraTag2",
			expectedRef:  reference("localhost:5000", "test-organization/test-repository", "test-tag"),
			expectedTags: []string{"extraTag1", "extraTag2"},
		},
		{
			input:        "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			expectedRef:  reference(DefaultRegistry, DefaultRepository, "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
			expectedTags: []string{},
		},
		{
			input:        "test-organization/test-repository:test-tag,extraTag1,extraTag2",
			expectedRef:  reference("localhost", "test-organization/test-repository", "test-tag"),
			expectedTags: []string{"extraTag1", "extraTag2"},
		},
		{
			input:        "a/b/c/d",
			expectedRef:  reference("localhost", "a/b/c/d", ""),
			expectedTags: []string{},
		},
		{
			input:        "test.io/a/b/c/d",
			expectedRef:  reference("test.io", "a/b/c/d", ""),
			expectedTags: []string{},
		},
		{
			input:        "testrepo@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			expectedRef:  reference(DefaultRegistry, "testrepo", "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
			expectedTags: []string{},
		},
		{
			input:        "testrepo:ignoredtag@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			expectedRef:  reference(DefaultRegistry, "testrepo", "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
			expectedTags: []string{},
		},
		{
			input:        "testorg/testrepo@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			expectedRef:  reference(DefaultRegistry, "testorg/testrepo", "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
			expectedTags: []string{},
		},
		{
			input:        "testorg.com/testrepo@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			expectedRef:  reference("testorg.com", "testrepo", "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"),
			expectedTags: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actualRef, actualTags, actualErr := ParseReference(tt.input)
			if tt.expectErr {
				assert.Error(t, actualErr)
				assert.Nil(t, actualRef)
				assert.Nil(t, actualTags)
			} else {
				if !assert.NoError(t, actualErr) {
					return
				}
				assert.Equal(t, tt.expectedRef, actualRef)
				assert.Equal(t, tt.expectedTags, actualTags)
			}
		})
	}
}

func reference(reg, repo, ref string) *registry.Reference {
	return &registry.Reference{
		Registry:   reg,
		Repository: repo,
		Reference:  ref,
	}
}
