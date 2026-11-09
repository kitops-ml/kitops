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

package kitinit

import (
	"fmt"
	"os"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/agent"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

// importAgent derives an Agent Definition from the agent configuration already
// present in the target directory. Detection is automatic: any project-scoped
// agent source produces an agent ModelKit.
func importAgent(opts *initOptions) (*agent.ImportResult, error) {
	if opts.ignoreAgent || opts.remote {
		return &agent.ImportResult{}, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Without a home directory local and user scope are unreadable, but
		// project scope still imports.
		output.Logf(output.LogLevelWarn, "Could not determine home directory; skipping local and user scope agent configuration: %s", err)
		homeDir = ""
	}
	return agent.Import(agent.ImportOptions{
		ContextDir:     opts.path,
		HomeDir:        homeDir,
		Framework:      opts.agentFramework,
		IncludeGlobals: opts.agentGlobals,
	})
}

// printAgentReport names every module created, every need declared, every
// literal dropped and every path packed, so the developer can check the import.
func printAgentReport(result *agent.ImportResult) {
	if !result.Found() {
		if len(result.NearMiss) > 0 {
			output.Infof("\nNo agent configuration detected: found %s, but no %s directory. It is packed as a prompt; no agent block was added.",
				strings.Join(result.NearMiss, ", "), strings.Join(agent.ProjectDirs(), " or "))
		}
		printSkipped(result)
		return
	}

	var b strings.Builder
	if result.DefinitionPath != "" {
		fmt.Fprintf(&b, "\nUsing the existing agent definition in %s. It is packed as-is.\n", result.DefinitionPath)
	} else {
		fmt.Fprintf(&b, "\nDetected agent configuration for %s\n", result.Framework)
	}

	if len(result.Sources) > 0 {
		b.WriteString("\nImported from\n")
		for _, source := range result.Sources {
			fmt.Fprintf(&b, "  %s\n", source)
		}
	}

	if result.Agent != nil && len(result.Agent.Includes) > 0 {
		b.WriteString("\nModules\n")
		for _, module := range result.Agent.Includes {
			fmt.Fprintf(&b, "  %-20s %s\n", module.Name, describeModule(module))
		}
	}

	if result.Agent != nil && len(result.Agent.Needs) > 0 {
		b.WriteString("\nNeeds\n")
		for _, need := range result.Agent.Needs {
			fmt.Fprintf(&b, "  %-20s %s\n", need.Name, need.Describe())
		}
	}

	if len(result.DroppedValues) > 0 {
		b.WriteString("\nValues dropped, not written to the Kitfile\n")
		for _, dropped := range result.DroppedValues {
			fmt.Fprintf(&b, "  %s\n", dropped)
		}
	}

	if len(result.PromptPaths) > 0 || len(result.RulePaths) > 0 || len(result.SkillPaths) > 0 || len(result.BundlePaths) > 0 {
		b.WriteString("\nAgent content packed in this ModelKit\n")
		for _, path := range result.PromptPaths {
			fmt.Fprintf(&b, "  %-40s prompt\n", path)
		}
		for _, path := range result.RulePaths {
			fmt.Fprintf(&b, "  %-40s rule\n", path)
		}
		for _, path := range result.SkillPaths {
			fmt.Fprintf(&b, "  %-40s skill\n", path)
		}
		for _, path := range result.BundlePaths {
			fmt.Fprintf(&b, "  %-40s mcp bundle\n", path)
		}
	}
	b.WriteString("\nThe rest of this directory is the agent's subject, not its contents, and is not packed.\nUse '--ignore-agent' to classify the directory as an ordinary ModelKit instead.\n")

	if len(result.UnvendoredServers) > 0 {
		fmt.Fprintf(&b, "\n%d local server(s) recorded as commands: %s.\nThey run as recorded by 'kit launch' here. Vendor them into MCP bundles before pushing.\n",
			len(result.UnvendoredServers), strings.Join(result.UnvendoredServers, ", "))
	}

	output.Infof("%s", b.String())

	for _, warning := range result.Warnings {
		output.Logf(output.LogLevelWarn, "%s", warning)
	}
	printSkipped(result)
}

// printSkipped reports the user-scope sources that were found and not imported.
func printSkipped(result *agent.ImportResult) {
	var b strings.Builder
	if len(result.SkippedUserEntries) > 0 {
		b.WriteString("\nSkipped, user scope describes your machine rather than this agent\n")
		for _, entry := range result.SkippedUserEntries {
			fmt.Fprintf(&b, "  %-20s %s\n", entry.Name, entry.Source)
		}
		b.WriteString("\nRun with '--agent-globals' to import them as project configuration.\n")
	}
	if len(result.SkippedUserSkills) > 0 {
		b.WriteString("\nSkipped, user-scope skills live outside the project directory\n")
		for _, path := range result.SkippedUserSkills {
			fmt.Fprintf(&b, "  %s\n", path)
		}
		b.WriteString("\nA packed path must stay inside the project directory, so these cannot be\npacked where they are. To share one across projects, pack it as its own\nModelKit ('kit init' then 'kit pack' in the skill's directory) and add it to\n'agent.includes' as a skill module. To tie it to this project alone, copy it in.\n")
	}
	if b.Len() > 0 {
		output.Infof("%s", b.String())
	}
}

func describeModule(module artifact.AgentModule) string {
	if module.Source == nil {
		return module.Type
	}
	switch {
	case module.Source.Endpoint != "":
		return fmt.Sprintf("remote  %s (%s)", module.Source.Endpoint, module.Transport)
	case module.Source.Command != "":
		command := strings.TrimSpace(module.Source.Command + " " + strings.Join(module.Source.Args, " "))
		return fmt.Sprintf("local   %s", command)
	case module.Source.ModelKit != "":
		if name, relative := artifact.RelativeModelKitName(module.Source.ModelKit); relative {
			return fmt.Sprintf("packed  %s (in this ModelKit)", name)
		}
		return fmt.Sprintf("modelkit %s", module.Source.ModelKit)
	default:
		return module.Type
	}
}

// addAgentToExistingKitfile adds the agent block to a Kitfile that is already
// present, leaving its other sections alone. Returns false when the caller
// should fall through to ordinary generation.
func addAgentToExistingKitfile(result *agent.ImportResult, opts *initOptions) (bool, error) {
	if result.Agent == nil || opts.overwrite || opts.outputPath == "-" {
		return false, nil
	}
	file, err := os.Open(opts.outputPath)
	if err != nil {
		// No existing Kitfile: ordinary generation applies.
		return false, nil
	}
	defer file.Close()

	existing := &artifact.KitFile{}
	if err := existing.LoadModel(file); err != nil {
		return false, output.Fatalf("Kitfile already exists at %s but could not be read: %s", opts.outputPath, err)
	}
	if existing.Agent != nil {
		return false, output.Fatalf("Kitfile at %s already contains an agent block. Use '--force' to regenerate it", opts.outputPath)
	}

	output.Infof("A Kitfile already exists at %s.", opts.outputPath)
	if util.IsInteractiveSession() {
		answer, err := util.PromptForInput("Add the detected agent configuration to it? [y/N]: ", false)
		if err != nil {
			return false, err
		}
		if !isAffirmative(answer) {
			return false, output.Fatalf("Cancelled; the existing Kitfile was not modified")
		}
	} else {
		return false, output.Fatalf("Refusing to modify the existing Kitfile without confirmation. Re-run in a terminal, or use '--force' to regenerate it")
	}

	existing.Agent = result.Agent
	result.AddContentModules(existing)
	bytes, err := existing.MarshalToYAML()
	if err != nil {
		return false, output.Fatalf("Error formatting Kitfile: %s", err)
	}
	output.Logf(output.LogLevelWarn, "Rewriting %s: comments and key ordering in the existing Kitfile are not preserved", opts.outputPath)

	opts.overwrite = true
	if err := writeKitfile(bytes, opts); err != nil {
		return false, err
	}
	printAgentReport(result)
	return true, nil
}

func isAffirmative(answer string) bool {
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
