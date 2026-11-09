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
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"go.yaml.in/yaml/v3"
)

// agentYAML renders a definition for comparison. The structs are full of
// pointers, so printing them compares addresses rather than values.
func agentYAML(t *testing.T, agent *artifact.Agent) string {
	t.Helper()
	data, err := yaml.Marshal(agent)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func moduleByName(t *testing.T, agent *artifact.Agent, name string) artifact.AgentModule {
	t.Helper()
	for _, module := range agent.Includes {
		if module.Name == name {
			return module
		}
	}
	t.Fatalf("module %q not found in %v", name, agent.Includes)
	return artifact.AgentModule{}
}

func needByName(t *testing.T, agent *artifact.Agent, name string) artifact.AgentNeed {
	t.Helper()
	for _, need := range agent.Needs {
		if need.Name == name {
			return need
		}
	}
	t.Fatalf("need %q not found in %v", name, agent.Needs)
	return artifact.AgentNeed{}
}

func TestImportNoAgentConfiguration(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.py"), "print('hi')")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Found() {
		t.Errorf("expected no agent definition, got %+v", result.Agent)
	}
}

func TestImportProjectMCPConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
	  "mcpServers": {
	    "github": {
	      "type": "http",
	      "url": "https://api.githubcopilot.com/mcp/",
	      "headers": {"Authorization": "Bearer ghp_A1b2C3d4E5f6G7h8I9j0K1l2M3n4"}
	    },
	    "filesystem": {
	      "command": "npx",
	      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"],
	      "env": {"ALLOWED_PATHS": "/workspace", "API_KEY": "abcd1234"}
	    },
	    "docs": {"url": "https://mcp.example.com/sse"}
	  }
	}`)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("expected an agent definition")
	}
	if result.Framework != "claude-code" {
		t.Errorf("framework = %q, want claude-code", result.Framework)
	}
	if len(result.Agent.Includes) != 3 {
		t.Fatalf("got %d modules, want 3: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}

	// A remote entry keeps its endpoint, and its transport is honored.
	github := moduleByName(t, result.Agent, "github")
	if github.Source.Endpoint != "https://api.githubcopilot.com/mcp/" || github.Transport != "http" {
		t.Errorf("github module = %+v", github)
	}
	// The credential is replaced by a reference and never recorded.
	if got := github.Headers["Authorization"]; got != "Bearer ${needs.GITHUB_TOKEN}" {
		t.Errorf("Authorization header = %q, want a needs reference", got)
	}
	if strings.Contains(agentYAML(t, result.Agent), "ghp_") {
		t.Error("the github token leaked into the definition")
	}
	needByName(t, result.Agent, "GITHUB_TOKEN")

	// Transport is inferred from the URL when the config does not declare one.
	if docs := moduleByName(t, result.Agent, "docs"); docs.Transport != "sse" {
		t.Errorf("docs transport = %q, want sse", docs.Transport)
	}

	// A stdio entry is recorded for kit pack to vendor, not launched as a loose
	// command.
	filesystem := moduleByName(t, result.Agent, "filesystem")
	if filesystem.Source.Command != "npx" || len(filesystem.Source.Args) != 3 {
		t.Errorf("filesystem module = %+v", filesystem)
	}
	if len(result.UnvendoredServers) != 1 || result.UnvendoredServers[0] != "filesystem" {
		t.Errorf("UnvendoredServers = %v, want [filesystem]", result.UnvendoredServers)
	}

	// A plain configuration value survives as a literal; a credential does not.
	if allowed := needByName(t, result.Agent, "ALLOWED_PATHS"); allowed.From.Text != "/workspace" {
		t.Errorf("ALLOWED_PATHS = %+v, want a literal", allowed.From)
	}
	apiKey := needByName(t, result.Agent, "API_KEY")
	if apiKey.From.Text != "" || apiKey.From.EnvFrom != "API_KEY" {
		t.Errorf("API_KEY = %+v, want envFrom with no literal", apiKey.From)
	}

	// The imported config is not agent content; the resolver writes that file at
	// launch, so nothing records it for packing.
	if len(result.BundlePaths) != 0 {
		t.Errorf("BundlePaths = %v, want none", result.BundlePaths)
	}
}

func TestImportScopes(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"),
		`{"mcpServers": {"project-server": {"url": "https://a.example.com/mcp", "type": "http"}}}`)

	claudeJSON := map[string]any{
		"mcpServers": map[string]any{
			"personal-server": map[string]any{"url": "https://personal.example.com/mcp", "type": "http"},
		},
		"projects": map[string]any{
			dir: map[string]any{
				"mcpServers": map[string]any{
					"local-server": map[string]any{"url": "https://local.example.com/mcp", "type": "http"},
				},
			},
		},
	}
	data, err := json.Marshal(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(home, ".claude.json"), string(data))

	// By default project and local scope import; user scope is listed only.
	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: home})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Agent.Includes) != 2 {
		t.Fatalf("got %d modules, want project + local: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}
	moduleByName(t, result.Agent, "project-server")
	moduleByName(t, result.Agent, "local-server")
	if len(result.SkippedUserEntries) != 1 || result.SkippedUserEntries[0].Name != "personal-server" {
		t.Errorf("SkippedUserEntries = %+v, want [personal-server]", result.SkippedUserEntries)
	}

	// --agent-globals promotes the developer's personal tools to project config.
	result, err = Import(ImportOptions{ContextDir: dir, HomeDir: home, IncludeGlobals: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Agent.Includes) != 3 {
		t.Fatalf("got %d modules, want all three: %+v", len(result.Agent.Includes), result.Agent.Includes)
	}
	moduleByName(t, result.Agent, "personal-server")
	if len(result.SkippedUserEntries) != 0 {
		t.Errorf("SkippedUserEntries = %+v, want none", result.SkippedUserEntries)
	}
}

func TestImportSkillsAndPromptsSelectFramework(t *testing.T) {
	dir := t.TempDir()
	// A project with skills but no MCP config is still an agent project.
	writeFile(t, filepath.Join(dir, ".claude", "skills", "review", "SKILL.md"), "# review")
	writeFile(t, filepath.Join(dir, ".claude", "skills", "notes.txt"), "not a skill dir")
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# instructions")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("expected an agent definition from skills alone")
	}
	if len(result.SkillPaths) != 1 || result.SkillPaths[0] != ".claude/skills/review" {
		t.Errorf("SkillPaths = %v, want the SKILL.md directory only", result.SkillPaths)
	}
	if len(result.PromptPaths) != 1 || result.PromptPaths[0] != "CLAUDE.md" {
		t.Errorf("PromptPaths = %v, want [CLAUDE.md]", result.PromptPaths)
	}
	if len(result.Agent.Includes) != 0 {
		t.Errorf("expected no modules, got %+v", result.Agent.Includes)
	}
}

func TestImportCursorSelectedWhenNoClaudeSources(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".cursor", "mcp.json"),
		`{"mcpServers": {"docs": {"url": "https://docs.example.com/mcp", "type": "http"}}}`)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != "cursor" {
		t.Errorf("framework = %q, want cursor", result.Framework)
	}

	// Claude Code sources win when both are present.
	writeFile(t, filepath.Join(dir, ".mcp.json"),
		`{"mcpServers": {"other": {"url": "https://other.example.com/mcp", "type": "http"}}}`)
	result, err = Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != "claude-code" {
		t.Errorf("framework = %q, want claude-code", result.Framework)
	}

	// An explicit override wins over detection.
	result, err = Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir(), Framework: "cursor"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != "cursor" {
		t.Errorf("framework = %q, want cursor", result.Framework)
	}
}

func TestImportUnknownFramework(t *testing.T) {
	_, err := Import(ImportOptions{ContextDir: t.TempDir(), HomeDir: t.TempDir(), Framework: "nope"})
	if err == nil || !strings.Contains(err.Error(), "claude-code") {
		t.Errorf("expected an error listing known frameworks, got %v", err)
	}
}

func TestImportMalformedConfigIsAnError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), "{not json")

	_, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected an error for a config the developer wrote but kit cannot parse")
	}
}

func TestImportEnvVarReferenceBecomesEnvFrom(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
	  "mcpServers": {
	    "srv": {"command": "srv", "env": {"MY_VALUE": "${SOME_OTHER_VAR}"}}
	  }
	}`)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	need := needByName(t, result.Agent, "MY_VALUE")
	if need.From.EnvFrom != "SOME_OTHER_VAR" {
		t.Errorf("MY_VALUE = %+v, want envFrom SOME_OTHER_VAR", need.From)
	}
}

func TestImportIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
	  "mcpServers": {
	    "zebra": {"url": "https://z.example.com/mcp", "type": "http"},
	    "alpha": {"command": "a", "env": {"Z_VAL": "z", "A_VAL": "a"}},
	    "mango": {"url": "https://m.example.com/mcp", "type": "http"}
	  }
	}`)

	first, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		next, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := agentYAML(t, next.Agent), agentYAML(t, first.Agent); got != want {
			t.Fatalf("import is not deterministic:\n%s\n%s", want, got)
		}
	}
}

// The definition kit init produces must be one kit can read back.
func TestImportedDefinitionValidates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
	  "mcpServers": {
	    "github": {"url": "https://api.githubcopilot.com/mcp/", "type": "http",
	               "headers": {"Authorization": "Bearer ghp_A1b2C3d4E5f6G7h8I9j0K1"}},
	    "filesystem": {"command": "npx", "args": ["-y", "server"], "env": {"LOG_LEVEL": "debug"}}
	  }
	}`)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	kitfile := &artifact.KitFile{ManifestVersion: "1.0.0", Agent: result.Agent}
	kitfile.Package.Name = "test"
	if _, err := kitfile.Validate(); err != nil {
		t.Fatalf("generated agent definition does not validate: %s", err)
	}

	// And one that survives a YAML round trip, since that is how it is packed.
	data, err := kitfile.MarshalToYAML()
	if err != nil {
		t.Fatal(err)
	}
	reloaded := &artifact.KitFile{}
	if err := reloaded.LoadModel(io.NopCloser(strings.NewReader(string(data)))); err != nil {
		t.Fatalf("generated Kitfile does not parse: %s\n%s", err, data)
	}
	if got, want := agentYAML(t, reloaded.Agent), agentYAML(t, kitfile.Agent); got != want {
		t.Errorf("agent block changed across a round trip:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
func TestImportPromptAloneIsNotAnAgent(t *testing.T) {
	// CLAUDE.md and AGENTS.md are useful on their own and appear in plenty of
	// repositories that are not agents. They are packed as prompts either way,
	// but they do not make a project an agent project.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# instructions")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Found() {
		t.Fatalf("a prompt file alone must not produce an agent definition, got %+v", result.Agent)
	}
	if len(result.NearMiss) != 1 || result.NearMiss[0] != "CLAUDE.md" {
		t.Errorf("NearMiss = %v, want [CLAUDE.md] so the developer learns why", result.NearMiss)
	}
}

func TestImportFrameworkDirectoryIsTheSignal(t *testing.T) {
	// The framework's own directory is created deliberately for the framework,
	// so it marks the project as an agent project even when it holds nothing
	// kit can import.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), "{}")
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# instructions")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("expected an agent definition from the framework directory")
	}
	if result.Agent.Framework != "claude-code" {
		t.Errorf("framework = %q, want claude-code", result.Agent.Framework)
	}
	if len(result.Agent.Includes) != 0 || len(result.Agent.Needs) != 0 {
		t.Errorf("expected an empty definition, got %+v", result.Agent)
	}
	// The prompt is still project content once the project is an agent project.
	if len(result.PromptPaths) != 1 || result.PromptPaths[0] != "CLAUDE.md" {
		t.Errorf("PromptPaths = %v, want [CLAUDE.md]", result.PromptPaths)
	}
}

