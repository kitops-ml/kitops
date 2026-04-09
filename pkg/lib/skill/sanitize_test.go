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
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"oras.land/oras-go/v2/registry"
)

func Test_sanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Awesome Skill", "my-awesome-skill"},
		{"claude-code", "claude-code"},
		{"CLAUDE", "claude"},
		{"../../../etc/passwd", "etc-passwd"},
		{"skill@2.0", "skill-2.0"},
		{"...leading-dots", "leading-dots"},
		{"trailing-dots...", "trailing-dots"},
		{"", ""},
		{"   ", ""},
		{"a/b/c", "a-b-c"},
		{"hello_world.v2", "hello_world.v2"},
		{strings.Repeat("a", 300), strings.Repeat("a", 255)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDeriveSkillName(t *testing.T) {
	ref := &registry.Reference{Repository: "myrepo/my-model", Reference: "v1"}

	tests := []struct {
		name            string
		frontmatterName string
		prompt          artifact.Prompt
		modelRef        *registry.Reference
		expected        string
	}{
		{
			name:            "frontmatter name takes priority",
			frontmatterName: "code-review",
			prompt:          artifact.Prompt{Path: "./code/SKILL.md"},
			modelRef:        ref,
			expected:        "code-review",
		},
		{
			name:            "prompt name from Kitfile",
			frontmatterName: "",
			prompt:          artifact.Prompt{Name: "my-skill", Path: "./skills/SKILL.md"},
			modelRef:        ref,
			expected:        "my-skill",
		},
		{
			name:            "parent directory for SKILL.md",
			frontmatterName: "",
			prompt:          artifact.Prompt{Path: "./tools/SKILL.md"},
			modelRef:        ref,
			expected:        "tools",
		},
		{
			name:            "root SKILL.md falls back to modelRef",
			frontmatterName: "",
			prompt:          artifact.Prompt{Path: "SKILL.md"},
			modelRef:        ref,
			expected:        "my-model",
		},
		{
			name:            "directory prompt",
			frontmatterName: "",
			prompt:          artifact.Prompt{Path: "skills-dir/"},
			modelRef:        ref,
			expected:        "skills-dir",
		},
		{
			name:            "nil modelRef uses unnamed-skill",
			frontmatterName: "",
			prompt:          artifact.Prompt{Path: "SKILL.md"},
			modelRef:        nil,
			expected:        "unnamed-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveSkillName(tt.frontmatterName, tt.prompt, tt.modelRef)
			if got != tt.expected {
				t.Errorf("DeriveSkillName() = %q, want %q", got, tt.expected)
			}
		})
	}
}
