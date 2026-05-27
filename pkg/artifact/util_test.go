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

func TestKitfileHasLayerInfo(t *testing.T) {
	full := func() *LayerInfo { return &LayerInfo{Digest: "sha256:aaa", DiffId: "sha256:bbb"} }
	digestOnly := func() *LayerInfo { return &LayerInfo{Digest: "sha256:aaa"} }
	diffIDOnly := func() *LayerInfo { return &LayerInfo{DiffId: "sha256:bbb"} }
	emptyInfo := func() *LayerInfo { return &LayerInfo{} }

	tests := []struct {
		name       string
		kitfile    *KitFile
		wantDigest bool
		wantDiffID bool
		wantErr    bool
		errRegexp  string
	}{
		{
			name:       "empty kitfile has no layers (vacuous true)",
			kitfile:    &KitFile{},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "model with nil LayerInfo",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin"},
			},
			wantDigest: false,
			wantDiffID: false,
		},
		{
			name: "model with empty LayerInfo struct",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: emptyInfo()},
			},
			wantDigest: false,
			wantDiffID: false,
		},
		{
			name: "model with full LayerInfo",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: full()},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "model with digest only",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: digestOnly()},
			},
			wantDigest: true,
			wantDiffID: false,
		},
		{
			name: "model with diffID only",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: diffIDOnly()},
			},
			wantDigest: false,
			wantDiffID: true,
		},
		{
			name: "all layer types with full LayerInfo",
			kitfile: &KitFile{
				Model: &Model{
					Path:      "model.bin",
					LayerInfo: full(),
					Parts: []ModelPart{
						{Path: "part1.bin", LayerInfo: full()},
						{Path: "part2.bin", LayerInfo: full()},
					},
				},
				Code:     []Code{{Path: "code", LayerInfo: full()}},
				DataSets: []DataSet{{Path: "data", LayerInfo: full()}},
				Docs:     []Docs{{Path: "docs", LayerInfo: full()}},
				Prompts:  []Prompt{{Path: "prompt", LayerInfo: full()}},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "all layer types with digest only",
			kitfile: &KitFile{
				Model: &Model{
					Path:      "model.bin",
					LayerInfo: digestOnly(),
					Parts:     []ModelPart{{Path: "part1.bin", LayerInfo: digestOnly()}},
				},
				Code:     []Code{{Path: "code", LayerInfo: digestOnly()}},
				DataSets: []DataSet{{Path: "data", LayerInfo: digestOnly()}},
				Docs:     []Docs{{Path: "docs", LayerInfo: digestOnly()}},
				Prompts:  []Prompt{{Path: "prompt", LayerInfo: digestOnly()}},
			},
			wantDigest: true,
			wantDiffID: false,
		},
		{
			name: "all layer types with no LayerInfo",
			kitfile: &KitFile{
				Model: &Model{
					Path:  "model.bin",
					Parts: []ModelPart{{Path: "part1.bin"}},
				},
				Code:     []Code{{Path: "code"}},
				DataSets: []DataSet{{Path: "data"}},
				Docs:     []Docs{{Path: "docs"}},
				Prompts:  []Prompt{{Path: "prompt"}},
			},
			wantDigest: false,
			wantDiffID: false,
		},
		{
			name: "model has digest but model part does not",
			kitfile: &KitFile{
				Model: &Model{
					Path:      "model.bin",
					LayerInfo: digestOnly(),
					Parts:     []ModelPart{{Path: "part1.bin"}},
				},
			},
			wantErr:   true,
			errRegexp: "invalid digests",
		},
		{
			name: "model has full info but a dataset has only diffID",
			kitfile: &KitFile{
				Model:    &Model{Path: "model.bin", LayerInfo: full()},
				DataSets: []DataSet{{Path: "data", LayerInfo: diffIDOnly()}},
			},
			wantErr:   true,
			errRegexp: "invalid digests",
		},
		{
			name: "model has full info but a code layer has only digest",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: full()},
				Code:  []Code{{Path: "code", LayerInfo: digestOnly()}},
			},
			wantErr:   true,
			errRegexp: "invalid diffIDs",
		},
		{
			name: "docs layer missing LayerInfo while others have it",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: full()},
				Docs:  []Docs{{Path: "docs"}},
			},
			wantErr:   true,
			errRegexp: "invalid digests",
		},
		{
			name: "prompt layer missing LayerInfo while others have it",
			kitfile: &KitFile{
				Model:   &Model{Path: "model.bin", LayerInfo: full()},
				Prompts: []Prompt{{Path: "prompt"}},
			},
			wantErr:   true,
			errRegexp: "invalid digests",
		},
		{
			name: "nil model with consistent other layers",
			kitfile: &KitFile{
				Code:     []Code{{Path: "code", LayerInfo: full()}},
				DataSets: []DataSet{{Path: "data", LayerInfo: full()}},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "model parts inconsistent with each other",
			kitfile: &KitFile{
				Model: &Model{
					Path:      "model.bin",
					LayerInfo: full(),
					Parts: []ModelPart{
						{Path: "part1.bin", LayerInfo: full()},
						{Path: "part2.bin", LayerInfo: digestOnly()},
					},
				},
			},
			wantErr:   true,
			errRegexp: "invalid diffIDs",
		},
		{
			// Remote model references don't carry their own layer (they're resolved
			// at unpack time) so the missing LayerInfo must not be treated as a gap.
			name: "model is remote ModelKit reference with no LayerInfo",
			kitfile: &KitFile{
				Model: &Model{Path: "registry.io/myorg/model:v1"},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "remote ModelKit model alongside local layers with full LayerInfo",
			kitfile: &KitFile{
				Model:    &Model{Path: "registry.io/myorg/model:v1"},
				Code:     []Code{{Path: "code", LayerInfo: full()}},
				DataSets: []DataSet{{Path: "data", LayerInfo: full()}},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			name: "remote ModelKit model with parts that have full LayerInfo",
			kitfile: &KitFile{
				Model: &Model{
					Path: "registry.io/myorg/model:v1",
					Parts: []ModelPart{
						{Path: "part1.bin", LayerInfo: full()},
					},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			// S3-backed datasets are downloaded at unpack time; the dataset entry
			// itself has no layer in the manifest, so missing LayerInfo is expected.
			name: "dataset with S3 RemotePath and no LayerInfo",
			kitfile: &KitFile{
				DataSets: []DataSet{
					{Path: "data", RemotePath: "s3://bucket/key", RemoteHash: "sha256:abc"},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			// ModelKit-referenced datasets are resolved at unpack time, same as above.
			name: "dataset with ModelKit RemotePath and no LayerInfo",
			kitfile: &KitFile{
				DataSets: []DataSet{
					{Path: "data", RemotePath: "registry.io/myorg/dataset:v1"},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			// Local model with full LayerInfo coexists with remote datasets that
			// (correctly) carry no LayerInfo — the remotes are skipped, so the
			// consistency check sees only the model and succeeds.
			name: "local model with remote S3 and ModelKit datasets",
			kitfile: &KitFile{
				Model: &Model{Path: "model.bin", LayerInfo: full()},
				DataSets: []DataSet{
					{Path: "s3data", RemotePath: "s3://bucket/key", RemoteHash: "sha256:abc"},
					{Path: "refdata", RemotePath: "registry.io/myorg/dataset:v1"},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			// Mixed: one local dataset (with LayerInfo) plus a remote dataset (no
			// LayerInfo). The remote one is skipped; the local one is consistent.
			name: "local dataset with LayerInfo plus remote dataset without",
			kitfile: &KitFile{
				DataSets: []DataSet{
					{Path: "local", LayerInfo: full()},
					{Path: "remote", RemotePath: "s3://bucket/key", RemoteHash: "sha256:abc"},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
		{
			// Fully remote Kitfile: model is a reference, all datasets remote, no
			// other layers. Nothing is added to the consistency check, so the
			// vacuous-true result is returned.
			name: "fully remote Kitfile yields vacuous true",
			kitfile: &KitFile{
				Model: &Model{Path: "registry.io/myorg/model:v1"},
				DataSets: []DataSet{
					{Path: "data", RemotePath: "s3://bucket/key", RemoteHash: "sha256:abc"},
				},
			},
			wantDigest: true,
			wantDiffID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDigest, gotDiffID, err := KitfileHasLayerInfo(tt.kitfile)
			if tt.wantErr {
				if !assert.Error(t, err) {
					return
				}
				assert.Regexp(t, tt.errRegexp, err.Error())
				assert.False(t, gotDigest)
				assert.False(t, gotDiffID)
				return
			}
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.wantDigest, gotDigest, "hasDigest")
			assert.Equal(t, tt.wantDiffID, gotDiffID, "hasDiffID")
		})
	}
}
