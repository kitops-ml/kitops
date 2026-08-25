// Copyright 2024 The KitOps Authors.
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
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

//go:embed SKILL.md
var skillContent []byte

const (
	shortDesc = `Install a skill file to teach AI agents how to use the kit CLI`
	longDesc  = `The skill command installs a SKILL.md file into AI agent skill directories.

The installed skill teaches the agent about kit workflows, Kitfile authoring,
pack/push/pull/unpack operations, tagging, and common errors. The skill content
is embedded in the binary and always matches the installed kit version.

With no flags, the skill content is printed to stdout.`

	example = `# Print the skill content to stdout:
kit skill

# Install the skill for Claude Code (writes to .claude/skills/kit.md):
kit skill --agent claude

# Install the skill for a generic agent (writes to .agents/skills/kit.md):
kit skill --agent generic

# Install the skill to a custom directory:
kit skill --dir ./my-agent-dir

# Overwrite an existing skill file:
kit skill --agent claude --force`
)

var agentDirs = map[string]string{
	"claude":  filepath.Join(".claude", "skills"),
	"generic": filepath.Join(".agents", "skills"),
}

type skillOptions struct {
	agent string
	dir   string
	force bool
}

func SkillCommand() *cobra.Command {
	opts := &skillOptions{}

	cmd := &cobra.Command{
		Use:     "skill",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.NoArgs,
	}

	cmd.Flags().StringVar(&opts.agent, "agent", "", "Agent to install skill for (claude, generic)")
	cmd.RegisterFlagCompletionFunc("agent", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"claude", "generic"}, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().StringVar(&opts.dir, "dir", "", "Custom directory to write skill file into")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Overwrite existing skill file")
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *skillOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if opts.agent != "" && opts.dir != "" {
			return output.Fatalf("Invalid arguments: --agent and --dir are mutually exclusive")
		}

		if opts.agent != "" {
			skillDir, ok := agentDirs[opts.agent]
			if !ok {
				return output.Fatalf("Unknown agent %q: supported values are claude, generic", opts.agent)
			}
			return writeSkill(skillDir, opts.force)
		}

		if opts.dir != "" {
			return writeSkill(opts.dir, opts.force)
		}

		// No flags: print to stdout
		fmt.Fprint(cmd.OutOrStdout(), string(skillContent))
		return nil
	}
}

func writeSkill(dir string, force bool) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return output.Fatalf("Failed to create directory %s: %s", dir, err)
	}

	destPath := filepath.Join(dir, "kit.md")

	if _, err := os.Stat(destPath); err == nil {
		if !force {
			return output.Fatalf("Skill file already exists at %s. Use '--force' to overwrite", destPath)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return output.Fatalf("Error checking for existing skill file: %s", err)
	}

	if err := os.WriteFile(destPath, skillContent, 0o644); err != nil {
		return output.Fatalf("Failed to write skill file to %s: %s", destPath, err)
	}

	output.Infof("Skill installed to %s", destPath)
	return nil
}
