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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"oras.land/oras-go/v2/registry"
)

// SkillInstallOptions configures how skills are installed.
type SkillInstallOptions struct {
	Agents         []string
	ProjectDir     string // if non-empty, install project-scoped; if empty, install globally
	Overwrite      bool
	IgnoreExisting bool
	ModelRef       *registry.Reference
}

// AgentInstallResult records the outcome for a single agent.
type AgentInstallResult struct {
	Agent   string
	Path    string // resolved absolute path
	Skipped bool   // true if skipped due to IgnoreExisting or path dedup
	Err     error
}

// InstallResult records the outcome of installing one skill across all agents.
type InstallResult struct {
	SkillName string
	Prompt    artifact.Prompt
	Agents    []AgentInstallResult
}

// HasErrors returns true if any agent installation failed.
func (r InstallResult) HasErrors() bool {
	for _, a := range r.Agents {
		if a.Err != nil {
			return true
		}
	}
	return false
}

// Errors returns all agent-level errors.
func (r InstallResult) Errors() []AgentInstallResult {
	var errs []AgentInstallResult
	for _, a := range r.Agents {
		if a.Err != nil {
			errs = append(errs, a)
		}
	}
	return errs
}

// resolveSkillsDir returns the absolute skills directory for an agent.
func resolveSkillsDir(agentName, projectDir string) (string, error) {
	if projectDir != "" {
		return GetProjectSkillsDir(agentName, projectDir)
	}
	return GetGlobalSkillsDir(agentName)
}

// normalizePath converts a path to use forward slashes and cleans it. Tar
// entries always use forward slashes, but prompt paths from Kitfiles packed on
// Windows may contain backslashes. Replacing backslashes with forward slashes
// before cleaning ensures consistent comparison on all platforms.
func normalizePath(p string) string {
	return path.Clean(strings.ReplaceAll(p, "\\", "/"))
}

// stripPromptPrefix removes the prompt's path prefix from tar entry names so
// that files are installed directly under the skill directory instead of nested
// under the original directory structure from the pack context.
//
// For example, if the prompt path is "skills/docx/" and a tar entry is
// "skills/docx/SKILL.md", the entry name becomes "SKILL.md".
// For a single-file prompt like "SKILL.md", the entry name is unchanged.
//
// Returns an error if stripping would produce zero entries (indicates a
// separator mismatch or path alignment bug).
func stripPromptPrefix(entries []TarEntry, promptPath string) ([]TarEntry, error) {
	if len(entries) == 0 {
		return entries, nil
	}

	// Normalize to forward slashes so comparisons work regardless of
	// whether the Kitfile was packed on Windows or Unix.
	prefix := normalizePath(promptPath)

	// "." means the entire context directory is the prompt — entries are
	// already at root, no prefix to strip.
	if prefix == "." {
		return entries, nil
	}

	// Determine if the prefix is a single file by checking if any entry
	// matches the prefix exactly as a regular file
	for _, e := range entries {
		if normalizePath(e.Header.Name) == prefix && e.Header.Typeflag != tar.TypeDir {
			// Single file prompt — strip to just the filename
			result := make([]TarEntry, len(entries))
			for i, entry := range entries {
				result[i] = entry
				if normalizePath(entry.Header.Name) == prefix {
					result[i].Header = cloneHeader(entry.Header)
					result[i].Header.Name = path.Base(prefix)
				}
			}
			return result, nil
		}
	}

	// Directory prompt — strip the normalized prefix from all descendant entries.
	// Since both sides use forward slashes, do this by removing prefix + "/"
	// from matching paths rather than using filepath-specific path handling.
	result := make([]TarEntry, 0, len(entries))
	for _, entry := range entries {
		cleaned := normalizePath(entry.Header.Name)
		// If cleaned == prefix, it's the directory entry itself — skip it.
		// Otherwise, if cleaned starts with prefix + "/", take everything after it.
		if cleaned == prefix {
			continue
		}
		trimmed := strings.TrimPrefix(cleaned, prefix+"/")
		if trimmed == cleaned {
			// Entry is outside the prefix — skip
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "..") {
			continue
		}
		newEntry := TarEntry{
			Header:  cloneHeader(entry.Header),
			Content: entry.Content,
		}
		newEntry.Header.Name = trimmed
		result = append(result, newEntry)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("prompt path %q does not match any tar entries in the layer", promptPath)
	}
	return result, nil
}

func cloneHeader(h *tar.Header) *tar.Header {
	clone := *h
	return &clone
}

