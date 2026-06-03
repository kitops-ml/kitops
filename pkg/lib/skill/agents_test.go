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
	"testing"
)

func TestGetAgentConfig(t *testing.T) {
	cfg, err := GetAgentConfig("claude-code")
	if err != nil {
		t.Fatalf("GetAgentConfig(claude-code) error: %v", err)
	}
	if cfg.Name != "claude-code" {
		t.Errorf("expected name claude-code, got %s", cfg.Name)
	}
	if cfg.SkillsDir != ".claude/skills" {
		t.Errorf("expected skillsDir .claude/skills, got %s", cfg.SkillsDir)
	}

	_, err = GetAgentConfig("nonexistent-agent")
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

// nonDetectableAgents lists the agents that intentionally have no detection
// probe (universal/placeholder entries). All other agents must define one.
var nonDetectableAgents = map[string]bool{
	"replit":    true,
	"universal": true,
}

// TestAgentRegistryWellFormed iterates the full registry and verifies that
// every entry has the fields required for skill installation and detection.
// This is generic — adding new agents does not require updating this test.
func TestAgentRegistryWellFormed(t *testing.T) {
	for name, cfg := range agentRegistry {
		t.Run(name, func(t *testing.T) {
			if cfg.Name != name {
				t.Errorf("registry key %q does not match cfg.Name %q", name, cfg.Name)
			}
			if cfg.DisplayName == "" {
				t.Errorf("DisplayName is empty")
			}
			if cfg.SkillsDir == "" {
				t.Errorf("SkillsDir is empty")
			}
			if cfg.GlobalSkillsDir == nil {
				t.Fatalf("GlobalSkillsDir is nil")
			}
			if got := cfg.GlobalSkillsDir(); got == "" {
				t.Errorf("GlobalSkillsDir() returned empty string")
			}
			if nonDetectableAgents[name] {
				if cfg.GlobalDetectDirs != nil {
					t.Errorf("expected GlobalDetectDirs to be nil for non-detectable agent")
				}
				return
			}
			if cfg.GlobalDetectDirs == nil {
				t.Fatalf("GlobalDetectDirs is nil")
			}
			if dirs := cfg.GlobalDetectDirs(); len(dirs) == 0 {
				t.Errorf("GlobalDetectDirs() returned no dirs")
			}
		})
	}
}
