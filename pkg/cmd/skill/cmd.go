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
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	libskill "github.com/kitops-ml/kitops/pkg/lib/skill"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

// embeddedSkillMD is the SKILL.md that teaches an AI agent how to use the kit
// CLI. It is embedded in the binary so it always matches the installed kit
// version and works without network access.
//
//go:embed SKILL.md
var embeddedSkillMD []byte

// skillName is the directory name the skill is installed under, matching the
// "name" field in the embedded SKILL.md frontmatter.
const skillName = "kitops"

const (
	shortDesc = `Install the kit agent skill so AI agents can use the kit CLI`
	longDesc  = `Install a SKILL.md that teaches an AI agent how to use the kit CLI.

The skill documents Kitfile authoring and the pack, tag, login, push, pull, and
unpack workflow. It is embedded in the kit binary, so it always matches the
installed kit version and works without network access.

The skill is written into each agent's skills directory (for example
'.claude/skills' or the shared '.agents/skills'). Without --agents, kit
auto-discovers installed agents by checking their global config directories.
With --agents, specify agents as a comma-separated list (e.g.
--agents=claude-code,cursor). By default the skill is installed globally
(user-scoped); when -d is specified it is installed into that project
directory instead.`

	example = `# Install the kit skill for auto-detected agents (user-scoped)
kit skill

# Install for specific agents
kit skill --agents=claude-code,cursor

# Install into a project directory
kit skill -d /path/to/project

# Overwrite an existing installation
kit skill -o`
)

type skillOptions struct {
	projectDir     string
	agents         string
	overwrite      bool
	ignoreExisting bool
}

// SkillCommand returns the `kit skill` cobra command.
func SkillCommand() *cobra.Command {
	opts := &skillOptions{}

	cmd := &cobra.Command{
		Use:     "skill [flags]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		Args:    cobra.NoArgs,
		RunE:    runCommand(opts),
	}

	cmd.Flags().StringVarP(&opts.projectDir, "dir", "d", "", "Install the skill into this project directory instead of globally (user-scoped)")
	cmd.Flags().StringVar(&opts.agents, "agents", "", "Agents to install the skill for, as a comma-separated list (e.g. claude-code,cursor). Without a value, auto-discovers installed agents")
	cmd.Flags().BoolVarP(&opts.overwrite, "overwrite", "o", false, "Overwrite the skill if it already exists")
	cmd.Flags().BoolVarP(&opts.ignoreExisting, "ignore-existing", "i", false, "Skip installation if the skill already exists")
	cmd.Flags().SortFlags = false

	cmd.RegisterFlagCompletionFunc("agents", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return libskill.ValidAgentNames(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveDefault)
	return cmd
}

func runCommand(opts *skillOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		agents, err := resolveAgents(opts.agents)
		if err != nil {
			return output.Fatalf("%s", err)
		}

		// Only scope to a project directory when -d was explicitly provided;
		// otherwise install globally (user-scoped), matching `kit unpack --as-skill`.
		projectDir := ""
		if cmd.Flags().Changed("dir") {
			abs, err := filepath.Abs(opts.projectDir)
			if err != nil {
				return output.Fatalf("failed to resolve absolute path %s: %s", opts.projectDir, err)
			}
			projectDir = abs
		}

		installOpts := &libskill.SkillInstallOptions{
			Agents:         agents,
			ProjectDir:     projectDir,
			Overwrite:      opts.overwrite,
			IgnoreExisting: opts.ignoreExisting,
		}

		// Wrap the embedded SKILL.md as a single tar entry so the shared
		// installer (used by `kit unpack --as-skill`) resolves the agent skills
		// directories and writes the file.
		entries := []libskill.TarEntry{
			{
				Header: &tar.Header{
					Name:     "SKILL.md",
					Typeflag: tar.TypeReg,
					Mode:     0644,
					Size:     int64(len(embeddedSkillMD)),
				},
				Content: embeddedSkillMD,
			},
		}

		result := libskill.InstallSkill(entries, skillName, artifact.Prompt{Path: "SKILL.md"}, installOpts)
		for _, ar := range result.Agents {
			switch {
			case ar.Err != nil:
				output.Infof("Failed to install skill '%s' for %s: %s", skillName, ar.Agent, ar.Err)
			case ar.Skipped:
				output.Infof("Skipped skill '%s' for %s: already exists (use -o to overwrite)", skillName, ar.Agent)
			default:
				output.Infof("Installed skill '%s' for %s → %s", skillName, ar.Agent, ar.Path)
			}
		}

		if result.HasErrors() {
			return output.Fatalf("failed to install skill for %d agent(s)", len(result.Errors()))
		}
		return nil
	}
}

// resolveAgents mirrors `kit unpack --as-skill` agent resolution: an empty value
// auto-discovers installed agents, otherwise it parses and validates a
// comma-separated list of agent names.
func resolveAgents(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		detected, err := libskill.DetectInstalledAgents()
		if err != nil {
			return nil, fmt.Errorf("failed to detect installed agents: %w", err)
		}
		if len(detected) == 0 {
			return nil, fmt.Errorf("no installed agents detected. Specify agents explicitly, e.g. --agents=claude-code,cursor")
		}
		output.Infof("Detected installed agents: %s", strings.Join(detected, ", "))
		return detected, nil
	}

	var agents []string
	for _, p := range strings.Split(value, ",") {
		name := strings.TrimSpace(p)
		if name == "" {
			return nil, fmt.Errorf("invalid agent list: empty agent name in '%s'", value)
		}
		if !libskill.IsValidAgentName(name) {
			return nil, fmt.Errorf("unknown agent '%s'. Valid agents: %s", name, strings.Join(libskill.ValidAgentNames(), ", "))
		}
		agents = append(agents, name)
	}
	return agents, nil
}
