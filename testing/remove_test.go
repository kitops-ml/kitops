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
)

const (
	testKitfile = `
manifestVersion: 1.0.0
package:
  name: test-delete-modelkit
model:
  path: .
`
)

func TestRemoveSingleModelkitTag(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:delete_testing")
	digest := digestFromPack(t, packOut)
	modelRegexp := fmt.Sprintf(`^test\s+delete_testing.*%s$`, digest)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	runCommand(t, expectNoError, "remove", "test:delete_testing")
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, false)
}

func TestRemoveSingleModelkitDigest(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:delete_testing")
	digest := digestFromPack(t, packOut)
	modelRegexp := fmt.Sprintf(`^test\s+delete_testing.*%s$`, digest)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	ref := fmt.Sprintf("test@%s", digest)
	runCommand(t, expectNoError, "remove", ref)
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, false)
}

func TestRemoveSingleModelkitNoTag(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath)
	digest := digestFromPack(t, packOut)
	modelRegexp := fmt.Sprintf(`^.*%s$`, digest)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	runCommand(t, expectNoError, "remove", digest)
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, modelRegexp, false)
}

func TestRemoveModelkitUntagsWhenMultiple(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:test_tag_1")
	digest := digestFromPack(t, packOut)
	firstModelRegexp := fmt.Sprintf(`^test\s+test_tag_1.*%s$`, digest)

	runCommand(t, expectNoError, "tag", "test:test_tag_1", "test:test_tag_2")
	secondModelRegexp := fmt.Sprintf(`^test\s+test_tag_2.*%s$`, digest)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, firstModelRegexp, true)
	assertContainsLineRegexp(t, listOut, secondModelRegexp, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	runCommand(t, expectNoError, "remove", "test:test_tag_1")
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, firstModelRegexp, false)
	assertContainsLineRegexp(t, listOut, secondModelRegexp, true)
}

func TestRemoveModelkitUntagsAllWhenDigest(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:test_tag_1")
	digest := digestFromPack(t, packOut)
	firstModelRegexp := fmt.Sprintf(`^test\s+test_tag_1.*%s$`, digest)

	runCommand(t, expectNoError, "tag", "test:test_tag_1", "test:test_tag_2")
	secondModelRegexp := fmt.Sprintf(`^test\s+test_tag_2.*%s$`, digest)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, firstModelRegexp, true)
	assertContainsLineRegexp(t, listOut, secondModelRegexp, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	ref := fmt.Sprintf("test@%s", digest)
	runCommand(t, expectNoError, "remove", ref)
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, firstModelRegexp, false)
	assertContainsLineRegexp(t, listOut, secondModelRegexp, false)
}

func TestRemoveModelkitUntagged(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestOne := digestFromPack(t, packOut)
	regexpOne := fmt.Sprintf("^test.*%s$", digestOne)

	// Create files to pack with a different digests
	setupFiles(t, modelKitPath, []string{"testfile-1"})
	packOut = runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestTwo := digestFromPack(t, packOut)
	regexpTwo := fmt.Sprintf("^test.*%s$", digestTwo)

	setupFiles(t, modelKitPath, []string{"testfile-2"})
	packOut = runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestThree := digestFromPack(t, packOut)
	regexpThree := fmt.Sprintf(`^test\s+testing-tag.*%s$`, digestThree)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, regexpOne, true)
	assertContainsLineRegexp(t, listOut, regexpTwo, true)
	assertContainsLineRegexp(t, listOut, regexpThree, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	runCommand(t, expectNoError, "remove", "--all")
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, regexpOne, false)
	assertContainsLineRegexp(t, listOut, regexpTwo, false)
	assertContainsLineRegexp(t, listOut, regexpThree, true)
}

func TestRemoveModelkitAll(t *testing.T) {
	// Set up temporary directory for work
	tmpDir := setupTempDir(t)

	modelKitPath, _, contextPath := setupTestDirs(t, tmpDir)
	t.Setenv(constants.KitopsHomeEnvVar, contextPath)

	kitfilePath := filepath.Join(modelKitPath, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, []byte(testKitfile), 0644); err != nil {
		t.Fatal(err)
	}

	// Pack model kit and tag it
	packOut := runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestOne := digestFromPack(t, packOut)
	regexpOne := fmt.Sprintf("^test.*%s$", digestOne)

	// Create files to pack with a different digests
	setupFiles(t, modelKitPath, []string{"testfile-1"})
	packOut = runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestTwo := digestFromPack(t, packOut)
	regexpTwo := fmt.Sprintf("^test.*%s$", digestTwo)

	setupFiles(t, modelKitPath, []string{"testfile-2"})
	packOut = runCommand(t, expectNoError, "pack", modelKitPath, "-t", "test:testing-tag")
	digestThree := digestFromPack(t, packOut)
	regexpThree := fmt.Sprintf(`^test\s+testing-tag.*%s$`, digestThree)

	// Ensure modelkit exists in output of 'kit list'
	listOut := runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, regexpOne, true)
	assertContainsLineRegexp(t, listOut, regexpTwo, true)
	assertContainsLineRegexp(t, listOut, regexpThree, true)

	// Remove modelkit and verify it's no longer in 'kit list'
	runCommand(t, expectNoError, "remove", "--all", "--force")
	listOut = runCommand(t, expectNoError, "list")
	assertContainsLineRegexp(t, listOut, regexpOne, false)
	assertContainsLineRegexp(t, listOut, regexpTwo, false)
	assertContainsLineRegexp(t, listOut, regexpThree, false)
}

func digestFromPack(t *testing.T, packOutput string) string {
	digestRegexp := regexp.MustCompile(`Model saved: (sha256:\w+)`)
	matches := digestRegexp.FindStringSubmatch(packOutput)
	if len(matches) != 2 {
		t.Fatal("Failed to find digest from 'kit pack' output")
	}
	t.Logf("Found digest from packing: %s", matches[1])
	return matches[1]
}

func assertContainsLineRegexp(t *testing.T, output, lineRegexp string, shouldContain bool) {
	re := regexp.MustCompile(lineRegexp)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	matches := false
	for _, line := range lines {
		if re.MatchString(line) {
			matches = true
		}
	}
	if shouldContain && !matches {
		t.Fatalf("Output should include regexp %s", lineRegexp)
	}
	if !shouldContain && matches {
		t.Fatalf("Output should not include regexp %s", lineRegexp)
	}
}