// InstallSkill writes the buffered tar entries as a skill to each agent's
// skill directory. Does NOT fail-fast: attempts every agent, returns
// per-agent results.
func InstallSkill(entries []TarEntry, skillName string, prompt artifact.Prompt, opts *SkillInstallOptions) InstallResult {
	// Strip the prompt's path prefix so files land at the skill root
	stripped, err := stripPromptPrefix(entries, prompt.Path)
	if err != nil {
		result := InstallResult{SkillName: skillName, Prompt: prompt}
		for _, agent := range opts.Agents {
			result.Agents = append(result.Agents, AgentInstallResult{
				Agent: agent,
				Err:   err,
			})
		}
		return result
	}
	entries = stripped

	result := InstallResult{
		SkillName: skillName,
		Prompt:    prompt,
	}

	// Deduplicate agents by resolved path
	type agentPath struct {
		agent string
		path  string
	}
	seen := map[string][]string{} // resolved path : list of agent names
	var order []agentPath

	for _, agent := range opts.Agents {
		baseDir, err := resolveSkillsDir(agent, opts.ProjectDir)
		if err != nil {
			result.Agents = append(result.Agents, AgentInstallResult{
				Agent: agent,
				Err:   err,
			})
			continue
		}
		skillDir := filepath.Join(baseDir, skillName)
		absDir, err := filepath.Abs(skillDir)
		if err != nil {
			result.Agents = append(result.Agents, AgentInstallResult{
				Agent: agent,
				Path:  skillDir,
				Err:   fmt.Errorf("resolving path: %w", err),
			})
			continue
		}

		if agents, ok := seen[absDir]; ok {
			// Duplicate path — record as skipped
			seen[absDir] = append(agents, agent)
			result.Agents = append(result.Agents, AgentInstallResult{
				Agent:   agent,
				Path:    absDir,
				Skipped: true,
			})
			continue
		}
		seen[absDir] = []string{agent}
		order = append(order, agentPath{agent: agent, path: absDir})
	}

	for _, ap := range order {
		agentResult := installForAgent(entries, skillName, ap.agent, ap.path, opts)
		result.Agents = append(result.Agents, agentResult)
	}

	return result
}

// installForAgent handles installation for a single agent at a resolved path.
func installForAgent(entries []TarEntry, skillName, agent, skillDir string, opts *SkillInstallOptions) AgentInstallResult {
	errResult := func(err error) AgentInstallResult {
		return AgentInstallResult{Agent: agent, Path: skillDir, Err: err}
	}

	baseDir := filepath.Dir(skillDir)
	if _, _, err := filesystem.VerifySubpath(baseDir, skillName); err != nil {
		return errResult(fmt.Errorf("invalid skill name %q: resolves outside skills directory", skillName))
	}

	for _, entry := range entries {
		if entry.Header.Name == "" {
			continue
		}
		if _, _, err := filesystem.VerifySubpath(skillDir, entry.Header.Name); err != nil {
			return errResult(fmt.Errorf("illegal file path in prompt layer: %s", entry.Header.Name))
		}
	}

	// Check if skill directory already exists
	if info, err := os.Stat(skillDir); err == nil && info.IsDir() {
		if opts.IgnoreExisting {
			return AgentInstallResult{Agent: agent, Path: skillDir, Skipped: true}
		}
		if !opts.Overwrite {
			return errResult(fmt.Errorf("skill directory already exists (use -o to overwrite)"))
		}
		if err := os.RemoveAll(skillDir); err != nil {
			return errResult(fmt.Errorf("removing existing skill directory: %w", err))
		}
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return errResult(fmt.Errorf("creating skill directory: %w", err))
	}

	for _, entry := range entries {
		outPath := filepath.Join(skillDir, entry.Header.Name)

		switch entry.Header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, entry.Header.FileInfo().Mode()); err != nil {
				return errResult(fmt.Errorf("creating directory %s: %w", entry.Header.Name, err))
			}
		default: // tar.TypeReg
			if dir := filepath.Dir(outPath); dir != skillDir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errResult(fmt.Errorf("creating parent directory for %s: %w", entry.Header.Name, err))
				}
			}
			if err := os.WriteFile(outPath, entry.Content, entry.Header.FileInfo().Mode()); err != nil {
				return errResult(fmt.Errorf("writing file %s: %w", entry.Header.Name, err))
			}
		}
	}

	return AgentInstallResult{Agent: agent, Path: skillDir}
}
