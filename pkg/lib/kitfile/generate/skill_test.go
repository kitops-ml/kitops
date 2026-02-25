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

package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

func TestParseSkillFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		expectNil  bool
		expectName string
		expectDesc string
		expectLic  string
	}{
		{
			name:       "valid frontmatter",
			content:    "---\nname: pdf-tools\ndescription: PDF processing\nlicense: Apache-2.0\n---\n# Skill body",
			expectName: "pdf-tools",
			expectDesc: "PDF processing",
			expectLic:  "Apache-2.0",
		},
		{
			name:       "partial frontmatter - name only",
			content:    "---\nname: my-skill\n---\n# Body",
			expectName: "my-skill",
		},
		{
			name:      "no frontmatter",
			content:   "# Just a markdown file\nNo frontmatter here.",
			expectNil: true,
		},
		{
			name:      "empty frontmatter",
			content:   "---\n---\n# Body",
			expectNil: true,
		},
		{
			name:      "malformed YAML",
			content:   "---\nname: [invalid yaml\n---\n# Body",
			expectNil: true,
		},
		{
			name:      "no closing delimiter",
			content:   "---\nname: test\n# No closing delimiter",
			expectNil: true,
		},
		{
			name:       "triple dashes in YAML value",
			content:    "---\nname: my-skill\ndescription: Use --- to separate sections\n---\n# Body",
			expectName: "my-skill",
			expectDesc: "Use --- to separate sections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "SKILL.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			fm := parseSkillFrontmatter(path)
			if tt.expectNil {
				if fm != nil {
					t.Errorf("expected nil, got %+v", fm)
				}
				return
			}
			if fm == nil {
				t.Fatal("expected non-nil frontmatter")
			}
			if fm.Name != tt.expectName {
				t.Errorf("Name = %q, want %q", fm.Name, tt.expectName)
			}
			if fm.Description != tt.expectDesc {
				t.Errorf("Description = %q, want %q", fm.Description, tt.expectDesc)
			}
			if fm.License != tt.expectLic {
				t.Errorf("License = %q, want %q", fm.License, tt.expectLic)
			}
		})
	}
}

func TestDirContainsSkillMD(t *testing.T) {
	tests := []struct {
		name   string
		files  []string
		expect bool
	}{
		{
			name:   "has SKILL.md",
			files:  []string{"SKILL.md", "README.md"},
			expect: true,
		},
		{
			name:   "has skill.md lowercase",
			files:  []string{"skill.md"},
			expect: true,
		},
		{
			name:   "has Skill.md mixed case",
			files:  []string{"Skill.md"},
			expect: true,
		},
		{
			name:   "no skill file",
			files:  []string{"README.md", "main.py"},
			expect: false,
		},
		{
			name:   "empty directory",
			files:  []string{},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := DirectoryListing{Name: "test", Path: "test"}
			for _, f := range tt.files {
				dir.Files = append(dir.Files, FileListing{Name: f, Path: "test/" + f})
			}

			found, _ := dirContainsSkillMD(dir)
			if found != tt.expect {
				t.Errorf("dirContainsSkillMD() = %v, want %v", found, tt.expect)
			}
		})
	}
}

func TestBuildPromptFromSkill(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	skillContent := "---\nname: test-skill\ndescription: A test skill\n---\n# Body"
	skillDir := filepath.Join(tmpDir, "myskill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	dir := DirectoryListing{
		Name: "myskill",
		Path: "myskill",
		Files: []FileListing{
			{Name: "SKILL.md", Path: "myskill/SKILL.md"},
		},
	}

	prompt := buildPromptFromSkill(dir)
	if prompt.Path != "myskill" {
		t.Errorf("Path = %q, want %q", prompt.Path, "myskill")
	}
	if prompt.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", prompt.Name, "test-skill")
	}
	if prompt.Description != "A test skill" {
		t.Errorf("Description = %q, want %q", prompt.Description, "A test skill")
	}
}

func TestGenerateKitfile_RootSkillMD(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	skillContent := "---\nname: root-skill\ndescription: Root level skill\nlicense: MIT\n---\n# Body"
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "run.py"), []byte("print('hi')"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	kitfile, err := GenerateKitfile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(kitfile.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(kitfile.Prompts))
	}
	if kitfile.Prompts[0].Path != "." {
		t.Errorf("prompt path = %q, want %q", kitfile.Prompts[0].Path, ".")
	}
	if kitfile.Prompts[0].Name != "root-skill" {
		t.Errorf("prompt name = %q, want %q", kitfile.Prompts[0].Name, "root-skill")
	}
	if kitfile.Package.Name != "root-skill" {
		t.Errorf("package name = %q, want %q", kitfile.Package.Name, "root-skill")
	}
	if kitfile.Package.License != "MIT" {
		t.Errorf("package license = %q, want %q", kitfile.Package.License, "MIT")
	}
	// Root SKILL.md should consolidate everything — no code/docs/datasets
	if len(kitfile.Code) != 0 {
		t.Errorf("expected 0 code layers, got %d", len(kitfile.Code))
	}
}

