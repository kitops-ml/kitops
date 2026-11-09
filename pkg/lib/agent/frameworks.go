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

// Package agent holds the framework tables and import logic shared by the
// commands that produce and consume an Agent Definition. Adding support for a
// framework means adding one row to Frameworks.
package agent

import (
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/skill"
)

// Framework is one row of the framework table. The Discovery fields are read by
// kit init to import an existing project's agent configuration; the Adapter
// fields are read by the resolver to write the framework's own configuration
// back out. Both live in one row so the two cannot disagree about a path.
type Framework struct {
	// Name is the identifier written to the Agent Definition's framework field.
	Name string

	// --- Discovery (kit init) ---

	// ProjectDir is the framework's own directory in the project, relative to
	// the working directory. Its presence is what marks the project as an agent
	// project: a developer creates it deliberately for this framework, unlike a
	// prompt file, which is useful on its own and carries no such intent.
	ProjectDir string
	// ProjectMCPConfig is the framework's MCP config file, relative to the
	// working directory. Project scope: imported.
	ProjectMCPConfig string
	// LocalMCPConfig is a file in the user's home directory that holds entries
	// bound to a specific project directory. Local scope: imported. Claude Code
	// keeps its "local" scope here rather than in the project.
	LocalMCPConfig string
	// UserMCPConfig is a file in the user's home directory holding entries that
	// describe the developer's personal environment. User scope: listed, and
	// imported only with --agent-globals.
	UserMCPConfig string
	// ExtraSkillDirs are project-relative skill locations specific to this
	// framework that the shared agent registry does not cover. The registry's
	// own path is always searched first; see SkillDirs.
	ExtraSkillDirs []string
	// ProjectRules are project-relative directories holding rule files, which
	// are recorded as content and never as skills.
	ProjectRules []string
	// Prompts are the framework's prompt files, relative to the working
	// directory. They are packed as project content when the project is an
	// agent project, but on their own they do not make it one.
	Prompts []string

	// --- Adapter (resolver) ---

	// MCPConfigPath is where the resolver writes the framework's MCP config,
	// relative to the working directory.
	MCPConfigPath string
	// Entrypoint is the command the resolver execs in the working directory.
	Entrypoint []string
}

// Frameworks is the framework table. A framework that cannot be expressed as a
// row here is a gap in the row schema to fix for every row, not a special case.
var Frameworks = []Framework{
	{
		Name:             "claude-code",
		ProjectDir:       ".claude",
		ProjectMCPConfig: ".mcp.json",
		LocalMCPConfig:   ".claude.json",
		UserMCPConfig:    ".claude.json",
		// Claude Code reads AGENTS.md as well as CLAUDE.md, and projects
		// commonly keep one as a symlink to the other. Listing both means the
		// instructions travel whichever name holds the regular file.
		Prompts:       []string{"CLAUDE.md", "AGENTS.md"},
		MCPConfigPath: ".mcp.json",
		Entrypoint:    []string{"claude"},
	},
	{
		Name:             "cursor",
		ProjectDir:       ".cursor",
		ProjectMCPConfig: ".cursor/mcp.json",
		UserMCPConfig:    ".cursor/mcp.json",
		// Cursor reads SKILL.md from .cursor/skills as well as the registry's
		// .agents/skills; the registry carries only the latter for this agent.
		ExtraSkillDirs: []string{".cursor/skills"},
		// Rules are a separate mechanism from skills and are not deprecated:
		// those with alwaysApply or globs trigger by explicit condition and
		// have no skills equivalent.
		ProjectRules:  []string{".cursor/rules"},
		Prompts:       []string{"AGENTS.md"},
		MCPConfigPath: ".cursor/mcp.json",
		Entrypoint:    []string{"cursor", "."},
	},
}

// FrameworkByName returns the row for name, or nil if the framework is unknown.
func FrameworkByName(name string) *Framework {
	for i, framework := range Frameworks {
		if framework.Name == name {
			return &Frameworks[i]
		}
	}
	return nil
}

// FrameworkNames returns every known framework name, for error messages that
// need to list the valid choices.
func FrameworkNames() []string {
	names := make([]string, 0, len(Frameworks))
	for _, framework := range Frameworks {
		names = append(names, framework.Name)
	}
	return names
}

// SkillDirs returns the project-relative directories holding this framework's
// skills: the agent registry's path, where `kit skill` installs, plus any
// ExtraSkillDirs.
func (f Framework) SkillDirs() []string {
	var dirs []string
	if cfg, err := skill.GetAgentConfig(f.Name); err == nil && cfg.SkillsDir != "" {
		dirs = append(dirs, normalizePath(cfg.SkillsDir))
	}
	for _, dir := range f.ExtraSkillDirs {
		dirs = append(dirs, normalizePath(dir))
	}
	slices.Sort(dirs)
	return slices.Compact(dirs)
}

// UserSkillsDir is the framework's user-scope skills directory, an absolute
// path. It comes from the agent registry, so environment overrides such as
// CLAUDE_CONFIG_DIR and XDG_CONFIG_HOME are honored rather than reimplemented.
func (f Framework) UserSkillsDir() string {
	cfg, err := skill.GetAgentConfig(f.Name)
	if err != nil || cfg.GlobalSkillsDir == nil {
		return ""
	}
	return cfg.GlobalSkillsDir()
}

// ProjectDirs returns each framework's project directory, for a message that
// tells a developer what kit init was looking for.
func ProjectDirs() []string {
	var dirs []string
	for _, framework := range Frameworks {
		if framework.ProjectDir != "" {
			dirs = append(dirs, normalizePath(framework.ProjectDir)+"/")
		}
	}
	return dirs
}

// normalizePath converts a table path to the slash-separated form used in
// Kitfile paths and DirectoryListing entries.
func normalizePath(p string) string {
	return strings.TrimSuffix(strings.ReplaceAll(p, "\\", "/"), "/")
}
