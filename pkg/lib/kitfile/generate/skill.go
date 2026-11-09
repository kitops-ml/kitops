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

package generate

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/skill"
	"github.com/kitops-ml/kitops/pkg/output"
)

// parseSkillFrontmatter reads a SKILL.md file and parses its frontmatter.
// skillMDPath is relative to contextDir.
func parseSkillFrontmatter(contextDir, skillMDPath string) *skill.SkillFrontmatter {
	fullPath := filepath.Join(contextDir, skillMDPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		output.Logf(output.LogLevelWarn, "Failed to read %s: %s", fullPath, err)
		return nil
	}

	fm := skill.ParseSkillFrontmatter(data)
	if fm == nil || (fm.Name == "" && fm.Description == "") {
		output.Logf(output.LogLevelWarn, "Could not read a name or description from SKILL.md frontmatter at %s; the generated Kitfile will have no name or description", fullPath)
	}
	return fm
}

func dirContainsSkillMD(dir DirectoryListing) (bool, string) {
	for _, file := range dir.Files {
		if file.Name == skill.FileName {
			return true, file.Path
		}
	}
	return false, ""
}

func buildPromptFromSkill(dir DirectoryListing) (artifact.Prompt, *skill.SkillFrontmatter) {
	if found, _ := dirContainsSkillMD(dir); !found {
		return artifact.Prompt{Path: unixWithTrailingSlash(dir.Path)}, nil
	}
	return promptFromSkillDir(dir.ContextDir, dir.Path)
}

// promptFromSkillDir builds the prompt layer for a skill directory, naming and
// describing it from its SKILL.md frontmatter. skillDir is slash-separated and
// relative to contextDir.
func promptFromSkillDir(contextDir, skillDir string) (artifact.Prompt, *skill.SkillFrontmatter) {
	prompt := artifact.Prompt{Path: unixWithTrailingSlash(skillDir)}
	fm := parseSkillFrontmatter(contextDir, path.Join(skillDir, skill.FileName))
	if fm != nil {
		prompt.Name = fm.Name
		prompt.Description = fm.Description
	}
	return prompt, fm
}

func applySkillMetadataToPackage(kitfile *artifact.KitFile, dir DirectoryListing) {
	var skillFrontmatters []*skill.SkillFrontmatter
	for _, subDir := range dir.Subdirs {
		if found, skillPath := dirContainsSkillMD(subDir); found {
			if fm := parseSkillFrontmatter(subDir.ContextDir, skillPath); fm != nil {
				skillFrontmatters = append(skillFrontmatters, fm)
			}
		}
	}

	if len(skillFrontmatters) == 0 {
		return
	}

	// Only promote name/description when the Kitfile contains exclusively
	// prompt layers. A mixed Kitfile (model + skill) should not inherit the
	// skill's name as the package name.
	hasOnlyPrompts := kitfile.Model == nil && len(kitfile.Code) == 0 &&
		len(kitfile.DataSets) == 0 && len(kitfile.Docs) == 0
	if hasOnlyPrompts && len(skillFrontmatters) == 1 {
		fm := skillFrontmatters[0]
		if kitfile.Package.Name == "" {
			kitfile.Package.Name = fm.Name
		}
		if kitfile.Package.Description == "" {
			kitfile.Package.Description = fm.Description
		}
	}

	first := skillFrontmatters[0]
	if kitfile.Package.License == "" && first.License != "" {
		kitfile.Package.License = first.License
		output.Logf(output.LogLevelTrace, "Using license %q from skill %q", first.License, first.Name)
	}
}

// AgentContent is the content an agent-only ModelKit carries, as paths relative
// to the context directory.
type AgentContent struct {
	// Definition is the project's own standalone Agent Definition, if any.
	Definition string
	// Prompts are the framework's prompt files (CLAUDE.md, AGENTS.md).
	Prompts []string
	// Skills are skill directories, each holding a SKILL.md.
	Skills []string
	// Rules are rule files, such as .cursor/rules/*.
	Rules []string
	// Bundles are .mcpb files in the working directory.
	Bundles []string
}

// GenerateAgentKitfile builds the Kitfile for an agent-only ModelKit, listing
// each content path individually. The rest of the directory is the repository
// the agent works on, not the agent, so ordinary layer classification is not
// run and no whole-directory layer is emitted.
func GenerateAgentKitfile(content AgentContent, packageOpt *artifact.Package, contextDir string) *artifact.KitFile {
	kitfile := &artifact.KitFile{ManifestVersion: "1.0.0"}
	if packageOpt != nil {
		kitfile.Package = *packageOpt
	}

	if content.Definition != "" {
		kitfile.Code = append(kitfile.Code, artifact.Code{Path: content.Definition})
	}
	for _, prompt := range content.Prompts {
		kitfile.Prompts = append(kitfile.Prompts, artifact.Prompt{Path: prompt})
	}
	for _, rule := range content.Rules {
		kitfile.Prompts = append(kitfile.Prompts, artifact.Prompt{Path: rule})
	}
	AddAgentSkills(kitfile, contextDir, content.Skills)
	for _, bundle := range content.Bundles {
		kitfile.MCPServers = append(kitfile.MCPServers, mcpServerFromPath(bundle))
	}
	return kitfile
}

// AddAgentSkills records a project's agent skills as prompt layers. Paths are
// slash-separated and relative to contextDir. Every skill gets a name, from its
// SKILL.md frontmatter or else from its path, since the skill module that
// declares it refers to the prompts entry by name.
func AddAgentSkills(kitfile *artifact.KitFile, contextDir string, skillPaths []string) {
	existing := map[string]bool{}
	for _, prompt := range kitfile.Prompts {
		existing[strings.TrimSuffix(prompt.Path, "/")] = true
	}

	for _, skillPath := range skillPaths {
		if existing[strings.TrimSuffix(skillPath, "/")] {
			continue
		}
		fullPath := filepath.Join(contextDir, filepath.FromSlash(skillPath))
		info, err := os.Stat(fullPath)
		if err != nil {
			output.Logf(output.LogLevelWarn, "Skipping agent skill %s: %s", skillPath, err)
			continue
		}
		if !info.IsDir() {
			output.Logf(output.LogLevelWarn, "Skipping agent skill %s: not a directory", skillPath)
			continue
		}

		prompt, _ := promptFromSkillDir(contextDir, skillPath)
		if prompt.Name == "" {
			prompt.Name = path.Base(strings.TrimSuffix(skillPath, "/"))
		}
		kitfile.Prompts = append(kitfile.Prompts, prompt)
	}
}