func TestGenerateKitfile_SubdirSkillMD(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	skillDir := filepath.Join(tmpDir, "pdf-tools")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := "---\nname: pdf-tools\ndescription: PDF processing\nlicense: Apache-2.0\n---\n# Body"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "scripts", "run.py"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	kitfile, err := GenerateKitfile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(kitfile.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(kitfile.Prompts))
	}
	if kitfile.Prompts[0].Path != "pdf-tools" {
		t.Errorf("prompt path = %q, want %q", kitfile.Prompts[0].Path, "pdf-tools")
	}
	if kitfile.Prompts[0].Name != "pdf-tools" {
		t.Errorf("prompt name = %q, want %q", kitfile.Prompts[0].Name, "pdf-tools")
	}
	if kitfile.Package.Name != "pdf-tools" {
		t.Errorf("package name = %q, want %q", kitfile.Package.Name, "pdf-tools")
	}
	if kitfile.Package.License != "Apache-2.0" {
		t.Errorf("package license = %q, want %q", kitfile.Package.License, "Apache-2.0")
	}
}

func TestGenerateKitfile_UserOverridesFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	skillContent := "---\nname: skill-name\ndescription: skill desc\nlicense: MIT\n---\n# Body"
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	userPkg := &artifact.Package{
		Name:        "user-name",
		Description: "user desc",
	}
	kitfile, err := GenerateKitfile(dir, userPkg)
	if err != nil {
		t.Fatal(err)
	}

	if kitfile.Package.Name != "user-name" {
		t.Errorf("package name = %q, want %q (user override)", kitfile.Package.Name, "user-name")
	}
	if kitfile.Package.Description != "user desc" {
		t.Errorf("package desc = %q, want %q (user override)", kitfile.Package.Description, "user desc")
	}
	// License not set by user, should come from frontmatter
	if kitfile.Package.License != "MIT" {
		t.Errorf("package license = %q, want %q (from frontmatter)", kitfile.Package.License, "MIT")
	}
}

func TestGenerateKitfile_MultiSkill(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create two skill directories
	for _, skill := range []struct {
		name, desc, license string
	}{
		{"docx", "Word processing", "Apache-2.0"},
		{"xlsx", "Spreadsheet processing", "Apache-2.0"},
	} {
		dir := filepath.Join(tmpDir, skill.name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		content := "---\nname: " + skill.name + "\ndescription: " + skill.desc + "\nlicense: " + skill.license + "\n---\n# Body"
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	kitfile, err := GenerateKitfile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(kitfile.Prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(kitfile.Prompts))
	}
	// Multi-skill: name/desc should NOT be promoted to package level
	if kitfile.Package.Name != "" {
		t.Errorf("package name should be empty for multi-skill, got %q", kitfile.Package.Name)
	}
	// License from first skill should be promoted
	if kitfile.Package.License != "Apache-2.0" {
		t.Errorf("package license = %q, want %q", kitfile.Package.License, "Apache-2.0")
	}
}

func TestGenerateKitfile_MultiSkillMixedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// One skill directory
	skillDir := filepath.Join(tmpDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: my-skill\n---\n# Body"), 0644); err != nil {
		t.Fatal(err)
	}

	// One regular docs directory
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	kitfile, err := GenerateKitfile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(kitfile.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(kitfile.Prompts))
	}
	if kitfile.Prompts[0].Name != "my-skill" {
		t.Errorf("prompt name = %q, want %q", kitfile.Prompts[0].Name, "my-skill")
	}
	if len(kitfile.Docs) != 1 {
		t.Fatalf("expected 1 docs layer, got %d", len(kitfile.Docs))
	}
}

func TestGenerateKitfile_RootSkillOverridesSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Root SKILL.md
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte("---\nname: root\n---\n# Root"), 0644); err != nil {
		t.Fatal(err)
	}
	// Subdirectory also has SKILL.md
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte("---\nname: sub\n---\n# Sub"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := DirectoryListingFromFS(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	kitfile, err := GenerateKitfile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Root SKILL.md takes precedence — single prompt with path "."
	if len(kitfile.Prompts) != 1 {
		t.Fatalf("expected 1 prompt (root takes precedence), got %d", len(kitfile.Prompts))
	}
	if kitfile.Prompts[0].Path != "." {
		t.Errorf("prompt path = %q, want %q", kitfile.Prompts[0].Path, ".")
	}
	if kitfile.Prompts[0].Name != "root" {
		t.Errorf("prompt name = %q, want %q", kitfile.Prompts[0].Name, "root")
	}
}
