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
	"fmt"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

// AddContentModules adds to the generated agent block an mcp module for each
// mcpServers entry of the Kitfile and a skill module for each prompts entry
// that is an imported skill, each with a relative source.modelkit naming the
// entry. A bundle named like a local server imported as a command is that
// server's vendored form and replaces the command; a bundle named like a
// remote server is left undeclared and reported.
func (r *ImportResult) AddContentModules(kf *artifact.KitFile) {
	if r.Agent == nil || kf == nil {
		return
	}

	for _, server := range kf.MCPServers {
		if server.Name == "" {
			continue
		}
		module := artifact.AgentModule{
			Name:   server.Name,
			Type:   artifact.AgentModuleMCP,
			Source: &artifact.AgentSource{ModelKit: artifact.RelativeModelKitPrefix + server.Name},
		}
		idx := r.moduleIndex(server.Name)
		if idx < 0 {
			r.Agent.Includes = append(r.Agent.Includes, module)
			continue
		}
		existing := r.Agent.Includes[idx]
		if existing.Type == artifact.AgentModuleMCP && existing.Source != nil && existing.Source.Command != "" {
			module.Config = existing.Config
			r.Agent.Includes[idx] = module
			r.UnvendoredServers = slices.DeleteFunc(r.UnvendoredServers, func(name string) bool { return name == server.Name })
			continue
		}
		r.Warnings = append(r.Warnings, fmt.Sprintf("MCP bundle %s has the same name as the imported module %q; the bundle is packed but not declared in agent.includes. Rename one of them to declare both", server.Path, server.Name))
	}

	skillPaths := map[string]bool{}
	for _, path := range r.SkillPaths {
		skillPaths[strings.TrimSuffix(path, "/")] = true
	}
	declared := map[string]bool{}
	for _, prompt := range kf.Prompts {
		if !skillPaths[strings.TrimSuffix(prompt.Path, "/")] {
			continue
		}
		if prompt.Name == "" {
			r.Warnings = append(r.Warnings, fmt.Sprintf("skill %s has no name, so no skill module declares it; give it a name in its prompts entry and add the module", prompt.Path))
			continue
		}
		if declared[prompt.Name] {
			continue
		}
		if countPromptsNamed(kf, prompt.Name) > 1 {
			r.Warnings = append(r.Warnings, fmt.Sprintf("more than one prompts entry is named %q, so no skill module declares them; a module names one entry", prompt.Name))
			declared[prompt.Name] = true
			continue
		}
		name := prompt.Name
		if r.moduleIndex(name) >= 0 {
			name += "-skill"
			if r.moduleIndex(name) >= 0 {
				r.Warnings = append(r.Warnings, fmt.Sprintf("skill %q shares its name with another module and no skill module declares it; rename one of them and add the module", prompt.Name))
				continue
			}
		}
		r.Agent.Includes = append(r.Agent.Includes, artifact.AgentModule{
			Name:   name,
			Type:   artifact.AgentModuleSkill,
			Source: &artifact.AgentSource{ModelKit: artifact.RelativeModelKitPrefix + prompt.Name},
		})
		declared[prompt.Name] = true
	}
}

// moduleIndex returns the position of the module by name, or -1.
func (r *ImportResult) moduleIndex(name string) int {
	for i, module := range r.Agent.Includes {
		if module.Name == name {
			return i
		}
	}
	return -1
}

func countPromptsNamed(kf *artifact.KitFile, name string) int {
	n := 0
	for _, prompt := range kf.Prompts {
		if prompt.Name == name {
			n++
		}
	}
	return n
}
