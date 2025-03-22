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

			runCommand(t, expectNoError, "pack", modelKitPath, "-t", modelKitTag)
			runCommand(t, expectNoError, "list")
			runCommand(t, expectNoError, "unpack", modelKitTag, "-d", unpackPath)

			checkFilesExist(t, unpackPath, tt.Files)
			checkFilesDoNotExist(t, unpackPath, append(tt.IgnoredFiles, ".kitignore"))
		})
	}
}

func TestPackReproducibility(t *testing.T) {
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	testKitfile := `
manifestVersion: 1.0.0
package:
  name: test-repack
model:
  path: test-file.txt
datasets:
  - path: test-dir/test-subfile.txt
`
	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}
	setupFiles(t, modelKitPath, []string{"test-file.txt", "test-dir/test-subfile.txt"})

	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:repack1")
	digestOne := digestFromPack(t, packOut)

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

	packOut = runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:repack2")
	digestTwo := digestFromPack(t, packOut)

	assert.Equal(t, digestOne, digestTwo, "Digests should be the same")
}
