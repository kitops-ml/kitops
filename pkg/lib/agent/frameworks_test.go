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

package agent

import (
	"slices"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/skill"
)

// Every framework must be a known agent, so that its skills paths come from the
// shared registry rather than being restated here. A row whose name the
// registry does not know would silently get no skills path at all.
func TestFrameworkNamesAreKnownAgents(t *testing.T) {
	for _, framework := range Frameworks {
		if !skill.IsValidAgentName(framework.Name) {
			t.Errorf("framework %q is not in the agent registry; kit init and kit skill would disagree about its skills path", framework.Name)
		}
	}
}

// kit init must look where kit skill installs. If these drift, a skill the CLI
// itself installed would not be found by an import in the same directory.
func TestSkillDirsMatchAgentRegistry(t *testing.T) {
	for _, framework := range Frameworks {
		cfg, err := skill.GetAgentConfig(framework.Name)
		if err != nil {
			t.Fatalf("framework %q: %s", framework.Name, err)
		}
		dirs := framework.SkillDirs()
		if !slices.Contains(dirs, normalizePath(cfg.SkillsDir)) {
			t.Errorf("framework %q searches %v, which omits the registry's skills path %q",
				framework.Name, dirs, cfg.SkillsDir)
		}
		for _, extra := range framework.ExtraSkillDirs {
			if !slices.Contains(dirs, normalizePath(extra)) {
				t.Errorf("framework %q omits its own extra skills path %q", framework.Name, extra)
			}
		}
	}
}

// The user-scope skills directory comes from the registry so that overrides
// like CLAUDE_CONFIG_DIR are honored instead of reimplemented.
func TestUserSkillsDirHonorsEnvironmentOverride(t *testing.T) {
	claudeCode := FrameworkByName("claude-code")
	if claudeCode == nil {
		t.Fatal("claude-code row missing")
	}
	t.Setenv("CLAUDE_CONFIG_DIR", "/tmp/custom-claude")
	if got, want := claudeCode.UserSkillsDir(), "/tmp/custom-claude/skills"; got != want {
		t.Errorf("UserSkillsDir() = %q, want %q", got, want)
	}
}
