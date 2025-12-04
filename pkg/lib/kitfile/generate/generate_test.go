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

package generate

import (
	"testing"
)

func TestDetermineFileType(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectedType fileType
	}{
		// Prompt files - should be recognized as code
		{
			name:         "prompt file without extension",
			filename:     "system.prompt",
			expectedType: fileTypeCode,
		},
		{
			name:         "prompt file with .md extension",
			filename:     "chain.prompt.md",
			expectedType: fileTypeCode,
		},
		{
			name:         "prompt file with .yaml extension",
			filename:     "my.prompt.yaml",
			expectedType: fileTypeCode,
		},
		{
			name:         "prompt file with .txt extension",
			filename:     "instruction.prompt.txt",
			expectedType: fileTypeCode,
		},
		{
			name:         "prompt file in subdirectory",
			filename:     "prompts/user.prompt",
			expectedType: fileTypeCode,
		},
		// Agent files - should be recognized as code
		{
			name:         "AGENTS.md file",
			filename:     "AGENTS.md",
			expectedType: fileTypeCode,
		},
		{
			name:         "agents.md lowercase",
			filename:     "agents.md",
			expectedType: fileTypeCode,
		},
		{
			name:         "SKILL.md file",
			filename:     "SKILL.md",
			expectedType: fileTypeCode,
		},
		{
			name:         "skill.md lowercase",
			filename:     "skill.md",
			expectedType: fileTypeCode,
		},
		{
			name:         "AGENTS.md in subdirectory",
			filename:     "docs/AGENTS.md",
			expectedType: fileTypeCode,
		},
		// Edge cases - should NOT be recognized as prompt/code
		{
			name:         "prompt without dot prefix (no leading dot)",
			filename:     "prompt.txt",
			expectedType: fileTypeMetadata, // .txt is in metadataSuffixes, doesn't match .prompt pattern
		},
		{
			name:         "prompt with underscore",
			filename:     "my_prompt.md",
			expectedType: fileTypeDocs, // .md suffix takes precedence
		},
		{
			name:         "file containing prompt in name",
			filename:     "prompter.py",
			expectedType: fileTypeUnknown,
		},
		// Regular files - should use existing logic
		{
			name:         "model file .gguf",
			filename:     "model.gguf",
			expectedType: fileTypeModel,
		},
		{
			name:         "dataset file .csv",
			filename:     "data.csv",
			expectedType: fileTypeDataset,
		},
		{
			name:         "docs file .md",
			filename:     "README.md",
			expectedType: fileTypeDocs,
		},
		{
			name:         "metadata file .json",
			filename:     "config.json",
			expectedType: fileTypeMetadata,
		},
		{
			name:         "unknown file .sh",
			filename:     "script.sh",
			expectedType: fileTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineFileType(tt.filename)
			if result != tt.expectedType {
				t.Errorf("determineFileType(%q) = %v, want %v", tt.filename, result, tt.expectedType)
			}
		})
	}
}