func TestImportSymlinkedPromptIsNotReportedAsPacked(t *testing.T) {
	// kit packs regular files and directories only, so a symlinked prompt is
	// found by the framework here but does not travel in the ModelKit. It must
	// not be reported as packed content.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), "{}")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "# instructions")
	if err := os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Skipf("symlinks unavailable: %s", err)
	}

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("a symlinked prompt still identifies the framework")
	}
	if result.Framework != "claude-code" {
		t.Errorf("framework = %q, want claude-code", result.Framework)
	}
	for _, path := range result.PromptPaths {
		if path == "CLAUDE.md" {
			t.Error("CLAUDE.md is a symlink and is not packed; it must not be reported as packed content")
		}
	}
	var warned bool
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "CLAUDE.md") && strings.Contains(warning, "symlink") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("expected a warning that the symlinked prompt will not travel, got %v", result.Warnings)
	}
}

func TestImportNearMissIsReported(t *testing.T) {
	// Agent-adjacent files with no framework directory should say so rather
	// than leaving the developer wondering why they got nothing.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "# instructions")
	writeFile(t, filepath.Join(dir, "main.go"), "package main")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Found() {
		t.Fatalf("expected no definition, got %+v", result.Agent)
	}
	if len(result.NearMiss) == 0 {
		t.Error("expected AGENTS.md to be reported as a near miss")
	}
}

