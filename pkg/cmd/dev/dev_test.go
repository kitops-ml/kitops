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

package dev

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModelFile(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    []string
		testPath      string // If set, test direct file path instead of directory
		expectError   bool
		errorContains string
	}{
		{
			name:       "single gguf file in directory",
			setupFiles: []string{"model.gguf"},
		},
		{
			name:       "gguf file in subdirectory",
			setupFiles: []string{"models/llama-7b.gguf", "other-file.txt"},
		},
		{
			name:       "direct file path",
			setupFiles: []string{"direct-model.gguf"},
			testPath:   "direct-model.gguf",
		},
		{
			name:          "multiple gguf files",
			setupFiles:    []string{"model1.gguf", "model2.gguf"},
			expectError:   true,
			errorContains: "multiple model files found",
		},
		{
			name:          "no gguf files",
			setupFiles:    []string{"readme.txt", "config.json"},
			expectError:   true,
			errorContains: "could not find model file",
		},
		{
			name:          "nonexistent path",
			setupFiles:    []string{},
			testPath:      "nonexistent.gguf",
			expectError:   true,
			errorContains: "", // Don't check specific error message (platform dependent)
		},
		{
			name:          "directory without gguf files",
			setupFiles:    []string{"subdir/readme.txt"},
			testPath:      "subdir",
			expectError:   true,
			errorContains: "could not find model file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "find-model-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Setup files
			for _, file := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, file)
				require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
				require.NoError(t, os.WriteFile(filePath, []byte("test model data"), 0644))
			}

			// Determine test path
			var testPath string
			if tt.testPath != "" {
				if filepath.IsAbs(tt.testPath) {
					testPath = tt.testPath
				} else {
					testPath = filepath.Join(tmpDir, tt.testPath)
				}
			} else {
				testPath = tmpDir
			}

			result, err := findModelFile(testPath)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Contains(t, result, ".gguf", "Result should be a .gguf file")

			stat, err := os.Stat(result)
			require.NoError(t, err)
			assert.True(t, stat.Mode().IsRegular(), "Result should be a regular file")
		})
	}
}
