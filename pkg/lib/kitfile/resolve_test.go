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

package kitfile

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/stretchr/testify/assert"
)

func TestMergeKitfiles_Prompts(t *testing.T) {
	prompt1 := artifact.Prompt{Name: "prompt1", Path: "prompts/p1.txt", Description: "first prompt"}
	prompt2 := artifact.Prompt{Name: "prompt2", Path: "prompts/p2.txt", Description: "second prompt"}
	prompt3 := artifact.Prompt{Name: "prompt3", Path: "prompts/p3.txt", Description: "third prompt"}

	tests := []struct {
		name     string
		into     *artifact.KitFile
		from     *artifact.KitFile
		expected []artifact.Prompt
	}{
		{
			name:     "both nil prompts",
			into:     &artifact.KitFile{},
			from:     &artifact.KitFile{},
			expected: nil,
		},
		{
			name:     "into has prompts, from has none",
			into:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt1, prompt2}},
			from:     &artifact.KitFile{},
			expected: []artifact.Prompt{prompt1, prompt2},
		},
		{
			name:     "from has prompts, into has none",
			into:     &artifact.KitFile{},
			from:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt2, prompt3}},
			expected: []artifact.Prompt{prompt2, prompt3},
		},
		{
			name:     "both have prompts — merged in order",
			into:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt1}},
			from:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt2, prompt3}},
			expected: []artifact.Prompt{prompt1, prompt2, prompt3},
		},
		{
			name:     "overlapping prompts — all preserved",
			into:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt1, prompt2}},
			from:     &artifact.KitFile{Prompts: []artifact.Prompt{prompt2, prompt3}},
			expected: []artifact.Prompt{prompt1, prompt2, prompt2, prompt3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeKitfiles(tt.into, tt.from)
			assert.Equal(t, tt.expected, result.Prompts, "Prompts mismatch")
		})
	}
}

func TestMergeKitfiles_PreservesOtherFields(t *testing.T) {
	into := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Package: artifact.Package{
			Name: "into-pkg",
		},
		Code:     []artifact.Code{{Path: "into/code.py"}},
		DataSets: []artifact.DataSet{{Path: "into/data.csv"}},
		Docs:     []artifact.Docs{{Path: "into/doc.md"}},
		Prompts:  []artifact.Prompt{{Name: "into-prompt", Path: "into/prompt.txt"}},
	}
	from := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Package: artifact.Package{
			Name: "from-pkg",
		},
		Code:     []artifact.Code{{Path: "from/code.py"}},
		DataSets: []artifact.DataSet{{Path: "from/data.csv"}},
		Docs:     []artifact.Docs{{Path: "from/doc.md"}},
		Prompts:  []artifact.Prompt{{Name: "from-prompt", Path: "from/prompt.txt"}},
	}

	result := mergeKitfiles(into, from)

	assert.Len(t, result.Prompts, 2, "should have prompts from both")
	assert.Equal(t, "into-prompt", result.Prompts[0].Name, "into prompts should come first")
	assert.Equal(t, "from-prompt", result.Prompts[1].Name, "from prompts should come second")

	assert.Len(t, result.Code, 2, "code should still merge")
	assert.Len(t, result.DataSets, 2, "datasets should still merge")
	assert.Len(t, result.Docs, 2, "docs should still merge")
	assert.Equal(t, "into-pkg", result.Package.Name, "package name should use firstNonEmpty")
}

func TestMergeKitfiles_DoesNotMutateInput(t *testing.T) {
	intoPrompts := []artifact.Prompt{{Name: "p1"}, {Name: "p2"}}
	fromPrompts := []artifact.Prompt{{Name: "p3"}}

	into := &artifact.KitFile{Prompts: intoPrompts}
	from := &artifact.KitFile{Prompts: fromPrompts}

	result := mergeKitfiles(into, from)

	assert.Len(t, result.Prompts, 3, "result should have 3 prompts")
	assert.Len(t, into.Prompts, 2, "input 'into' should be unmodified")
	assert.Len(t, from.Prompts, 1, "input 'from' should be unmodified")
}
