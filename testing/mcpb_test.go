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
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpbMediaType = "application/vnd.kitops.modelkit.mcpb.v1.raw"

const mcpbKitfile = `
manifestVersion: 1.0.0
package:
  name: agent-mcp-bundle
  version: 1.0.0
mcpServers:
  - name: filesystem
    path: ./filesystem.mcpb
  - name: postgres
    path: ./postgres.mcpb
docs:
  - path: ./README.md
`

// writeMCPB creates a ZIP archive at path containing the given entries
func writeMCPB(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()
	zw := zip.NewWriter(file)
	for name, contents := range entries {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(contents))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
}

func sha256OfFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	sum := sha256.Sum256(contents)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func setupMCPBModelKit(t *testing.T, modelKitPath string) {
	setupKitfileAndKitignore(t, modelKitPath, mcpbKitfile, "")
	setupFiles(t, modelKitPath, []string{"README.md"})
	writeMCPB(t, filepath.Join(modelKitPath, "filesystem.mcpb"), map[string]string{
		"manifest.json":   `{"name": "filesystem-server"}`,
		"server/index.js": "console.log('fs')",
	})
	writeMCPB(t, filepath.Join(modelKitPath, "postgres.mcpb"), map[string]string{
		"manifest.json":   `{"name": "postgres-server"}`,
		"server/index.js": "console.log('pg')",
	})
}

// TestMCPBPackUnpack packs a ModelKit with two MCP bundles, verifies each is
// stored as its own layer whose digest is the SHA-256 of the bundle's bytes,
// and verifies unpacking yields byte-identical archives.
func TestMCPBPackUnpack(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	modelKitPath, unpackPath, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	setupMCPBModelKit(t, modelKitPath)

	filesystemDigest := sha256OfFile(t, filepath.Join(modelKitPath, "filesystem.mcpb"))
	postgresDigest := sha256OfFile(t, filepath.Join(modelKitPath, "postgres.mcpb"))

	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", modelKitTag)
	// The layer digest must be the SHA-256 of the .mcpb file's bytes (stored verbatim)
	assert.Contains(t, packOut, "Saved mcpb layer: "+filesystemDigest)
	assert.Contains(t, packOut, "Saved mcpb layer: "+postgresDigest)

	inspectOut := runCommand(t, expectNoError, "inspect", modelKitTag)
	assert.Contains(t, inspectOut, mcpbMediaType)
	assert.Contains(t, inspectOut, filesystemDigest)
	assert.Contains(t, inspectOut, postgresDigest)

	runCommand(t, expectNoError, "unpack", modelKitTag, "-d", unpackPath)

	for _, bundle := range []string{"filesystem.mcpb", "postgres.mcpb"} {
		original, err := os.ReadFile(filepath.Join(modelKitPath, bundle))
		require.NoError(t, err)
		unpacked, err := os.ReadFile(filepath.Join(unpackPath, bundle))
		require.NoError(t, err)
		assert.Equal(t, original, unpacked, "Unpacked %s should be byte-identical to the original", bundle)
	}
}

// TestMCPBUnpackFilter verifies that mcpb layers can be selected via the
// mcpservers filter type, by name.
func TestMCPBUnpackFilter(t *testing.T) {
	testPreflight(t)

	tmpDir := setupTempDir(t)
	modelKitPath, unpackPath, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	setupMCPBModelKit(t, modelKitPath)

	runCommand(t, expectNoError, "pack", modelKitPath, "-t", modelKitTag)
	runCommand(t, expectNoError, "unpack", modelKitTag, "--filter=mcpservers:postgres", "-d", unpackPath)

	checkFilesExist(t, unpackPath, []string{"postgres.mcpb"})
	checkFilesDoNotExist(t, unpackPath, []string{"filesystem.mcpb", "README.md"})
}

// TestMCPBPackFailures verifies that packing fails for malformed bundles and
// for the modelpack format.
func TestMCPBPackFailures(t *testing.T) {
	testPreflight(t)

	kitfile := `
manifestVersion: 1.0.0
mcpServers:
  - name: bad
    path: ./bad.mcpb
`

	t.Run("missing manifest.json", func(t *testing.T) {
		tmpDir := setupTempDir(t)
		modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
		t.Setenv(constants.KitopsHomeEnvVar, contextPath)

		setupKitfileAndKitignore(t, modelKitPath, kitfile, "")
		writeMCPB(t, filepath.Join(modelKitPath, "bad.mcpb"), map[string]string{
			"server/index.js": "console.log('no manifest')",
		})

		runCommand(t, expectError, "pack", modelKitPath, "-t", modelKitTag)
	})

	t.Run("not a zip archive", func(t *testing.T) {
		tmpDir := setupTempDir(t)
		modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
		t.Setenv(constants.KitopsHomeEnvVar, contextPath)

		setupKitfileAndKitignore(t, modelKitPath, kitfile, "")
		require.NoError(t, os.WriteFile(filepath.Join(modelKitPath, "bad.mcpb"), []byte("not a zip"), 0o644))

		runCommand(t, expectError, "pack", modelKitPath, "-t", modelKitTag)
	})

	t.Run("modelpack format", func(t *testing.T) {
		tmpDir := setupTempDir(t)
		modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
		t.Setenv(constants.KitopsHomeEnvVar, contextPath)

		setupMCPBModelKit(t, modelKitPath)

		runCommand(t, expectError, "pack", modelKitPath, "-t", modelKitTag, "--use-model-pack")
	})
}
