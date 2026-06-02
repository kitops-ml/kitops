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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/stretchr/testify/assert"
)

type packUnpackTestcase struct {
	Name         string
	Description  string   `yaml:"description"`
	Kitfile      string   `yaml:"kitfile"`
	Kitignore    string   `yaml:"kitignore"`
	Files        []string `yaml:"files"`
	IgnoredFiles []string `yaml:"ignored"`
}

func (t packUnpackTestcase) withName(name string) packUnpackTestcase {
	t.Name = name
	return t
}

// TestPackUnpack tests kit functionality by generating a file tree, packing it,
// unpacking it, and verifying that the unpacked contents match expectations.
// We work in a new temporary directory for each test to avoid interaction between
// tests.
func TestPackUnpack(t *testing.T) {
	testPreflight(t)

	tests := loadAllTestCasesOrPanic[packUnpackTestcase](t, filepath.Join("testdata", "pack-unpack"))
	packArgCases := [][]string{
		{},                                // default case; tar format, no compression
		{"--compression", "gzip"},         // tar + gzip
		{"--compression", "gzip-fastest"}, // tar + gzip (fastest)
		{"--compression", "zstd"},         // tar + zstd
	}

	for _, args := range packArgCases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			for _, tt := range tests {
				t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
					// Set up temporary directory for work
					tmpDir := setupTempDir(t)

					// Set up paths to use for test
					modelKitPath, unpackPath, contextPath := setupTestDirs(t, tmpDir)
					t.Setenv(constants.KitopsHomeEnvVar, contextPath)

					// Create Kitfile
					setupKitfileAndKitignore(t, modelKitPath, tt.Kitfile, tt.Kitignore)
					// Create files for test case
					setupFiles(t, modelKitPath, append(tt.Files, tt.IgnoredFiles...))

					packArgs := []string{"pack", modelKitPath, "-t", modelKitTag}
					packArgs = append(packArgs, args...)
					runCommand(t, expectNoError, packArgs...)
					runCommand(t, expectNoError, "list")
					runCommand(t, expectNoError, "unpack", modelKitTag, "-d", unpackPath)

					checkFilesExistWithContent(t, unpackPath, tt.Files)
					checkFilesDoNotExist(t, unpackPath, append(tt.IgnoredFiles, ".kitignore"))
				})
			}
		})
	}
}

// TestPackUnpack tests kit raw-layer functionality by generating a file tree, packing it,
// unpacking it, and verifying that the unpacked contents match expectations.
// We work in a new temporary directory for each test to avoid interaction between
// tests. Since raw layers do not support directories, these tests skip kitignore support
// and use kit init to generate a kitfile for every file in the test case.
func TestPackUnpackRaw(t *testing.T) {
	testPreflight(t)

	tests := loadAllTestCasesOrPanic[packUnpackTestcase](t, filepath.Join("testdata", "pack-unpack"))
	packArgCases := [][]string{
		{},                                // raw, no compression
		{"--compression", "gzip"},         // raw + gzip
		{"--compression", "gzip-fastest"}, // raw + gzip (fastest)
		{"--compression", "zstd"},         // raw + zstd
	}

	for _, args := range packArgCases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			for _, tt := range tests {
				t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
					if slices.Contains(tt.Files, "./code/SKILL.md") {
						// Workaround: this test depends on kit init, which handles skills separately and always stores
						// them as directories, which is incompatible with raw layers
						t.Skip("Skipping raw handling for skills; only tar should be used here")
					}

					// Set up temporary directory for work
					tmpDir := setupTempDir(t)

					// Set up paths to use for test
					modelKitPath, unpackPath, contextPath := setupTestDirs(t, tmpDir)
					t.Setenv(constants.KitopsHomeEnvVar, contextPath)

					// Create files for test case
					setupFiles(t, modelKitPath, append(tt.Files, tt.IgnoredFiles...))
					// Generate a kitfile that lists every file individually
					runCommand(t, expectNoError, "init", "--depth", "-1", modelKitPath)

					packArgs := []string{"pack", modelKitPath, "-t", modelKitTag, "--layer-format", "raw"}
					packArgs = append(packArgs, args...)
					runCommand(t, expectNoError, packArgs...)
					runCommand(t, expectNoError, "list")
					runCommand(t, expectNoError, "unpack", modelKitTag, "-d", unpackPath)

					// For the raw case, .kitignore does not make sense (since each path has to be a separate file and
					// only those files are packed. To reuse the test cases here, lets just ensure all files are accounted
					// for.
					checkFilesExistWithContent(t, unpackPath, tt.Files)
					checkFilesExistWithContent(t, unpackPath, tt.IgnoredFiles)
				})
			}
		})
	}
}

func TestPackReproducibility(t *testing.T) {
	packArgCases := [][]string{
		{},                        // default case; tar format, no compression
		{"--layer-format", "raw"}, // raw format
		{"--compression", "gzip"}, // tar + gzip
		{"--layer-format", "raw", "--compression", "gzip"}, // raw + gzip
	}

	testKitfile := `
manifestVersion: 1.0.0
package:
  name: test-repack
model:
  path: test-file.txt
datasets:
  - path: test-dir/test-subfile.txt
`

	for _, args := range packArgCases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			tmpDir := setupTempDir(t)

			modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
			t.Setenv(constants.KitopsHomeEnvVar, contextPath)

			kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
			if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
				t.Fatal(err)
			}
			setupFiles(t, modelKitPath, []string{"test-file.txt", "test-dir/test-subfile.txt"})

			tag1, tag2 := "test:repack1", "test:repack2"

			pack1Args := []string{"pack", modelKitPath, "-t", tag1}
			pack1Args = append(pack1Args, args...)
			_ = runCommand(t, expectNoError, pack1Args...)

			// Change timestamps on file to simulate an unpacked modelkit at a future time
			futureTime := time.Now().Add(time.Hour)
			if err := os.Chtimes(filepath.Join(modelKitPath, "test-file.txt"), futureTime, futureTime); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(filepath.Join(modelKitPath, "test-dir"), futureTime, futureTime); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(filepath.Join(modelKitPath, "test-dir/test-subfile.txt"), futureTime, futureTime); err != nil {
				t.Fatal(err)
			}

			pack2Args := []string{"pack", modelKitPath, "-t", tag2}
			pack2Args = append(pack2Args, args...)
			pack2Out := runCommand(t, expectNoError, pack2Args...)

			// Overall ModelKit digests may not be the same due to a creation time annotation; nonetheless, layer digests _should_
			// be reproducible.
			assert.Contains(t, pack2Out, "Already saved model layer")
			assert.Contains(t, pack2Out, "Already saved dataset layer")
		})
	}
}
