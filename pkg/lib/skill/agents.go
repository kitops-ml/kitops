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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AgentConfig defines the directory layout and detection logic for a coding agent.
type AgentConfig struct {
	Name             string
	DisplayName      string
	SkillsDir        string          // project-relative skills path (e.g., ".claude/skills")
	GlobalDetectDirs func() []string // returns dirs to probe for detection; first existing dir wins
	GlobalSkillsDir  func() string   // computes absolute global skills path
}

// envOrDefault returns the value of the named environment variable if set and
// non-empty (after trimming whitespace), otherwise returns the default.
func envOrDefault(key, defaultVal string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultVal
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

func xdgConfigHome() string {
	return envOrDefault("XDG_CONFIG_HOME", filepath.Join(homeDir(), ".config"))
}

func claudeHome() string {
	return envOrDefault("CLAUDE_CONFIG_DIR", filepath.Join(homeDir(), ".claude"))
}

func codexHome() string {
	return envOrDefault("CODEX_HOME", filepath.Join(homeDir(), ".codex"))
}

// homeAgent builds an AgentConfig where detection and global skills are both
// under ~/.<dotDir>. This covers the majority of agents.
func homeAgent(name, displayName, skillsDir, dotDir string) AgentConfig {
	return AgentConfig{
		Name: name, DisplayName: displayName,
		SkillsDir:        skillsDir,
		GlobalDetectDirs: func() []string { return []string{filepath.Join(homeDir(), dotDir)} },
		GlobalSkillsDir:  func() string { return filepath.Join(homeDir(), dotDir, "skills") },
	}
}

// homeAgentCustomGlobal builds an AgentConfig where detection is ~/.<detectDir>
// but the global skills path differs from the detect path.
func homeAgentCustomGlobal(name, displayName, skillsDir, detectDir string, globalSkillsFn func() string) AgentConfig {
	return AgentConfig{
		Name: name, DisplayName: displayName,
		SkillsDir:        skillsDir,
		GlobalDetectDirs: func() []string { return []string{filepath.Join(homeDir(), detectDir)} },
		GlobalSkillsDir:  globalSkillsFn,
	}
}

// agentRegistry holds the complete agent registry. Ported from
// https://github.com/vercel-labs/skills/blob/main/src/agents.ts
var agentRegistry = map[string]AgentConfig{
	// Agents using XDG or env-var-based paths
	"amp": {
		Name: "amp", DisplayName: "Amp",
		SkillsDir:        ".agents/skills",
		GlobalDetectDirs: func() []string { return []string{filepath.Join(xdgConfigHome(), "amp")} },
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "agents", "skills") },
	},
	"claude-code": {
		Name: "claude-code", DisplayName: "Claude Code",
		SkillsDir:        ".claude/skills",
		GlobalDetectDirs: func() []string { return []string{claudeHome()} },
		GlobalSkillsDir:  func() string { return filepath.Join(claudeHome(), "skills") },
	},
	"codex": {
		Name: "codex", DisplayName: "Codex",
		SkillsDir:        ".agents/skills",
		GlobalDetectDirs: func() []string { return []string{codexHome()} },
		GlobalSkillsDir:  func() string { return filepath.Join(codexHome(), "skills") },
	},
	"crush": {
		Name: "crush", DisplayName: "Crush",
		SkillsDir:        ".crush/skills",
		GlobalDetectDirs: func() []string { return []string{filepath.Join(xdgConfigHome(), "crush")} },
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "crush", "skills") },
	},
	"goose": {
		Name: "goose", DisplayName: "Goose",
		SkillsDir:        ".goose/skills",
		GlobalDetectDirs: func() []string { return []string{filepath.Join(xdgConfigHome(), "goose")} },
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "goose", "skills") },
	},
	"opencode": {
		Name: "opencode", DisplayName: "OpenCode",
		SkillsDir:        ".agents/skills",
		GlobalDetectDirs: func() []string { return []string{filepath.Join(xdgConfigHome(), "opencode")} },
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "opencode", "skills") },
	},
	"openclaw": {
		Name: "openclaw", DisplayName: "OpenClaw",
		SkillsDir: "skills",
		GlobalDetectDirs: func() []string {
			h := homeDir()
			return []string{
				filepath.Join(h, ".openclaw"),
				filepath.Join(h, ".clawdbot"),
				filepath.Join(h, ".moltbot"),
			}
		},
		GlobalSkillsDir: func() string { return filepath.Join(homeDir(), ".openclaw", "skills") },
	},

	// Agents where detect dir and global skills dir diverge
	"antigravity": homeAgentCustomGlobal("antigravity", "Antigravity", ".agents/skills", ".gemini/antigravity", func() string { return filepath.Join(homeDir(), ".gemini", "antigravity", "skills") }),
	"cline":       homeAgentCustomGlobal("cline", "Cline", ".agents/skills", ".cline", func() string { return filepath.Join(homeDir(), ".agents", "skills") }),
	"cortex":      homeAgentCustomGlobal("cortex", "Cortex Code", ".cortex/skills", ".snowflake/cortex", func() string { return filepath.Join(homeDir(), ".snowflake", "cortex", "skills") }),
	"deepagents":  homeAgentCustomGlobal("deepagents", "Deep Agents", ".agents/skills", ".deepagents", func() string { return filepath.Join(homeDir(), ".deepagents", "agent", "skills") }),
	"kimi-cli":    homeAgentCustomGlobal("kimi-cli", "Kimi Code CLI", ".agents/skills", ".kimi", func() string { return filepath.Join(xdgConfigHome(), "agents", "skills") }),
	"pi":          homeAgentCustomGlobal("pi", "Pi", ".pi/skills", ".pi/agent", func() string { return filepath.Join(homeDir(), ".pi", "agent", "skills") }),
	"warp":        homeAgentCustomGlobal("warp", "Warp", ".agents/skills", ".warp", func() string { return filepath.Join(homeDir(), ".agents", "skills") }),
	"windsurf":    homeAgentCustomGlobal("windsurf", "Windsurf", ".windsurf/skills", ".codeium/windsurf", func() string { return filepath.Join(homeDir(), ".codeium", "windsurf", "skills") }),

	// Simple agents: detect at ~/.<dir>, skills at ~/.<dir>/skills
	"adal":           homeAgent("adal", "AdaL", ".adal/skills", ".adal"),
	"augment":        homeAgent("augment", "Augment", ".augment/skills", ".augment"),
	"bob":            homeAgent("bob", "IBM Bob", ".bob/skills", ".bob"),
	"codebuddy":      homeAgent("codebuddy", "CodeBuddy", ".codebuddy/skills", ".codebuddy"),
	"command-code":   homeAgent("command-code", "Command Code", ".commandcode/skills", ".commandcode"),
	"continue":       homeAgent("continue", "Continue", ".continue/skills", ".continue"),
	"cursor":         homeAgentCustomGlobal("cursor", "Cursor", ".agents/skills", ".cursor", func() string { return filepath.Join(homeDir(), ".cursor", "skills") }),
	"droid":          homeAgent("droid", "Droid", ".factory/skills", ".factory"),
	"firebender":     homeAgentCustomGlobal("firebender", "Firebender", ".agents/skills", ".firebender", func() string { return filepath.Join(homeDir(), ".firebender", "skills") }),
	"gemini-cli":     homeAgentCustomGlobal("gemini-cli", "Gemini CLI", ".agents/skills", ".gemini", func() string { return filepath.Join(homeDir(), ".gemini", "skills") }),
	"github-copilot": homeAgentCustomGlobal("github-copilot", "GitHub Copilot", ".agents/skills", ".copilot", func() string { return filepath.Join(homeDir(), ".copilot", "skills") }),
	"iflow-cli":      homeAgent("iflow-cli", "iFlow CLI", ".iflow/skills", ".iflow"),
	"junie":          homeAgent("junie", "Junie", ".junie/skills", ".junie"),
	"kilo":           homeAgent("kilo", "Kilo Code", ".kilocode/skills", ".kilocode"),
	"kiro-cli":       homeAgent("kiro-cli", "Kiro CLI", ".kiro/skills", ".kiro"),
	"kode":           homeAgent("kode", "Kode", ".kode/skills", ".kode"),
	"mcpjam":         homeAgent("mcpjam", "MCPJam", ".mcpjam/skills", ".mcpjam"),
	"mistral-vibe":   homeAgent("mistral-vibe", "Mistral Vibe", ".vibe/skills", ".vibe"),
	"mux":            homeAgent("mux", "Mux", ".mux/skills", ".mux"),
	"neovate":        homeAgent("neovate", "Neovate", ".neovate/skills", ".neovate"),
	"openhands":      homeAgent("openhands", "OpenHands", ".openhands/skills", ".openhands"),
	"pochi":          homeAgent("pochi", "Pochi", ".pochi/skills", ".pochi"),
	"qoder":          homeAgent("qoder", "Qoder", ".qoder/skills", ".qoder"),
	"qwen-code":      homeAgent("qwen-code", "Qwen Code", ".qwen/skills", ".qwen"),
	"roo":            homeAgent("roo", "Roo Code", ".roo/skills", ".roo"),
	"trae":           homeAgent("trae", "Trae", ".trae/skills", ".trae"),
	"trae-cn":        homeAgentCustomGlobal("trae-cn", "Trae CN", ".trae/skills", ".trae-cn", func() string { return filepath.Join(homeDir(), ".trae-cn", "skills") }),
	"zencoder":       homeAgent("zencoder", "Zencoder", ".zencoder/skills", ".zencoder"),

	// Non-detectable agents
	"replit": {
		Name: "replit", DisplayName: "Replit",
		SkillsDir:        ".agents/skills",
		GlobalDetectDirs: nil,
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "agents", "skills") },
	},
	"universal": {
		Name: "universal", DisplayName: "Universal",
		SkillsDir:        ".agents/skills",
		GlobalDetectDirs: nil,
		GlobalSkillsDir:  func() string { return filepath.Join(xdgConfigHome(), "agents", "skills") },
	},
}

