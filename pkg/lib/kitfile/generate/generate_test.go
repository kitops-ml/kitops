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

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/stretchr/testify/assert"
)

func TestDetermineFileType(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectedType fileType
	}{
		// Prompt files - should be recognized as prompts
		{
			name:         "prompt file without extension",
			filename:     "system.prompt",
			expectedType: fileTypePrompt,
		},
		{
			name:         "prompt file with .md extension",
			filename:     "chain.prompt.md",
			expectedType: fileTypePrompt,
		},
		{
			name:         "prompt file with .yaml extension",
			filename:     "my.prompt.yaml",
			expectedType: fileTypePrompt,
		},
		{
			name:         "prompt file with .txt extension",
			filename:     "instruction.prompt.txt",
			expectedType: fileTypePrompt,
		},
		{
			name:         "prompt file in subdirectory",
			filename:     "prompts/user.prompt",
			expectedType: fileTypePrompt,
		},
		// Agent files - should be recognized as prompts
		{
			name:         "AGENTS.md file",
			filename:     "AGENTS.md",
			expectedType: fileTypePrompt,
		},
		{
			name:         "agents.md lowercase",
			filename:     "agents.md",
			expectedType: fileTypePrompt,
		},
		{
			name:         "SKILL.md file",
			filename:     "SKILL.md",
			expectedType: fileTypePrompt,
		},
		{
			name:         "skill.md lowercase",
			filename:     "skill.md",
			expectedType: fileTypePrompt,
		},
		{
			name:         "AGENTS.md in subdirectory",
			filename:     "docs/AGENTS.md",
			expectedType: fileTypePrompt,
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

func TestGenerateWithDepth(t *testing.T) {
	testDirListing := DirectoryListing{
		Name: "root-dir",
		Path: "root-dir",
		Files: []FileListing{
			{Name: "root-one", Path: "root-one.md", Size: 100},
			{Name: "root-two", Path: "root-two.md", Size: 100},
			{Name: "root-three", Path: "root-three.md", Size: 100},
		},
		Subdirs: []DirectoryListing{
			{
				Name: "subdir-one",
				Path: "subdir-one",
				Files: []FileListing{
					{Name: "subdir-one-one", Path: "subdir-one/subdir-one-one.md", Size: 100},
					{Name: "subdir-one-two", Path: "subdir-one/subdir-one-two.md", Size: 100},
					{Name: "subdir-one-three", Path: "subdir-one/subdir-one-three.md", Size: 100},
				},
				Subdirs: []DirectoryListing{
					{
						Name: "subdir-two",
						Path: "subdir-one/subdir-two",
						Files: []FileListing{
							{Name: "subdir-two-one", Path: "subdir-one/subdir-two/subdir-two-one.md", Size: 100},
							{Name: "subdir-two-two", Path: "subdir-one/subdir-two/subdir-two-two.md", Size: 100},
							{Name: "subdir-two-three", Path: "subdir-one/subdir-two/subdir-two-three.md", Size: 100},
						},
						Subdirs: []DirectoryListing{
							{
								Name: "subdir-three",
								Path: "subdir-one/subdir-two/subdir-three",
								Files: []FileListing{
									{Name: "subdir-three-one", Path: "subdir-one/subdir-two/subdir-three/subdir-three-one.md", Size: 100},
									{Name: "subdir-three-two", Path: "subdir-one/subdir-two/subdir-three/subdir-three-two.md", Size: 100},
									{Name: "subdir-three-three", Path: "subdir-one/subdir-two/subdir-three/subdir-three-three.md", Size: 100},
								},
							},
						},
					},
				},
			},
		},
	}

	t.Run("Depth zero Kitfile", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 0)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "root-one.md"},
			{Path: "root-two.md"},
			{Path: "root-three.md"},
		}
		for _, expectedDoc := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expectedDoc)
		}
		assert.Contains(t, kitfile.Code, artifact.Code{Path: "subdir-one/"})
	})

	t.Run("Depth one Kitfile", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 1)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "root-one.md"},
			{Path: "root-two.md"},
			{Path: "root-three.md"},
			{Path: "subdir-one/subdir-one-one.md"},
			{Path: "subdir-one/subdir-one-two.md"},
			{Path: "subdir-one/subdir-one-three.md"},
		}
		for _, expectedDoc := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expectedDoc)
		}
		assert.Contains(t, kitfile.Code, artifact.Code{Path: "subdir-one/subdir-two/"})
	})

	t.Run("Depth two Kitfile", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 2)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "root-one.md"},
			{Path: "root-two.md"},
			{Path: "root-three.md"},
			{Path: "subdir-one/subdir-one-one.md"},
			{Path: "subdir-one/subdir-one-two.md"},
			{Path: "subdir-one/subdir-one-three.md"},
			{Path: "subdir-one/subdir-two/subdir-two-one.md"},
			{Path: "subdir-one/subdir-two/subdir-two-two.md"},
			{Path: "subdir-one/subdir-two/subdir-two-three.md"},
		}
		for _, expectedDoc := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expectedDoc)
		}
		assert.Contains(t, kitfile.Code, artifact.Code{Path: "subdir-one/subdir-two/subdir-three/"})
	})

	t.Run("Depth three Kitfile", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 3)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "root-one.md"},
			{Path: "root-two.md"},
			{Path: "root-three.md"},
			{Path: "subdir-one/subdir-one-one.md"},
			{Path: "subdir-one/subdir-one-two.md"},
			{Path: "subdir-one/subdir-one-three.md"},
			{Path: "subdir-one/subdir-two/subdir-two-one.md"},
			{Path: "subdir-one/subdir-two/subdir-two-two.md"},
			{Path: "subdir-one/subdir-two/subdir-two-three.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-one.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-two.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-three.md"},
		}
		for _, expectedDoc := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expectedDoc)
		}
		assert.Len(t, kitfile.Code, 0, "Should not contain any code layers")
	})

	t.Run("Full depth Kitfile", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, -1)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "root-one.md"},
			{Path: "root-two.md"},
			{Path: "root-three.md"},
			{Path: "subdir-one/subdir-one-one.md"},
			{Path: "subdir-one/subdir-one-two.md"},
			{Path: "subdir-one/subdir-one-three.md"},
			{Path: "subdir-one/subdir-two/subdir-two-one.md"},
			{Path: "subdir-one/subdir-two/subdir-two-two.md"},
			{Path: "subdir-one/subdir-two/subdir-two-three.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-one.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-two.md"},
			{Path: "subdir-one/subdir-two/subdir-three/subdir-three-three.md"},
		}
		for _, expectedDoc := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expectedDoc)
		}
		assert.Len(t, kitfile.Code, 0, "Should not contain any code layers")
	})
}

