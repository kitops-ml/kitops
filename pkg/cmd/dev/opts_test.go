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
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry"
)

func TestDevStartOptions_Complete_ReferenceDetection(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		setupKitfile     bool
		expectModelRef   bool
		expectContextDir bool
	}{
		{
			name:           "directory path treated as reference",
			args:           []string{"./test-dir"},
			setupKitfile:   true,
			expectModelRef: false,
		},
		{
			name:           "simple modelkit reference",
			args:           []string{"myrepo/my-model:latest"},
			expectModelRef: true,
		},
		{
			name:           "registry url reference",
			args:           []string{"registry.example.com/models/llama:7b"},
			expectModelRef: true,
		},
		{
			name:           "localhost reference",
			args:           []string{"localhost:5000/model:tag"},
			expectModelRef: true,
		},
		{
			name:           "reference with port and path",
			args:           []string{"my-registry.com:8080/org/model:v1.0"},
			expectModelRef: true,
		},
		{
			name:             "no arguments defaults to current directory",
			args:             []string{},
			setupKitfile:     true,
			expectContextDir: false, // Empty contextDir for current directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directory
			tmpDir, err := os.MkdirTemp("", "dev-opts-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory for tests
			origWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origWd)
			require.NoError(t, os.Chdir(tmpDir))

			// Setup test directory with kitfile if needed
			if tt.setupKitfile {
				var testDir string
				if len(tt.args) > 0 && tt.args[0] != "" {
					testDir = filepath.Join(tmpDir, "test-dir")
					require.NoError(t, os.MkdirAll(testDir, 0755))
				} else {
					testDir = tmpDir // Current directory test
				}

				kitfilePath := filepath.Join(testDir, constants.DefaultKitfileName)
				kitfileContent := `manifestVersion: 1.0.0
package:
  name: test-model
model:
  path: model.gguf`
				require.NoError(t, os.WriteFile(kitfilePath, []byte(kitfileContent), 0644))
			}

			opts := &DevStartOptions{}
			configHome := filepath.Join(tmpDir, ".kitops")

			// Create context with config key as expected by the complete method
			ctx := context.WithValue(context.Background(), constants.ConfigKey{}, configHome)

			// Initialize NetworkOptions with defaults (as would happen in command setup)
			opts.Concurrency = 5
			opts.TLSVerify = true

			// Execute
			err = opts.complete(ctx, tt.args)
			require.NoError(t, err)

			// Verify detection logic
			if tt.expectModelRef {
				assert.NotNil(t, opts.modelRef, "Expected modelRef to be set for: %v", tt.args)
				assert.Equal(t, "", opts.contextDir, "Expected contextDir to be empty for ModelKit reference")
			}

			if tt.expectContextDir {
				assert.NotEmpty(t, opts.contextDir, "Expected contextDir to be set for directory path")
				assert.Nil(t, opts.modelRef, "Expected modelRef to be nil for directory path")
			}
		})
	}
}

func TestDevStartOptions_Complete_ErrorCases(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		setupKitfile        bool
		expectErrorContains string
	}{
		{
			name: "nonexistent directory treated as reference",
			args: []string{"./nonexistent"},
			// This will be treated as a ModelKit reference, not a directory
			// So it won't error in complete(), but would fail later during extraction
		},
		{
			name:                "no arguments with no kitfile in current dir",
			args:                []string{},
			expectErrorContains: "no directory or ModelKit reference provided",
		},
		{
			name:                "empty string argument",
			args:                []string{""},
			expectErrorContains: "no directory or ModelKit reference provided",
		},
		{
			name: "path with many parts treated as reference",
			args: []string{"not/a/valid/reference/with/too/many/parts"},
			// This gets parsed as a valid reference, no error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "dev-opts-error-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			origWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origWd)
			require.NoError(t, os.Chdir(tmpDir))

			opts := &DevStartOptions{}
			configHome := filepath.Join(tmpDir, ".kitops")

			// Create context with config key as expected by the complete method
			ctx := context.WithValue(context.Background(), constants.ConfigKey{}, configHome)

			// Initialize NetworkOptions with defaults (as would happen in command setup)
			opts.Concurrency = 5
			opts.TLSVerify = true

			err = opts.complete(ctx, tt.args)

			if tt.expectErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorContains)
			} else {
				// These cases don't error in complete(), they would fail later
				require.NoError(t, err)
			}
		})
	}
}

func TestDevStartOptions_Cleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-cleanup-test-*")
	require.NoError(t, err)

	// Create a test file in the temp directory
	testFile := filepath.Join(tmpDir, "test-model.gguf")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	opts := &DevStartOptions{
		contextDir: tmpDir,
		modelRef:   &registry.Reference{}, // Simulate ModelKit reference for cleanup
	}

	// Verify directory and file exist
	_, err = os.Stat(tmpDir)
	require.NoError(t, err)
	_, err = os.Stat(testFile)
	require.NoError(t, err)

	// Test cleanup
	err = opts.cleanup()
	assert.NoError(t, err)

	// Verify directory was removed
	_, err = os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err))
}