func TestImportRecordsSkillsWithoutFrameworkSettings(t *testing.T) {
	// The framework's directory is its working space and holds machine-local
	// settings. It is never packed as a whole; the agent content inside it is
	// recorded individually by path.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), `{"permissions":{"allow":["Bash"]}}`)
	writeFile(t, filepath.Join(dir, ".claude", "skills", "review", "SKILL.md"), "# review")
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{"mcpServers":{}}`)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	// The skill inside the framework directory is recorded by path, so it
	// travels without the machine-local settings beside it.
	if !slices.Contains(result.SkillPaths, ".claude/skills/review") {
		t.Errorf("SkillPaths = %v, want the skill inside the framework directory", result.SkillPaths)
	}
	if slices.Contains(result.SkillPaths, ".claude/settings.local.json") {
		t.Errorf("SkillPaths = %v, machine-local settings must not be recorded", result.SkillPaths)
	}
}

func TestImportDiscoversMCPBundles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{"mcpServers":{}}`)
	writeFile(t, filepath.Join(dir, "filesystem.mcpb"), "PK")
	writeFile(t, filepath.Join(dir, "postgres.mcpb"), "PK")
	writeFile(t, filepath.Join(dir, "notes.md"), "not a bundle")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"filesystem.mcpb", "postgres.mcpb"}
	if !slices.Equal(result.BundlePaths, want) {
		t.Errorf("BundlePaths = %v, want %v", result.BundlePaths, want)
	}
}

func TestImportPromptsCoverBothNames(t *testing.T) {
	// Claude Code reads CLAUDE.md and AGENTS.md, and projects commonly make one
	// a symlink to the other. Whichever holds the regular file must be packed,
	// or the agent ships with no instructions.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), "{}")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "# instructions")
	if err := os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Skipf("symlinks unavailable: %s", err)
	}

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(result.PromptPaths, "AGENTS.md") {
		t.Errorf("PromptPaths = %v, want AGENTS.md: the symlink's target carries the instructions", result.PromptPaths)
	}
	if slices.Contains(result.PromptPaths, "CLAUDE.md") {
		t.Errorf("PromptPaths = %v, the symlink itself is not packed", result.PromptPaths)
	}
}

func TestImportCursorSkillsAndRulesAreDistinct(t *testing.T) {
	// Cursor reads SKILL.md from both .cursor/skills and .agents/skills, and
	// keeps rules as a separate mechanism that triggers by explicit condition.
	// Recording a rule as a skill would misreport when it fires.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".cursor", "mcp.json"), `{"mcpServers":{}}`)
	writeFile(t, filepath.Join(dir, ".cursor", "skills", "deploy", "SKILL.md"), "# deploy")
	writeFile(t, filepath.Join(dir, ".agents", "skills", "review", "SKILL.md"), "# review")
	writeFile(t, filepath.Join(dir, ".cursor", "rules", "style.mdc"), "---\nalwaysApply: true\n---\nrule")

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != "cursor" {
		t.Fatalf("framework = %q, want cursor", result.Framework)
	}
	for _, want := range []string{".agents/skills/review", ".cursor/skills/deploy"} {
		if !slices.Contains(result.SkillPaths, want) {
			t.Errorf("SkillPaths = %v, want it to contain %s", result.SkillPaths, want)
		}
	}
	if slices.Contains(result.SkillPaths, ".cursor/rules/style.mdc") {
		t.Error("a rule must not be recorded as a skill: rules trigger by explicit condition, skills by relevance")
	}
	if !slices.Contains(result.RulePaths, ".cursor/rules/style.mdc") {
		t.Errorf("RulePaths = %v, want the rule recorded as content", result.RulePaths)
	}
}

