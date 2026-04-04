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
	"bytes"
	"strings"
	"testing"
)

func createTar(t *testing.T, files map[string]string) *tar.Reader {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range files {
		if strings.HasSuffix(name, "/") {
			if err := tw.WriteHeader(&tar.Header{
				Name:     name,
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}); err != nil {
				t.Fatalf("writing dir header: %v", err)
			}
			continue
		}
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
			Mode:     0644,
		}); err != nil {
			t.Fatalf("writing header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("writing content: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar: %v", err)
	}
	return tar.NewReader(&buf)
}

func TestReadSkillLayer_WithSkillMd(t *testing.T) {
	content := "---\nname: my-skill\ndescription: test\n---\n\n# My Skill\n"
	tr := createTar(t, map[string]string{
		"SKILL.md": content,
	})

	entries, isSkill, fmName, err := ReadSkillLayer(tr)
	if err != nil {
		t.Fatalf("ReadSkillLayer error: %v", err)
	}
	if !isSkill {
		t.Error("expected isSkill=true")
	}
	if fmName != "my-skill" {
		t.Errorf("expected frontmatterName=my-skill, got %q", fmName)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestReadSkillLayer_WithoutSkillMd(t *testing.T) {
	tr := createTar(t, map[string]string{
		"CLAUDE.md": "# Claude prompt",
	})

	_, isSkill, _, err := ReadSkillLayer(tr)
	if err != nil {
		t.Fatalf("ReadSkillLayer error: %v", err)
	}
	if isSkill {
		t.Error("expected isSkill=false for CLAUDE.md")
	}
}

func TestReadSkillLayer_NestedSkillMd(t *testing.T) {
	tr := createTar(t, map[string]string{
		"skills/":              "",
		"skills/docx/":         "",
		"skills/docx/SKILL.md": "---\nname: nested-skill\ndescription: deeply nested\n---\n\nContent",
	})

	entries, isSkill, fmName, err := ReadSkillLayer(tr)
	if err != nil {
		t.Fatalf("ReadSkillLayer error: %v", err)
	}
	if !isSkill {
		t.Error("expected isSkill=true for nested SKILL.md at skills/docx/SKILL.md")
	}
	if fmName != "nested-skill" {
		t.Errorf("expected frontmatterName=nested-skill, got %q", fmName)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestReadSkillLayer_RejectsSymlinks(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name:     "link",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	}); err != nil {
		t.Fatalf("writing header: %v", err)
	}
	tw.Close()

	tr := tar.NewReader(&buf)
	_, _, _, err := ReadSkillLayer(tr)
	if err == nil {
		t.Error("expected error for symlink entry")
	}
	if !strings.Contains(err.Error(), "symlinks and hard links are not allowed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestReadSkillLayer_RejectsOversizedEntry(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	// Write a header claiming huge size but don't actually write the content
	if err := tw.WriteHeader(&tar.Header{
		Name:     "SKILL.md",
		Size:     MaxSkillEntrySize + 1,
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}); err != nil {
		t.Fatalf("writing header: %v", err)
	}
	tw.Close()

	tr := tar.NewReader(&buf)
	_, _, _, err := ReadSkillLayer(tr)
	if err == nil {
		t.Error("expected error for oversized entry")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectNil   bool
		expectedName string
	}{
		{
			name:        "basic frontmatter",
			content:     "---\nname: my-skill\ndescription: test\n---\n\nBody",
			expectedName: "my-skill",
		},
		{
			name:      "no frontmatter",
			content:   "# Just a heading\n\nSome content",
			expectNil: true,
		},
		{
			name:        "no name field",
			content:     "---\ndescription: test\n---\n",
			expectedName: "",
		},
		{
			name:      "empty content",
			content:   "",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := ParseSkillFrontmatter([]byte(tt.content))
			if tt.expectNil {
				if fm != nil {
					t.Errorf("expected nil, got %+v", fm)
				}
				return
			}
			if fm == nil {
				t.Fatal("expected non-nil frontmatter")
			}
			if fm.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", fm.Name, tt.expectedName)
			}
		})
	}
}
