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
