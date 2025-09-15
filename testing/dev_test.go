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

package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/stretchr/testify/assert"
)

type devTestcase struct {
	Name        string
	Description string `yaml:"description"`
	Setup       struct {
		ModelKits []struct {
			Tag     string   `yaml:"tag"`
			Kitfile string   `yaml:"kitfile"`
			Files   []string `yaml:"files"`
		} `yaml:"modelkits"`
	} `yaml:"setup"`
	Tests []struct {
		Name             string   `yaml:"name"`
		Args             string   `yaml:"args"`
		ExpectStartError bool     `yaml:"expectStartError"`
		ExpectedCacheDir bool     `yaml:"expectedCacheDir"`
		ExpectedCleanup  bool     `yaml:"expectedCleanup"`
		StartErrorRegexp string   `yaml:"startErrorRegexp"`
		ExpectedFiles    []string `yaml:"expectedFiles"`
	} `yaml:"tests"`
}

func (t devTestcase) withName(name string) devTestcase {
	t.Name = name
	return t
}

// assertContainsRegexp checks if output matches the given regex pattern
func assertContainsRegexp(t *testing.T, output, pattern string, shouldContain bool) {
	matched, err := regexp.MatchString(pattern, output)
	if err != nil {
		t.Fatalf("Invalid regex pattern %s: %v", pattern, err)
	}
	if shouldContain && !matched {
		t.Fatalf("Output should match regexp %s, but got:\n%s", pattern, output)
	}
	if !shouldContain && matched {
		t.Fatalf("Output should not match regexp %s, but got:\n%s", pattern, output)
	}
}

func TestDevCommand(t *testing.T) {
	testPreflight(t)

	tests := loadAllTestCasesOrPanic[devTestcase](t, filepath.Join("testdata", "dev"))
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
			// Set up temporary directory for work
			tmpDir := setupTempDir(t)
			contextPath := filepath.Join(tmpDir, ".kitops")
			t.Setenv(constants.KitopsHomeEnvVar, contextPath)

			// Pack ModelKits for testing
			for i, modelkit := range tt.Setup.ModelKits {
				modelKitDir := filepath.Join(tmpDir, fmt.Sprintf("modelkit-%d", i))
				if err := os.MkdirAll(modelKitDir, 0755); err != nil {
					t.Fatal(err)
				}

				setupKitfileAndKitignore(t, modelKitDir, modelkit.Kitfile, "")
				setupFiles(t, modelKitDir, modelkit.Files)
				runCommand(t, expectNoError, "pack", modelKitDir, "-t", modelkit.Tag)
			}

			// Run individual test cases
			for _, testCase := range tt.Tests {
				t.Run(testCase.Name, func(t *testing.T) {
					cacheDir := filepath.Join(contextPath, "dev-models", "current")

					// Clear cache directory and stop any running dev server before each test
					os.RemoveAll(cacheDir)
					_ = runCommand(t, expectError, "dev", "stop") // Ignore errors

					// Parse arguments
					var args []string
					if testCase.Args != "" {
						args = strings.Fields(testCase.Args)
					}
					fullArgs := append([]string{"dev", "start"}, args...)

					if testCase.ExpectStartError {
						output := runCommand(t, expectError, fullArgs...)
						if testCase.StartErrorRegexp != "" {
							assertContainsRegexp(t, output, testCase.StartErrorRegexp, true)
						}

						// For reference-based tests, check if cache directory was created
						if testCase.ExpectedCacheDir {
							// Cache directory should exist even if dev start failed
							// (because extraction happens before harness failure)
							_, err := os.Stat(cacheDir)
							if !assert.NoError(t, err, "Cache directory should be created during extraction") {
								return
							}

							// Check if expected files were extracted
							if len(testCase.ExpectedFiles) > 0 {
								checkFilesExist(t, cacheDir, testCase.ExpectedFiles)
							}
						}
					} else {
						// Test successful start - run in background and then stop
						output := runCommand(t, expectNoError, fullArgs...)

						// Log successful start
						t.Logf("Dev server started successfully: %s", output)

						// Check cache directory if expected
						if testCase.ExpectedCacheDir {
							_, err := os.Stat(cacheDir)
							assert.NoError(t, err, "Cache directory should be created during extraction")

							if len(testCase.ExpectedFiles) > 0 {
								checkFilesExist(t, cacheDir, testCase.ExpectedFiles)
							}
						}
					}

					// Test cleanup if expected
					if testCase.ExpectedCleanup {
						// For successful starts, dev stop should succeed and clean up cache
						// For failed starts, dev stop will fail but we don't test cleanup in that case
						if !testCase.ExpectStartError {
							runCommand(t, expectNoError, "dev", "stop") // Should succeed since server is running

							// Verify cache directory was cleaned up
							_, err := os.Stat(cacheDir)
							assert.True(t, os.IsNotExist(err), "Cache directory should be cleaned up after successful dev stop")
						}
					}
				})
			}
		})
	}
}

