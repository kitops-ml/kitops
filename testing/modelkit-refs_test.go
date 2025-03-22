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

	"github.com/kitops-ml/kitops/pkg/lib/constants"
)

type modelkitRefTestcase struct {
	Name        string
	Description string `yaml:"description"`
	Modelkits   []struct {
		Tag           string   `yaml:"tag"`
		Kitfile       string   `yaml:"kitfile"`
		Kitignore     string   `yaml:"kitignore"`
		Files         []string `yaml:"files"`
		IgnoredFiles  []string `yaml:"ignored"`
		PackErrRegexp *string  `yaml:"packErrRegexp"`
	} `yaml:"modelkits"`
}

func (t modelkitRefTestcase) withName(name string) modelkitRefTestcase {
	t.Name = name
	return t
}

func TestModelKitReferences(t *testing.T) {
	testPreflight(t)

	tests := loadAllTestCasesOrPanic[modelkitRefTestcase](t, filepath.Join("testdata", "modelkit-refs"))
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
			// Set up temporary directory for work
			tmpDir := setupTempDir(t)

			// Set up directory for KITOPS_HOME
			contextPath := filepath.Join(tmpDir, ".kitops")
			if err := os.MkdirAll(contextPath, 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv(constants.KitopsHomeEnvVar, contextPath)

			// Set up temporary directories for modelkits; note we create one extra for the final unpack
			for i := 0; i <= len(tt.Modelkits); i++ {
				if err := os.MkdirAll(filepath.Join(tmpDir, fmt.Sprintf("modelkit-%d", i)), 0755); err != nil {
					t.Fatal(err)
				}
			}

			var allFiles []string
			for idx, modelkit := range tt.Modelkits {
				curDir := filepath.Join(tmpDir, fmt.Sprintf("modelkit-%d", idx))
				nextDir := filepath.Join(tmpDir, fmt.Sprintf("modelkit-%d", idx+1))
				// Set up modelkit directory
				setupKitfileAndKitignore(t, curDir, modelkit.Kitfile, modelkit.Kitignore)
				setupFiles(t, curDir, append(modelkit.Files, modelkit.IgnoredFiles...))

				// Save files that are expected to exist to verify they all end up in the modelkit
				allFiles = append(allFiles, modelkit.Files...)

				// Pack the current dir, unpack it into the next dir. If we expect this to fail, assert that
				// output contains expected text
				if modelkit.PackErrRegexp != nil {
					packOutput := runCommand(t, expectError, "pack", curDir, "-t", modelkit.Tag)
					assertContainsLineRegexp(t, packOutput, *modelkit.PackErrRegexp, true)
					continue
				}
				runCommand(t, expectNoError, "pack", curDir, "-t", modelkit.Tag)
				runCommand(t, expectNoError, "list")
				runCommand(t, expectNoError, "unpack", modelkit.Tag, "-d", nextDir)

				// Verify unpacked contents
				checkFilesExist(t, nextDir, allFiles)
				checkFilesDoNotExist(t, nextDir, modelkit.IgnoredFiles)
			}
		})
	}
}
