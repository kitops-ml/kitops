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
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

func TestInstallSkill_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     12,
			},
			Content: []byte("# My Skill\n"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "my-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
		return
	}

	// Verify file was written
	skillPath := filepath.Join(tmpDir, ".claude", "skills", "my-skill", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading skill file: %v", err)
	}
	if string(content) != "# My Skill\n" {
		t.Errorf("unexpected content: %q", string(content))
	}
}

func TestInstallSkill_MultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     7,
			},
			Content: []byte("# Test\n"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code", "windsurf"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "test-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
		return
	}

	// Verify both paths
	for _, sub := range []string{".claude/skills/test-skill/SKILL.md", ".windsurf/skills/test-skill/SKILL.md"} {
		path := filepath.Join(tmpDir, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file at %s", path)
		}
	}
}

func TestInstallSkill_PathDedup(t *testing.T) {
	tmpDir := t.TempDir()

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     7,
			},
			Content: []byte("# Test\n"),
		},
	}

	// cursor and cline both resolve to .agents/skills in project scope
	opts := &SkillInstallOptions{
		Agents:     []string{"cursor", "cline"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "test-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
	}

	// One should be written, one should be skipped (dedup)
	var skipped int
	for _, ar := range result.Agents {
		if ar.Skipped {
			skipped++
		}
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped (dedup), got %d", skipped)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, ".agents", "skills", "test-skill", "SKILL.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file at %s", path)
	}
}

func TestInstallSkill_IgnoreExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the skill directory
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "existing-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     3,
			},
			Content: []byte("new"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:         []string{"claude-code"},
		ProjectDir:     tmpDir,
		IgnoreExisting: true,
	}

	result := InstallSkill(entries, "existing-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if result.HasErrors() {
		t.Errorf("unexpected errors: %v", result.Errors())
	}

	// Content should be unchanged (old)
	content, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if string(content) != "old" {
		t.Errorf("expected old content preserved, got %q", string(content))
	}
}

func TestInstallSkill_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the skill directory
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "overwrite-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     3,
			},
			Content: []byte("new"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
		Overwrite:  true,
	}

	result := InstallSkill(entries, "overwrite-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if result.HasErrors() {
		t.Errorf("unexpected errors: %v", result.Errors())
	}

	content, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if string(content) != "new" {
		t.Errorf("expected new content, got %q", string(content))
	}
}

func TestInstallSkill_ExistingWithoutFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the skill directory
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "existing-skill")
	os.MkdirAll(skillDir, 0755)

	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     3,
			},
			Content: []byte("new"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "existing-skill", artifact.Prompt{Path: "SKILL.md"}, opts)
	if !result.HasErrors() {
		t.Error("expected error when directory exists without -o or -i")
	}
}

func TestInstallSkill_DirectoryPrefixStripping(t *testing.T) {
	tmpDir := t.TempDir()

	// Tar entries have the prompt path as prefix, matching how kit pack works.
	// Tests that prefix stripping places SKILL.md at the skill root.
	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "skills/docx/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			},
		},
		{
			Header: &tar.Header{
				Name:     "skills/docx/SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     7,
			},
			Content: []byte("# Skill"),
		},
		{
			Header: &tar.Header{
				Name:     "skills/docx/subdir/helper.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     6,
			},
			Content: []byte("helper"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "docx", artifact.Prompt{Path: "skills/docx/"}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
		return
	}

	skillRoot := filepath.Join(tmpDir, ".claude", "skills", "docx")

	// SKILL.md at skill root, not nested under skills/docx/
	for _, rel := range []string{"SKILL.md", "subdir/helper.md"} {
		p := filepath.Join(skillRoot, rel)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected file at %s", p)
		}
	}

	// Must NOT exist at nested path
	badPath := filepath.Join(skillRoot, "skills", "docx", "SKILL.md")
	if _, err := os.Stat(badPath); err == nil {
		t.Errorf("SKILL.md should not be nested at %s", badPath)
	}
}

func TestInstallSkill_WindowsPromptPathOnUnix(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate a ModelKit packed on Windows: prompt.Path has backslashes
	// but tar entries always use forward slashes
	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "skills/docx/SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     7,
			},
			Content: []byte("# Skill"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	// Backslash path as stored in Kitfile from Windows pack
	result := InstallSkill(entries, "docx", artifact.Prompt{Path: "skills\\docx\\"}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
		return
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "docx", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("SKILL.md not found at skill root: %s", skillPath)
	}
}

func TestInstallSkill_DotPath(t *testing.T) {
	tmpDir := t.TempDir()

	// prompt path is "." — entire context dir packed as one layer
	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     12,
			},
			Content: []byte("# Copy Edit\n"),
		},
		{
			Header: &tar.Header{
				Name:     "examples/test.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     4,
			},
			Content: []byte("test"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "copy-edit", artifact.Prompt{Path: "."}, opts)
	if result.HasErrors() {
		for _, e := range result.Errors() {
			t.Errorf("agent %s error: %v", e.Agent, e.Err)
		}
		return
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "copy-edit", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("SKILL.md not found at %s", skillPath)
	}
	examplePath := filepath.Join(tmpDir, ".claude", "skills", "copy-edit", "examples", "test.md")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Errorf("examples/test.md not found at %s", examplePath)
	}
}

func TestInstallSkill_PrefixMismatchErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Prompt path doesn't match tar entry prefix — should error, not silently install empty dir
	entries := []TarEntry{
		{
			Header: &tar.Header{
				Name:     "actual-dir/SKILL.md",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     7,
			},
			Content: []byte("# Skill"),
		},
	}

	opts := &SkillInstallOptions{
		Agents:     []string{"claude-code"},
		ProjectDir: tmpDir,
	}

	result := InstallSkill(entries, "test", artifact.Prompt{Path: "wrong-dir/"}, opts)
	if !result.HasErrors() {
		t.Error("expected error when prompt path doesn't match tar entries")
	}
}
