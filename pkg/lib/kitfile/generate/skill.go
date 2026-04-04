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

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/skill"
	"github.com/kitops-ml/kitops/pkg/output"
)

func parseSkillFrontmatter(skillMDPath string) *skill.SkillFrontmatter {
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		output.Logf(output.LogLevelWarn, "Failed to read %s: %s", skillMDPath, err)
		return nil
	}

	fm := skill.ParseSkillFrontmatter(data)
	if fm == nil {
		output.Logf(output.LogLevelWarn, "No valid frontmatter in %s", skillMDPath)
	}
	return fm
}

func dirContainsSkillMD(dir DirectoryListing) (bool, string) {
	for _, file := range dir.Files {
		if file.Name == "SKILL.md" {
			return true, file.Path
		}
	}
	return false, ""
}

func buildPromptFromSkill(dir DirectoryListing) (artifact.Prompt, *skill.SkillFrontmatter) {
	prompt := artifact.Prompt{
		Path: dir.Path,
	}

	found, skillPath := dirContainsSkillMD(dir)
	if !found {
		return prompt, nil
	}

	fm := parseSkillFrontmatter(skillPath)
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
			if fm := parseSkillFrontmatter(skillPath); fm != nil {
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
