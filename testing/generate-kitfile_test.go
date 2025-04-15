// Copyright 2025 The KitOps Authors.
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

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/stretchr/testify/assert"
)

type kitfileGenTestCase struct {
	// Set to filename when loading test
	Name            string
	ModelName       string            `yaml:"modelName"`
	Description     string            `yaml:"description"`
	Files           []string          `yaml:"files"`
	ExpectedKitfile *artifact.KitFile `yaml:"expectedKitfile"`
}

func (t kitfileGenTestCase) withName(name string) kitfileGenTestCase {
	t.Name = name
	return t
}

func TestKitfileGeneration(t *testing.T) {
	testPreflight(t)

	tests := loadAllTestCasesOrPanic[kitfileGenTestCase](t, filepath.Join("testdata", "kitfile-generation"))
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

			// Set up separate directory for files
			testPath := filepath.Join(tmpDir, "testdata")
			if err := os.MkdirAll(testPath, 0755); err != nil {
				t.Fatal(err)
			}

			// Create files listed in test case
			setupFiles(t, testPath, tt.Files)

			runCommand(t, expectNoError, "init", testPath, "--name", tt.ModelName, "--desc", tt.Description, "--force")

			actualKitfile := &artifact.KitFile{}
			modelfile, err := os.Open(filepath.Join(testPath, constants.DefaultKitfileName))
			if !assert.NoError(t, err) {
				return
			}
			t.Cleanup(func() {
				if err := modelfile.Close(); err != nil {
					t.Fatalf("Error closing model file: %s", err)
				}
			})
			if err := actualKitfile.LoadModel(modelfile); !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.ExpectedKitfile, actualKitfile, "Generated Kitfile should match")
		})
	}
}
