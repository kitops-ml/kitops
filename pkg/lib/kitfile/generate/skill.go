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
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/output"

	"go.yaml.in/yaml/v3"
)

type SkillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	License     string `yaml:"license,omitempty"`
}

func parseSkillFrontmatter(skillMDPath string) *SkillFrontmatter {
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		output.Logf(output.LogLevelWarn, "Failed to read %s: %s", skillMDPath, err)
		return nil
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return nil
	}

	end := strings.Index(content[4:], "\n---")
	if end == -1 {
		return nil
	}

	frontmatterYAML := content[4 : 4+end]
	if strings.TrimSpace(frontmatterYAML) == "" {
		return nil
	}

	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		output.Logf(output.LogLevelWarn, "Malformed frontmatter in %s: %s", skillMDPath, err)
		return nil
	}
	return &fm
}

func dirContainsSkillMD(dir DirectoryListing) (bool, string) {
	for _, file := range dir.Files {
		if strings.EqualFold(file.Name, "skill.md") {
			return true, file.Path
		}
	}
	return false, ""
}

func buildPromptFromSkill(dir DirectoryListing) artifact.Prompt {
	prompt := artifact.Prompt{
		Path: dir.Path,
	}

	found, skillPath := dirContainsSkillMD(dir)
	if !found {
		return prompt
	}

	fm := parseSkillFrontmatter(skillPath)
	if fm != nil {
		prompt.Name = fm.Name
		prompt.Description = fm.Description
	}
	return prompt
}

func applySkillMetadataToPackage(kitfile *artifact.KitFile, dir DirectoryListing) {
	var skillFrontmatters []*SkillFrontmatter
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

	if len(skillFrontmatters) == 1 {
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
		output.Logf(output.LogLevelWarn, "Using license from skill %q", first.Name)
	}
}