func TestGenerateReadmeAndLicenseInSubdirs(t *testing.T) {
	testDirListing := DirectoryListing{
		Name: "root-dir",
		Path: "root-dir",
		Files: []FileListing{
			{Name: "README.md", Path: "README.md", Size: 100},
			{Name: "LICENSE", Path: "LICENSE", Size: 100},
		},
		Subdirs: []DirectoryListing{
			{
				Name: "subdir-one",
				Path: "subdir-one",
				Files: []FileListing{
					{Name: "README.md", Path: "subdir-one/README.md", Size: 100},
					{Name: "LICENSE", Path: "subdir-one/LICENSE", Size: 100},
				},
				Subdirs: []DirectoryListing{
					{
						Name: "subdir-two",
						Path: "subdir-one/subdir-two",
						Files: []FileListing{
							{Name: "README.md", Path: "subdir-one/subdir-two/README.md", Size: 100},
							{Name: "LICENSE", Path: "subdir-one/subdir-two/LICENSE", Size: 100},
						},
					},
				},
			},
		},
	}

	t.Run("Depth zero only includes root README and LICENSE", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 0)
		assert.NoError(t, err)
		assert.Len(t, kitfile.Docs, 2)
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "README.md", Description: "Readme file"})
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "LICENSE", Description: "License file"})
	})

	t.Run("Depth one includes subdir README and LICENSE with full path", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, 1)
		assert.NoError(t, err)
		assert.Len(t, kitfile.Docs, 4)
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "README.md", Description: "Readme file"})
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "LICENSE", Description: "License file"})
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "subdir-one/README.md", Description: "Readme file"})
		assert.Contains(t, kitfile.Docs, artifact.Docs{Path: "subdir-one/LICENSE", Description: "License file"})
	})

	t.Run("Full depth includes all README and LICENSE files with full paths", func(t *testing.T) {
		kitfile, err := GenerateKitfile(&testDirListing, nil, -1)
		assert.NoError(t, err)
		expectedDocs := []artifact.Docs{
			{Path: "README.md", Description: "Readme file"},
			{Path: "LICENSE", Description: "License file"},
			{Path: "subdir-one/README.md", Description: "Readme file"},
			{Path: "subdir-one/LICENSE", Description: "License file"},
			{Path: "subdir-one/subdir-two/README.md", Description: "Readme file"},
			{Path: "subdir-one/subdir-two/LICENSE", Description: "License file"},
		}
		for _, expected := range expectedDocs {
			assert.Contains(t, kitfile.Docs, expected)
		}
	})
}