// TestDevDirectory tests the original directory-based dev functionality
func TestDevDirectory(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	// Create a test model file
	kitfileContent := `manifestVersion: 1.0.0
package:
  name: test-dev-model
model:
  path: test-model.gguf`

	setupKitfileAndKitignore(t, modelKitPath, kitfileContent, "")
	setupFiles(t, modelKitPath, []string{"test-model.gguf"})

	// Test that dev start works with directory paths (original functionality)
	output := runCommand(t, expectNoError, "dev", "start", modelKitPath)

	// Should succeed completely
	assert.Contains(t, output, "Development server started")
	assert.NotContains(t, output, "failed to find Kitfile")
	assert.NotContains(t, output, "could not find model file")

	// Clean up the running server
	runCommand(t, expectNoError, "dev", "stop")
}

// TestDevCleanup tests that dev stop properly cleans up resources
func TestDevCleanup(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	contextPath := filepath.Join(tmpDir, ".kitops")
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	// Create a fake cache directory structure
	cacheDir := filepath.Join(contextPath, "dev-models", "current")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files in cache
	testFiles := []string{"test-model.gguf", "Kitfile"}
	for _, file := range testFiles {
		testFile := filepath.Join(cacheDir, file)
		if err := os.WriteFile(testFile, []byte("test model data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Verify cache directory and files exist
	_, err := os.Stat(cacheDir)
	assert.NoError(t, err, "Cache directory should exist before cleanup")

	for _, file := range testFiles {
		_, err := os.Stat(filepath.Join(cacheDir, file))
		assert.NoError(t, err, "Test file %s should exist before cleanup", file)
	}

	// Run dev stop (will fail due to no running server), but should still clean up cache
	runCommand(t, expectError, "dev", "stop")
	_, err = os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err), "Cache directory should be removed by dev stop")
}

// TestDevErrorScenarios tests various error conditions
func TestDevErrorScenarios(t *testing.T) {
	testPreflight(t)

	tests := []struct {
		name           string
		args           []string
		expectedRegexp string
	}{
		{
			name:           "no arguments and no kitfile",
			args:           []string{"dev", "start"},
			expectedRegexp: "no Kitfile found in directory",
		},
		{
			name:           "nonexistent modelkit reference",
			args:           []string{"dev", "start", "nonexistent/model:tag"},
			expectedRegexp: "failed to extract ModelKit.*not found",
		},
		{
			name:           "empty argument",
			args:           []string{"dev", "start", ""},
			expectedRegexp: "no Kitfile found in directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTempDir(t)
			contextPath := filepath.Join(tmpDir, ".kitops")
			t.Setenv(constants.KitopsHomeEnvVar, contextPath)

			output := runCommand(t, expectError, tt.args...)
			assertContainsRegexp(t, output, tt.expectedRegexp, true) // Use our function with shouldContain=true
		})
	}
}

// TestDevReferenceExtraction tests that ModelKit references are properly extracted
func TestDevReferenceExtraction(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	contextPath := filepath.Join(tmpDir, ".kitops")
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	// Create a test ModelKit
	modelKitDir := filepath.Join(tmpDir, "test-modelkit")
	if err := os.MkdirAll(modelKitDir, 0755); err != nil {
		t.Fatal(err)
	}

	kitfileContent := `manifestVersion: 1.0.0
package:
  name: test-extraction-model
model:
  path: models/test-model.gguf`

	setupKitfileAndKitignore(t, modelKitDir, kitfileContent, "")
	setupFiles(t, modelKitDir, []string{"models/test-model.gguf"})

	// Pack the ModelKit
	runCommand(t, expectNoError, "pack", modelKitDir, "-t", "test/extraction:latest")

	// Try to start dev server with ModelKit reference
	// This should extract the ModelKit to cache directory and start successfully
	cacheDir := filepath.Join(contextPath, "dev-models", "current")

	output := runCommand(t, expectNoError, "dev", "start", "test/extraction:latest")

	// Should succeed - server should start
	assert.Contains(t, output, "Development server started")
	assert.NotContains(t, output, "failed to extract ModelKit")
	assert.NotContains(t, output, "not found")

	// Cache directory should be created
	_, err := os.Stat(cacheDir)
	assert.NoError(t, err, "Cache directory should be created during extraction")

	// Expected files should be extracted
	expectedFiles := []string{"Kitfile", "models/test-model.gguf"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(cacheDir, file)
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "File %s should be extracted to cache", file)
	}

	// Clean up - should succeed since server is running
	runCommand(t, expectNoError, "dev", "stop")

	// Verify cleanup - should work since stop was successful
	_, err = os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err), "Cache directory should be cleaned up")
}
