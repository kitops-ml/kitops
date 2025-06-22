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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/stretchr/testify/assert"
)

const listKitfile = `
manifestVersion: 1.0.0
package:
  name: test-list-modelkit
model:
  path: .
`

func TestListFormatVariants(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(listKitfile), 0644); err != nil {
		t.Fatal(err)
	}
	setupFiles(t, modelKitPath, []string{"file1"})
	runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:first")

	// Default equals table
	defaultOut := runCommand(t, expectNoError, "list")
	tableOut := runCommand(t, expectNoError, "list", "--format", "table")
	if defaultOut != tableOut {
		t.Fatalf("default output should match table format")
	}

	// Add second modelkit for JSON and template tests
	setupFiles(t, modelKitPath, []string{"file2"})
	runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:second")

	jsonOut := runCommand(t, expectNoError, "list", "--format", "json")
	start := strings.LastIndex(jsonOut, "\n[")
	if start >= 0 {
		start++
	} else {
		start = strings.Index(jsonOut, "[")
	}
	end := strings.LastIndex(jsonOut, "]")
	if start >= 0 && end > start {
		jsonOut = jsonOut[start : end+1]
	}
	var infos []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOut), &infos); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}

	assert.Equal(t, 2, len(infos), "expected 2 entries in json output")

	tmplOut := runCommand(t, expectNoError, "list", "--format", "{{ .Repo }}")
	tmplLines := filterNonDebugLines(tmplOut)

	assert.Equal(t, 2, len(tmplLines), "expected 2 lines from template output")

	complex := "{{ .Repo }}:{{ index .Tags 0 }} - {{ .ModelName }}"
	tmplOut = runCommand(t, expectNoError, "list", "--format", complex)
	tmplLines = filterNonDebugLines(tmplOut)
	assert.Equal(t, 2, len(tmplLines), "expected 2 lines from complex template output")
}

func TestListFormatEmpty(t *testing.T) {
	testPreflight(t)
	tmpDir := setupTempDir(t)
	_, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	jsonOut := runCommand(t, expectNoError, "list", "--format", "json")
	start := strings.LastIndex(jsonOut, "\n[")
	if start >= 0 {
		start++
	} else {
		start = strings.Index(jsonOut, "[")
	}
	end := strings.LastIndex(jsonOut, "]")
	if start >= 0 && end > start {
		jsonOut = jsonOut[start : end+1]
	}
	assert.JSONEq(t, "[]", jsonOut, "expected empty json output")
}

func filterNonDebugLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.HasPrefix(l, "[") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}
