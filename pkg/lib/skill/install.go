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

// relativeEntryName computes the path an entry should be written to relative
// to the skill directory. It strips the prompt path prefix so that files land
// at the skill root instead of being nested under the original pack structure.
//
// Tar entries always use forward slashes. Prompt paths from Kitfiles packed on
// Windows may contain backslashes, so both are normalized to forward slashes.
//
// Returns empty string for entries that should be skipped (e.g. the directory
// entry matching the prefix itself).
func relativeEntryName(entryName, promptPath string) string {
	normalize := func(p string) string {
		return path.Clean(strings.ReplaceAll(p, "\\", "/"))
	}

	prefix := normalize(promptPath)
	cleaned := normalize(entryName)

	// "." prefix means entries are already at root
	if prefix == "." {
		return cleaned
	}

	// Single file: entry name matches the prefix exactly — use just the filename
	if cleaned == prefix {
		return path.Base(cleaned)
	}

	// Directory: strip prefix + "/" from the entry
	trimmed := strings.TrimPrefix(cleaned, prefix+"/")
	if trimmed == cleaned {
		return "" // outside the prefix
	}
	if trimmed == "" || strings.HasPrefix(trimmed, "..") {
		return ""
	}
	return trimmed
}

// InstallSkill writes the buffered tar entries as a skill to each agent's
// skill directory. Does NOT fail-fast: attempts every agent, returns
// per-agent results.
func InstallSkill(entries []TarEntry, skillName string, prompt artifact.Prompt, opts *SkillInstallOptions) InstallResult {
	result := InstallResult{
		SkillName: skillName,
		Prompt:    prompt,
	}

	// Deduplicate agents by resolved path
	type agentPath struct {
		agent string
		path  string
	}
	seen := map[string][]string{}
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

		if _, ok := seen[absDir]; ok {
			seen[absDir] = append(seen[absDir], agent)
			result.Agents = append(result.Agents, AgentInstallResult{
				Agent:   agent,
				Path:    absDir,
				Skipped: true,
			})
			continue
		}
		seen[absDir] = append(seen[absDir], agent)
		order = append(order, agentPath{agent: agent, path: absDir})
	}

	for _, ap := range order {
		agentResult := installForAgent(entries, skillName, prompt.Path, ap.agent, ap.path, opts)
		result.Agents = append(result.Agents, agentResult)
	}

	return result
}

// installForAgent handles installation for a single agent at a resolved path.
// It strips the prompt path prefix from entry names at write time.
func installForAgent(entries []TarEntry, skillName, promptPath, agent, skillDir string, opts *SkillInstallOptions) AgentInstallResult {
	errResult := func(err error) AgentInstallResult {
		return AgentInstallResult{Agent: agent, Path: skillDir, Err: err}
	}

	// Validate skill name — it must be a safe relative path component
	if _, _, err := filesystem.VerifySubpath(".", skillName); err != nil {
		return errResult(fmt.Errorf("invalid skill name '%s': %w", skillName, err))
	}

	// Validate all entries before writing anything
	var hasWritableEntries bool
	for _, entry := range entries {
		if entry.Header.Name == "" {
			return errResult(fmt.Errorf("tar entry with empty name"))
		}
		rel := relativeEntryName(entry.Header.Name, promptPath)
		if rel == "" {
			continue // directory prefix entry — will be skipped during write
		}
		if _, _, err := filesystem.VerifySubpath(skillDir, rel); err != nil {
			return errResult(fmt.Errorf("illegal file path in prompt layer: %s", entry.Header.Name))
		}
		hasWritableEntries = true
	}

	if !hasWritableEntries {
		return errResult(fmt.Errorf("prompt path '%s' does not match any tar entries in the layer", promptPath))
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

	// Write entries, stripping the prompt path prefix at write time
	for _, entry := range entries {
		rel := relativeEntryName(entry.Header.Name, promptPath)
		if rel == "" {
			continue
		}
		outPath := filepath.Join(skillDir, rel)

		switch entry.Header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, entry.Header.FileInfo().Mode()); err != nil {
				return errResult(fmt.Errorf("creating directory %s: %w", rel, err))
			}
		default: // tar.TypeReg
			if dir := filepath.Dir(outPath); dir != skillDir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errResult(fmt.Errorf("creating parent directory for %s: %w", rel, err))
				}
			}
			if err := os.WriteFile(outPath, entry.Content, entry.Header.FileInfo().Mode()); err != nil {
				return errResult(fmt.Errorf("writing file %s: %w", rel, err))
			}
		}
	}

	return AgentInstallResult{Agent: agent, Path: skillDir}
}
