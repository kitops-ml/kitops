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

package skill

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	libskill "github.com/kitops-ml/kitops/pkg/lib/skill"
)

// TestSkillCommand_InstallsEmbeddedSkill drives `kit skill` end-to-end for an
// explicit agent in a project directory and verifies the embedded SKILL.md is
// written to the expected agent skills directory with byte-identical content.
func TestSkillCommand_InstallsEmbeddedSkill(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := SkillCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dir", tmpDir, "--agents", "claude-code"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit skill failed: %v", err)
	}

	// claude-code installs under .claude/skills; the skill directory name is the
	// frontmatter name ("kitops").
	skillPath := filepath.Join(tmpDir, ".claude", "skills", skillName, "SKILL.md")
	got, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading installed skill: %v", err)
	}

	if !bytes.Equal(got, embeddedSkillMD) {
		t.Errorf("installed SKILL.md does not match embedded content (%d vs %d bytes)", len(got), len(embeddedSkillMD))
	}
}

// TestSkillCommand_MultipleAgents verifies the skill lands in each requested
// agent's skills directory.
func TestSkillCommand_MultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := SkillCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dir", tmpDir, "--agents", "claude-code,windsurf"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit skill failed: %v", err)
	}

	for _, sub := range []string{
		filepath.Join(".claude", "skills", skillName, "SKILL.md"),
		filepath.Join(".windsurf", "skills", skillName, "SKILL.md"),
	} {
		if _, err := os.Stat(filepath.Join(tmpDir, sub)); err != nil {
			t.Errorf("expected skill at %s: %v", sub, err)
		}
	}
}

// TestSkillCommand_UnknownAgent verifies an invalid agent name is rejected.
func TestSkillCommand_UnknownAgent(t *testing.T) {
	cmd := SkillCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dir", t.TempDir(), "--agents", "not-a-real-agent"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for unknown agent, got nil")
	}
}

// TestEmbeddedSkillFrontmatter guards against drift between the skillName
// constant and the embedded SKILL.md frontmatter, and ensures the file is a
// valid agent skill doc (name + description present).
func TestEmbeddedSkillFrontmatter(t *testing.T) {
	fm := libskill.ParseSkillFrontmatter(embeddedSkillMD)
	if fm == nil {
		t.Fatal("embedded SKILL.md has no parseable frontmatter")
	}
	if fm.Name != skillName {
		t.Errorf("frontmatter name = %q, want %q (must match install dir)", fm.Name, skillName)
	}
	if fm.Description == "" {
		t.Error("frontmatter description is empty")
	}
}
