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
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

func TestAddContentModulesDeclaresBundlesAndSkills(t *testing.T) {
	kf := &artifact.KitFile{
		Prompts: []artifact.Prompt{
			{Path: "CLAUDE.md"},
			{Name: "review", Path: ".claude/skills/review/"},
			{Name: "style", Path: ".cursor/rules/style.mdc"},
		},
		MCPServers: []artifact.MCPServer{{Name: "tickets", Path: "tickets.mcpb"}},
	}
	result := &ImportResult{
		Agent: &artifact.Agent{Framework: "claude-code", Includes: []artifact.AgentModule{
			{Name: "github", Type: artifact.AgentModuleMCP, Source: &artifact.AgentSource{Endpoint: "https://e.example.com"}, Transport: "http"},
		}},
		SkillPaths: []string{".claude/skills/review"},
	}

	result.AddContentModules(kf)

	if len(result.Agent.Includes) != 3 {
		t.Fatalf("got %d modules, want github + tickets + review: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}
	tickets := moduleByName(t, result.Agent, "tickets")
	if tickets.Type != artifact.AgentModuleMCP || tickets.Source == nil || tickets.Source.ModelKit != "./tickets" {
		t.Errorf("tickets = %+v, want an mcp module sourced from the mcpServers entry", tickets)
	}
	review := moduleByName(t, result.Agent, "review")
	if review.Type != artifact.AgentModuleSkill || review.Source == nil || review.Source.ModelKit != "./review" {
		t.Errorf("review = %+v, want a skill module sourced from the prompts entry", review)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}
}

func TestAddContentModulesBundleReplacesCommand(t *testing.T) {
	kf := &artifact.KitFile{MCPServers: []artifact.MCPServer{{Name: "filesystem", Path: "filesystem.mcpb"}}}
	result := &ImportResult{
		Agent: &artifact.Agent{Includes: []artifact.AgentModule{
			{Name: "filesystem", Type: artifact.AgentModuleMCP, Source: &artifact.AgentSource{Command: "npx", Args: []string{"-y", "server-filesystem"}}, Config: map[string]any{"root": "/"}},
		}},
		UnvendoredServers: []string{"filesystem", "other"},
	}

	result.AddContentModules(kf)

	if len(result.Agent.Includes) != 1 {
		t.Fatalf("got %d modules, want the one replaced in place: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}
	module := result.Agent.Includes[0]
	if module.Source.Command != "" || module.Source.ModelKit != "./filesystem" {
		t.Errorf("module source = %+v, want the bundle in place of the command", module.Source)
	}
	if module.Config == nil {
		t.Errorf("the module's config must survive vendoring")
	}
	if len(result.UnvendoredServers) != 1 || result.UnvendoredServers[0] != "other" {
		t.Errorf("UnvendoredServers = %v, want the vendored server removed", result.UnvendoredServers)
	}
}

func TestAddContentModulesBundleConflictsWithRemote(t *testing.T) {
	kf := &artifact.KitFile{MCPServers: []artifact.MCPServer{{Name: "github", Path: "github.mcpb"}}}
	result := &ImportResult{
		Agent: &artifact.Agent{Includes: []artifact.AgentModule{
			{Name: "github", Type: artifact.AgentModuleMCP, Source: &artifact.AgentSource{Endpoint: "https://e.example.com"}, Transport: "http"},
		}},
	}

	result.AddContentModules(kf)

	if len(result.Agent.Includes) != 1 || result.Agent.Includes[0].Source.Endpoint == "" {
		t.Errorf("the remote module must be kept: %+v", result.Agent.Includes)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "github.mcpb") {
		t.Errorf("Warnings = %v, want the conflict reported", result.Warnings)
	}
}

func TestAddContentModulesSkillNaming(t *testing.T) {
	kf := &artifact.KitFile{
		Prompts: []artifact.Prompt{
			{Name: "github", Path: "skills/github/"},
			{Path: "skills/unnamed/"},
			{Name: "twice", Path: "skills/a/"},
			{Name: "twice", Path: "skills/b/"},
		},
	}
	result := &ImportResult{
		Agent: &artifact.Agent{Includes: []artifact.AgentModule{
			{Name: "github", Type: artifact.AgentModuleMCP, Source: &artifact.AgentSource{Endpoint: "https://e.example.com"}, Transport: "http"},
		}},
		SkillPaths: []string{"skills/github", "skills/unnamed", "skills/a", "skills/b"},
	}

	result.AddContentModules(kf)

	if len(result.Agent.Includes) != 2 {
		t.Fatalf("got %d modules, want github + github-skill: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}
	skill := moduleByName(t, result.Agent, "github-skill")
	if skill.Source == nil || skill.Source.ModelKit != "./github" {
		t.Errorf("github-skill = %+v, want a relative source.modelkit naming the skill", skill)
	}
	if len(result.Warnings) != 2 {
		t.Errorf("Warnings = %v, want the unnamed and the duplicated skills reported once each", result.Warnings)
	}
}

func TestAddContentModulesWithoutGeneratedBlock(t *testing.T) {
	kf := &artifact.KitFile{MCPServers: []artifact.MCPServer{{Name: "tickets", Path: "tickets.mcpb"}}}
	result := &ImportResult{DefinitionPath: "agent-definition.yaml"}
	result.AddContentModules(kf)
	if result.Agent != nil {
		t.Errorf("a definition file declares its own modules; got %+v", result.Agent)
	}
}