// GetAgentConfig returns the configuration for the named agent.
func GetAgentConfig(name string) (AgentConfig, error) {
	cfg, ok := agentRegistry[name]
	if !ok {
		return AgentConfig{}, fmt.Errorf("unknown agent %q", name)
	}
	return cfg, nil
}

// ValidAgentNames returns a sorted list of all valid agent names.
func ValidAgentNames() []string {
	names := make([]string, 0, len(agentRegistry))
	for name := range agentRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsValidAgentName returns true if the name is a known agent.
func IsValidAgentName(name string) bool {
	_, ok := agentRegistry[name]
	return ok
}

// GetProjectSkillsDir returns the absolute skills directory for the given
// agent within a project directory.
func GetProjectSkillsDir(agentName, projectDir string) (string, error) {
	cfg, err := GetAgentConfig(agentName)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, cfg.SkillsDir), nil
}

// GetGlobalSkillsDir returns the absolute global skills directory for the
// given agent.
func GetGlobalSkillsDir(agentName string) (string, error) {
	cfg, err := GetAgentConfig(agentName)
	if err != nil {
		return "", err
	}
	return cfg.GlobalSkillsDir(), nil
}

// DetectInstalledAgents probes global config directories to find installed
// agents. Always checks global paths regardless of installation scope.
// Returns a sorted list of detected agent names.
// Agents with no detection probe (replit, universal) are silently skipped.
func DetectInstalledAgents() ([]string, error) {
	var detected []string
	for name, cfg := range agentRegistry {
		if cfg.GlobalDetectDirs == nil {
			continue
		}
		dirs := cfg.GlobalDetectDirs()
		for _, dir := range dirs {
			info, err := os.Stat(dir)
			if err == nil && info.IsDir() {
				detected = append(detected, name)
				break
			}
		}
	}
	sort.Strings(detected)
	return detected, nil
}
