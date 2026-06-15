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

package filesystem

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zw := zip.NewWriter(file)
	for name, contents := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMCPB(t *testing.T) {
	tmpDir := t.TempDir()

	validPath := filepath.Join(tmpDir, "valid.mcpb")
	writeZip(t, validPath, map[string]string{
		"manifest.json":   `{"name": "test-server"}`,
		"server/index.js": "console.log('hi')",
	})

	noManifestPath := filepath.Join(tmpDir, "no-manifest.mcpb")
	writeZip(t, noManifestPath, map[string]string{
		"server/index.js": "console.log('hi')",
	})

	nestedManifestPath := filepath.Join(tmpDir, "nested-manifest.mcpb")
	writeZip(t, nestedManifestPath, map[string]string{
		"subdir/manifest.json": `{"name": "test-server"}`,
	})

	notZipPath := filepath.Join(tmpDir, "not-a-zip.mcpb")
	if err := os.WriteFile(notZipPath, []byte("this is not a zip archive"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		path      string
		errRegexp string
	}{
		{name: "valid bundle", path: validPath},
		{name: "missing manifest.json", path: noManifestPath, errRegexp: "does not contain a manifest.json at its root"},
		{name: "nested-only manifest.json", path: nestedManifestPath, errRegexp: "does not contain a manifest.json at its root"},
		{name: "not a zip archive", path: notZipPath, errRegexp: "is not a valid ZIP archive"},
		{name: "nonexistent file", path: filepath.Join(tmpDir, "missing.mcpb"), errRegexp: "failed to read MCP bundle"},
		{name: "directory path", path: tmpDir, errRegexp: "must be a single .mcpb file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCPB(tt.path)
			if tt.errRegexp == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.Regexp(t, tt.errRegexp, err.Error())
			}
		})
	}
}
