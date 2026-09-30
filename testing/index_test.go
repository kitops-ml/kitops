// Copyright 2026 The KitOps Authors.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const indexTestKitfile = `
manifestVersion: 1.0.0
package:
  name: index-testing
model:
  path: model.bin
`

// setupIndexTest packs two variants of a ModelKit into one repository and returns their digests.
func setupIndexTest(t *testing.T) (q4Digest, q8Digest string) {
	t.Helper()
	testPreflight(t)

	tmpDir := setupTempDir(t)
	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	require.NoError(t, os.WriteFile(kitfilePath, []byte(indexTestKitfile), 0644))

	require.NoError(t, os.WriteFile(filepath.Join(modelKitPath, "model.bin"), []byte("weights-q4"), 0644))
	q4Digest = digestFromPack(t, runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test/model:q4_0"))

	require.NoError(t, os.WriteFile(filepath.Join(modelKitPath, "model.bin"), []byte("weights-q8"), 0644))
	q8Digest = digestFromPack(t, runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test/model:q8_0"))

	return q4Digest, q8Digest
}

func TestIndexRoundTrip(t *testing.T) {
	q4Digest, q8Digest := setupIndexTest(t)

	runCommand(t, expectNoError, "index", "create", "test/model:all")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q4_0")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q8_0")

	infoOut := runCommand(t, expectNoError, "index", "info", "test/model:all")
	assert.Contains(t, infoOut, q4Digest)
	assert.Contains(t, infoOut, q8Digest)

	inspectOut := runCommand(t, expectNoError, "index", "inspect", "test/model:all")
	assert.Contains(t, inspectOut, "application/vnd.kitops.modelkit.index.v1+json")
	assert.Contains(t, inspectOut, q4Digest)

	runCommand(t, expectNoError, "index", "remove", "test/model:all", "test/model:q4_0")
	infoOut = runCommand(t, expectNoError, "index", "info", "test/model:all")
	assert.NotContains(t, infoOut, q4Digest)
	assert.Contains(t, infoOut, q8Digest)

	runCommand(t, expectNoError, "index", "delete", "test/model:all")
	runCommand(t, expectError, "index", "info", "test/model:all")
}

func TestIndexLabelsAndAnnotations(t *testing.T) {
	_, q8Digest := setupIndexTest(t)

	runCommand(t, expectNoError, "index", "create", "test/model:all")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q4_0",
		"-l", "quantization=q4_0", "-l", "vram=6GB", "-l", `gpuArchs:=["sm_80", "sm_90"]`, "--annotate", "org.example.tested=true")

	infoOut := runCommand(t, expectNoError, "index", "info", "test/model:all")
	assert.Contains(t, infoOut, "quantization=q4_0")
	assert.Contains(t, infoOut, "vram=6GB")
	assert.Contains(t, infoOut, `gpuArchs=["sm_80","sm_90"]`)
	assert.Contains(t, infoOut, "org.example.tested=true")
	// The tag the ModelKit was added by is recorded and shown in its own column.
	assert.Contains(t, infoOut, "test/model:q4_0")
	assert.NotContains(t, infoOut, "ml.kitops.modelkit.original-tag")

	inspectOut := runCommand(t, expectNoError, "index", "inspect", "test/model:all")
	assert.Contains(t, inspectOut, `"quantization": "q4_0"`)
	assert.Contains(t, inspectOut, `"org.example.tested": "true"`)
	assert.Contains(t, inspectOut, `"ml.kitops.modelkit.original-tag": "q4_0"`)

	// Re-adding merges: the untouched label survives, the removed one is gone.
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q4_0", "-l", "vram-")
	infoOut = runCommand(t, expectNoError, "index", "info", "test/model:all")
	assert.Contains(t, infoOut, "quantization=q4_0")
	assert.NotContains(t, infoOut, "vram=")
	assert.Contains(t, infoOut, "org.example.tested=true")

	// Adding by digest records no reference.
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model@"+q8Digest)
	infoOut = runCommand(t, expectNoError, "index", "info", "test/model:all")
	assert.Contains(t, infoOut, "<none>")

	out := runCommand(t, expectError, "index", "add", "test/model:all", "test/model:q8_0", "-l", "bad")
	assert.Contains(t, out, "expected key=value to set, or key- to remove")
}

// TestUnpackIndexSelectsByLabel covers 'kit unpack' against an index: unpacking always produces
// a ModelKit, chosen by label, the way docker pull resolves a multi-architecture image.
func TestUnpackIndexSelectsByLabel(t *testing.T) {
	setupIndexTest(t)

	runCommand(t, expectNoError, "index", "create", "test/model:all")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q4_0", "-l", "quantization=q4_0")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q8_0", "-l", "quantization=q8_0")

	unpackDir := filepath.Join(t.TempDir(), "unpacked")
	runCommand(t, expectNoError, "unpack", "test/model:all", "-l", "quantization=q8_0", "-d", unpackDir)
	modelBytes, err := os.ReadFile(filepath.Join(unpackDir, "model.bin"))
	require.NoError(t, err)
	assert.Equal(t, "weights-q8", string(modelBytes))

	out := runCommand(t, expectError, "unpack", "test/model:all", "-l", "quantization=q2_k", "-d", t.TempDir())
	assert.Contains(t, out, "ModelKit in the index has quantization=q2_k")

	out = runCommand(t, expectError, "unpack", "test/model:all", "-d", t.TempDir())
	assert.Contains(t, out, "select the ModelKit to unpack with --label")

	// Labels select from an index, so a local ModelKit is not used for them.
	out = runCommand(t, expectError, "unpack", "test/model:q4_0", "-l", "quantization=q8_0", "-d", t.TempDir())
	assert.Contains(t, out, "find ModelKit index")
}

// TestIndexDeletePreservesModelKits guards the oras AutoGC cascade end to end: the ModelKits
// an index refers to must survive deleting that index.
func TestIndexDeletePreservesModelKits(t *testing.T) {
	q4Digest, q8Digest := setupIndexTest(t)

	runCommand(t, expectNoError, "index", "create", "test/model:all")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q4_0")
	runCommand(t, expectNoError, "index", "add", "test/model:all", "test/model:q8_0")
	runCommand(t, expectNoError, "index", "delete", "test/model:all")

	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, fmt.Sprintf(`^test/model\s+q4_0.*%s$`, q4Digest), true)
	assertContainsLineRegexp(t, listOut, fmt.Sprintf(`^test/model\s+q8_0.*%s$`, q8Digest), true)

	// inspect and info both read the manifest and its config blob, so they fail if either
	// was collected.
	for _, tag := range []string{"test/model:q4_0", "test/model:q8_0"} {
		inspectOut := runCommand(t, expectNoError, "inspect", tag)
		assert.Contains(t, inspectOut, "index-testing")

		infoOut := runCommand(t, expectNoError, "info", tag)
		assert.Contains(t, infoOut, "index-testing")
	}
}

func TestIndexHiddenFromList(t *testing.T) {
	setupIndexTest(t)

	runCommand(t, expectNoError, "index", "create", "test/model:all")

	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, `^test/model\s+all\s`, false)
}
