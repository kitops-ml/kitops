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

package skill

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSkill_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	err := writeSkill(dir, false)
	require.NoError(t, err)

	destPath := filepath.Join(dir, "kit.md")
	content, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, skillContent, content)
}

func TestWriteSkill_CreatesNestedDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "skills")
	err := writeSkill(dir, false)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "kit.md"))
	assert.NoError(t, err)
}

func TestWriteSkill_ErrorsIfFileExists(t *testing.T) {
	dir := t.TempDir()
	destPath := filepath.Join(dir, "kit.md")
	require.NoError(t, os.WriteFile(destPath, []byte("old"), 0o644))

	err := writeSkill(dir, false)
	assert.Error(t, err)

	// file should be unchanged
	content, _ := os.ReadFile(destPath)
	assert.Equal(t, []byte("old"), content)
}

func TestWriteSkill_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	destPath := filepath.Join(dir, "kit.md")
	require.NoError(t, os.WriteFile(destPath, []byte("old"), 0o644))

	err := writeSkill(dir, true)
	require.NoError(t, err)

	content, _ := os.ReadFile(destPath)
	assert.Equal(t, skillContent, content)
}

func TestSkillCommand_NoFlags_PrintsToStdout(t *testing.T) {
	cmd := SkillCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, string(skillContent), buf.String())
}

func TestSkillCommand_MutuallyExclusiveFlags(t *testing.T) {
	cmd := SkillCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--agent", "claude", "--dir", "./some-dir"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSkillCommand_UnknownAgent(t *testing.T) {
	cmd := SkillCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--agent", "unknown-agent"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSkillCommand_AgentFlag_WritesFile(t *testing.T) {
	dir := t.TempDir()

	// override the claude agent dir to point inside our temp dir
	original := agentDirs["claude"]
	agentDirs["claude"] = filepath.Join(dir, ".claude", "skills")
	defer func() { agentDirs["claude"] = original }()

	cmd := SkillCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--agent", "claude"})

	err := cmd.Execute()
	require.NoError(t, err)

	destPath := filepath.Join(agentDirs["claude"], "kit.md")
	content, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, skillContent, content)
}

func TestSkillCommand_DirFlag_WritesFile(t *testing.T) {
	dir := t.TempDir()

	cmd := SkillCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", dir})

	err := cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "kit.md"))
	require.NoError(t, err)
	assert.Equal(t, skillContent, content)
}