func TestImportLocalScopeSelectsFramework(t *testing.T) {
	// Local scope is bound to this project even though it lives in the user's
	// home, so servers configured there are this project's configuration and
	// must be enough to detect an agent project on their own.
	dir := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.py"), "x = 1")
	claudeJSON, err := json.Marshal(map[string]any{
		"projects": map[string]any{
			dir: map[string]any{"mcpServers": map[string]any{
				"localsrv": map[string]any{"url": "https://l.example.com/mcp", "type": "http"},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(home, ".claude.json"), string(claudeJSON))

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: home})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("local-scope servers alone must select the framework")
	}
	moduleByName(t, result.Agent, "localsrv")
}

func TestImportBareProjectEntryDoesNotSelectFramework(t *testing.T) {
	// A framework may write a project entry for every directory it has ever run
	// in. Those carry no servers and must not turn unrelated directories into
	// agent projects.
	dir := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.py"), "x = 1")
	claudeJSON, err := json.Marshal(map[string]any{
		"projects": map[string]any{
			dir: map[string]any{"mcpServers": map[string]any{}, "history": []any{"a prompt"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(home, ".claude.json"), string(claudeJSON))

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: home})
	if err != nil {
		t.Fatal(err)
	}
	if result.Found() {
		t.Fatalf("a bare project entry must not select a framework, got %+v", result.Agent)
	}
}

func TestImportUserSkillsAreListedNotImported(t *testing.T) {
	// A Kitfile layer path cannot leave the context directory, so a skill under
	// the user's home cannot be packed by reference at any scope. It is
	// reported so the developer knows it is not traveling.
	dir := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{"mcpServers":{}}`)
	writeFile(t, filepath.Join(home, ".claude", "skills", "personal", "SKILL.md"), "# personal")
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))

	for _, globals := range []bool{false, true} {
		result, err := Import(ImportOptions{ContextDir: dir, HomeDir: home, IncludeGlobals: globals})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.SkippedUserSkills) != 1 {
			t.Fatalf("IncludeGlobals=%v: SkippedUserSkills = %v, want the one user skill", globals, result.SkippedUserSkills)
		}
		for _, path := range result.SkillPaths {
			if strings.Contains(path, "personal") {
				t.Errorf("IncludeGlobals=%v: a user-scope skill must never be recorded for packing, got %q", globals, path)
			}
		}
	}
}

const handwrittenDefinition = `apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: handwritten
spec:
  framework: claude-code
  includes:
    - name: github
      type: mcp
      source:
        endpoint: https://api.githubcopilot.com/mcp/
      transport: http
`

func TestImportExistingDefinitionIsPackagedNotModified(t *testing.T) {
	// A definition the developer wrote is the agent. kit init packs it as-is,
	// never rewrites it, and never writes a second definition beside it, so the
	// project's framework config is not imported.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), "{}")
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# instructions")
	writeFile(t, filepath.Join(dir, ".mcp.json"),
		`{"mcpServers": {"other": {"url": "https://o.example.com/mcp", "type": "http"}}}`)
	writeFile(t, filepath.Join(dir, DefinitionFileName), handwrittenDefinition)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() {
		t.Fatal("an existing definition is an agent source")
	}
	if result.DefinitionPath != DefinitionFileName {
		t.Errorf("DefinitionPath = %q, want %q", result.DefinitionPath, DefinitionFileName)
	}
	if result.Agent != nil {
		t.Errorf("no agent block may be generated beside an existing definition, got %+v", result.Agent)
	}
	if len(result.Sources) != 0 || len(result.UnvendoredServers) != 0 {
		t.Errorf("framework config must not be imported: sources=%v unvendored=%v", result.Sources, result.UnvendoredServers)
	}
	if !slices.Contains(result.PromptPaths, "CLAUDE.md") {
		t.Errorf("PromptPaths = %v, the agent's content is still collected", result.PromptPaths)
	}
	data, err := os.ReadFile(filepath.Join(dir, DefinitionFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != handwrittenDefinition {
		t.Error("the definition file was modified")
	}
}

func TestImportExistingDefinitionSelectsAgentPathAlone(t *testing.T) {
	// A definition is the strongest agent signal there is; it needs no framework
	// directory to make the project an agent project.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.py"), "x = 1")
	writeFile(t, filepath.Join(dir, DefinitionFileName), handwrittenDefinition)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found() || result.DefinitionPath != DefinitionFileName {
		t.Fatalf("expected the definition to be found, got %+v", result)
	}
	if result.Framework != "claude-code" {
		t.Errorf("Framework = %q, want the definition's own default", result.Framework)
	}
}

func TestImportDefinitionDocumentSelection(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{"not an agent definition", "kind: SomethingElse\n", ""},
		{"empty file", "", ""},
		{"second document", "kind: SomethingElse\n---\n" + handwrittenDefinition, DefinitionFileName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, DefinitionFileName), tt.contents)
			path, _, err := existingDefinition(dir)
			if err != nil {
				t.Fatal(err)
			}
			if path != tt.want {
				t.Errorf("existingDefinition() = %q, want %q", path, tt.want)
			}
		})
	}
}

func TestImportDefinitionKitCannotCarryIsAnError(t *testing.T) {
	// The developer wrote a definition; failing at init beats an agent
	// ModelKit that turns out at launch to carry no definition.
	t.Run("malformed", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, DefinitionFileName), "kind: [unclosed")
		if _, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()}); err == nil {
			t.Error("expected an error for a definition that cannot be parsed")
		}
	})
	t.Run("symlink", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "real.yaml"), handwrittenDefinition)
		if err := os.Symlink("real.yaml", filepath.Join(dir, DefinitionFileName)); err != nil {
			t.Skipf("symlinks unavailable: %s", err)
		}
		if _, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()}); err == nil {
			t.Error("expected an error: a symlinked definition would not be packed")
		}
	})
}

func TestImportHiddenDefinitionFile(t *testing.T) {
	// The dot-prefixed spelling is supported and packed under its own name.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, HiddenDefinitionFileName), handwrittenDefinition)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.DefinitionPath != HiddenDefinitionFileName {
		t.Errorf("DefinitionPath = %q, want %q", result.DefinitionPath, HiddenDefinitionFileName)
	}
	if result.Agent != nil {
		t.Errorf("no agent block may be generated beside an existing definition, got %+v", result.Agent)
	}
}

func TestImportBothDefinitionFilesIsAnError(t *testing.T) {
	// A ModelKit may carry only one definition in its code layer, and kit has no
	// basis for choosing between the two spellings.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, DefinitionFileName), handwrittenDefinition)
	writeFile(t, filepath.Join(dir, HiddenDefinitionFileName), handwrittenDefinition)

	_, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected an error when both files hold a definition")
	}
	for _, name := range []string{DefinitionFileName, HiddenDefinitionFileName} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q should name %s", err, name)
		}
	}
}

func TestImportBothFilesOnlyOneADefinition(t *testing.T) {
	// A file by one of the names that holds something else is not a definition,
	// so it neither conflicts nor wins.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, DefinitionFileName), "kind: SomethingElse\n")
	writeFile(t, filepath.Join(dir, HiddenDefinitionFileName), handwrittenDefinition)

	result, err := Import(ImportOptions{ContextDir: dir, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if result.DefinitionPath != HiddenDefinitionFileName {
		t.Errorf("DefinitionPath = %q, want %q", result.DefinitionPath, HiddenDefinitionFileName)
	}
}