func TestGenerateModelPartsDepth(t *testing.T) {
	testDirListing := DirectoryListing{
		Name: "root-dir",
		Path: "root-dir",
		Files: []FileListing{
			{Name: "root-one", Path: "root-one.onnx", Size: 100},
			{Name: "root-two", Path: "root-two.onnx", Size: 100},
			{Name: "root-meta", Path: "root-meta.json", Size: 100},
		},
		Subdirs: []DirectoryListing{
			{
				Name: "subdir-one",
				Path: "subdir-one",
				Files: []FileListing{
					{Name: "subdir-one-one", Path: "subdir-one/subdir-one-one.onnx", Size: 100},
					{Name: "subdir-one-two", Path: "subdir-one/subdir-one-two.onnx", Size: 100},
					{Name: "subdir-one-meta", Path: "subdir-one/subdir-one-meta.json", Size: 100},
				},
				Subdirs: []DirectoryListing{
					{
						Name: "subdir-two",
						Path: "subdir-one/subdir-two",
						Files: []FileListing{
							{Name: "subdir-two-one", Path: "subdir-one/subdir-two/subdir-two-one.onnx", Size: 100},
							{Name: "subdir-two-two", Path: "subdir-one/subdir-two/subdir-two-two.onnx", Size: 100},
							{Name: "subdir-two-meta", Path: "subdir-one/subdir-two/subdir-two-meta.json", Size: 100},
						},
						Subdirs: []DirectoryListing{
							{
								Name: "subdir-three",
								Path: "subdir-one/subdir-two/subdir-three",
								Files: []FileListing{
									{Name: "subdir-three-one", Path: "subdir-one/subdir-two/subdir-three/subdir-three-one.onnx", Size: 100},
									{Name: "subdir-three-two", Path: "subdir-one/subdir-two/subdir-three/subdir-three-two.onnx", Size: 100},
									{Name: "subdir-three-meta", Path: "subdir-one/subdir-two/subdir-three/subdir-three-meta.json", Size: 100},
								},
							},
						},
					},
				},
			},
		},
	}
	kitfile, err := GenerateKitfile(&testDirListing, nil, -1)
	assert.NoError(t, err)
	assert.Equal(t, "root-one.onnx", kitfile.Model.Path)
	assert.Len(t, kitfile.Model.Parts, 11)
	expectedModelParts := []artifact.ModelPart{
		{Path: "root-two.onnx"},
		{Path: "subdir-one/subdir-one-one.onnx"},
		{Path: "subdir-one/subdir-one-two.onnx"},
		{Path: "subdir-one/subdir-two/subdir-two-one.onnx"},
		{Path: "subdir-one/subdir-two/subdir-two-two.onnx"},
		{Path: "subdir-one/subdir-two/subdir-three/subdir-three-one.onnx"},
		{Path: "subdir-one/subdir-two/subdir-three/subdir-three-two.onnx"},
		{Path: "root-meta.json"},
		{Path: "subdir-one/subdir-one-meta.json"},
		{Path: "subdir-one/subdir-two/subdir-two-meta.json"},
		{Path: "subdir-one/subdir-two/subdir-three/subdir-three-meta.json"},
	}
	for _, expectedModelPart := range expectedModelParts {
		assert.Contains(t, kitfile.Model.Parts, expectedModelPart)
	}
}
