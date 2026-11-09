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
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/skill"
)

// ImportOptions controls how kit init derives an Agent Definition from a
// project's existing agent configuration.
type ImportOptions struct {
	// ContextDir is the directory being initialized. Discovery examines only
	// the fixed paths in the framework table relative to this directory; it does
	// not search subdirectories for nested agent projects.
	ContextDir string
	// HomeDir is the user's home directory, where a framework may keep
	// project-bound (local) and personal (user) scope configuration.
	HomeDir string
	// Framework overrides framework selection. Empty selects by detection.
	Framework string
	// IncludeGlobals imports user-scope entries, promoting the developer's
	// personal tools to team dependencies.
	IncludeGlobals bool
}

// ImportResult is everything kit init learned, so the caller can build the
// Kitfile and report accurately what was and was not imported.
type ImportResult struct {
	// Agent is the definition to write, or nil when no agent sources were found
	// or when the project already carries its own definition.
	Agent *artifact.Agent
	// DefinitionPath is the project's own standalone Agent Definition, relative
	// to the context directory, when it has one. It is packed as-is: kit init
	// never rewrites it and never writes a second definition beside it.
	DefinitionPath string
	// Framework is the selected framework's name.
	Framework string
	// Sources are the files entries were read from, for reporting.
	Sources []string
	// SkippedUserEntries are user-scope entries that were found and not
	// imported.
	SkippedUserEntries []ServerEntry
	// SkippedUserSkills are user-scope skill directories. They are listed and
	// never imported, since a Kitfile path cannot leave the context directory.
	SkippedUserSkills []string
	// UnvendoredServers are local stdio servers recorded as commands. They run
	// as recorded when kit launch runs in this working directory, and must be
	// vendored into MCPB bundles before the ModelKit leaves the machine.
	UnvendoredServers []string
	// DroppedValues names each place a literal was dropped and replaced by a
	// needs reference.
	DroppedValues []string
	// SkillPaths, RulePaths and PromptPaths are project content, relative to
	// the context directory, that travels in the ModelKit's prompt layers.
	SkillPaths  []string
	RulePaths   []string
	PromptPaths []string
	// BundlePaths are .mcpb files in the working directory, carried as
	// mcpServers entries.
	BundlePaths []string
	// Warnings are non-fatal problems the developer should see.
	Warnings []string
	// NearMiss names agent-adjacent files found when no framework was selected,
	// so a developer who expected an agent ModelKit learns why they did not get
	// one.
	NearMiss []string
}

// Found reports whether any agent configuration was discovered.
func (r *ImportResult) Found() bool {
	return r != nil && (r.Agent != nil || r.DefinitionPath != "")
}

var nonNameChars = regexp.MustCompile(`[^A-Za-z0-9]+`)

// Import derives an Agent Definition from the agent configuration already
// present in a project. It performs no network access and reads only the fixed
// paths named in the framework table.
func Import(opts ImportOptions) (*ImportResult, error) {
	definition, definitionFramework, err := existingDefinition(opts.ContextDir)
	if err != nil {
		return nil, err
	}
	if definition != "" {
		return importExistingDefinition(opts, definition, definitionFramework)
	}

	framework, err := selectFramework(opts)
	if err != nil {
		return nil, err
	}
	if framework == nil {
		return &ImportResult{NearMiss: nearMisses(opts.ContextDir)}, nil
	}

	result := &ImportResult{Framework: framework.Name}
	builder := &definitionBuilder{result: result}

	projectEntries, err := framework.ProjectEntries(opts.ContextDir)
	if err != nil {
		return nil, err
	}
	if len(projectEntries) > 0 {
		result.Sources = append(result.Sources, framework.ProjectMCPConfig)
	}

	localEntries, err := framework.LocalEntries(opts.HomeDir, opts.ContextDir)
	if err != nil {
		return nil, err
	}
	if len(localEntries) > 0 {
		result.Sources = append(result.Sources, fmt.Sprintf("%s (local scope for this project)", framework.LocalMCPConfig))
	}

	userEntries, err := framework.UserEntries(opts.HomeDir)
	if err != nil {
		return nil, err
	}
	// A framework that keeps local and user scope in one file reports the same
	// entries for both; entries bound to this project are local, not user.
	userEntries = excludeByName(userEntries, localEntries)

	for _, entry := range projectEntries {
		builder.addServer(entry)
	}
	for _, entry := range localEntries {
		builder.addServer(entry)
	}
	if opts.IncludeGlobals {
		for _, entry := range userEntries {
			builder.addServer(entry)
		}
		if len(userEntries) > 0 {
			result.Sources = append(result.Sources, fmt.Sprintf("%s (user scope)", framework.UserMCPConfig))
		}
	} else {
		result.SkippedUserEntries = userEntries
	}

	if err := collectContent(result, framework, opts.ContextDir); err != nil {
		return nil, err
	}

	// A framework was selected, so the project is an agent project: the block is
	// emitted even when it carries no modules, recording the framework and
	// giving the developer somewhere to add them.
	needs := builder.needs
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })
	result.Agent = &artifact.Agent{
		Framework: framework.Name,
		Needs:     needs,
		Includes:  builder.modules,
	}
	return result, nil
}

// importExistingDefinition packages a project's own Agent Definition as it is,
// without importing framework configuration, and still collects the agent's
// content. The definition's framework decides whose content to collect; a
// definition that names no known framework falls back to detection.
func importExistingDefinition(opts ImportOptions, definition, definitionFramework string) (*ImportResult, error) {
	result := &ImportResult{DefinitionPath: definition}

	framework := FrameworkByName(definitionFramework)
	if framework == nil || opts.Framework != "" {
		var err error
		if framework, err = selectFramework(opts); err != nil {
			return nil, err
		}
	}
	if framework != nil {
		result.Framework = framework.Name
	}
	if err := collectContent(result, framework, opts.ContextDir); err != nil {
		return nil, err
	}
	return result, nil
}

// collectContent records the agent content the project carries: the selected
// framework's prompts, rules and skills, and any MCP bundles.
func collectContent(result *ImportResult, framework *Framework, contextDir string) error {
	if framework != nil {
		result.SkillPaths = framework.projectSkills(contextDir)
		result.RulePaths = framework.projectRules(contextDir)
		var unpackablePrompts []string
		result.PromptPaths, unpackablePrompts = framework.projectPrompts(contextDir)
		for _, prompt := range unpackablePrompts {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s is a symlink; it identifies the framework but is not packed. Its target travels under its own name", prompt))
		}
		result.SkippedUserSkills = framework.userSkills()
	}
	bundles, err := mcpBundles(contextDir)
	if err != nil {
		return err
	}
	result.BundlePaths = bundles
	return nil
}

// selectFramework picks the framework whose project-scoped sources are present,
// in table order. An explicit override skips detection but must name a row.
func selectFramework(opts ImportOptions) (*Framework, error) {
	if opts.Framework != "" {
		framework := FrameworkByName(opts.Framework)
		if framework == nil {
			return nil, fmt.Errorf("unknown framework %q: known frameworks are %s",
				opts.Framework, strings.Join(FrameworkNames(), ", "))
		}
		return framework, nil
	}
	for i := range Frameworks {
		found, err := Frameworks[i].hasProjectSources(opts.ContextDir, opts.HomeDir)
		if err != nil {
			return nil, err
		}
		if found {
			return &Frameworks[i], nil
		}
	}
	return nil, nil
}

// hasProjectSources reports whether the project carries this framework's
// configuration: its directory, an MCP config, a skills directory, or servers
// in local scope. A prompt file alone is not enough, since CLAUDE.md and
// AGENTS.md appear in plenty of repositories that are not agents.
func (f Framework) hasProjectSources(contextDir, homeDir string) (bool, error) {
	if f.ProjectDir != "" && dirExists(filepath.Join(contextDir, filepath.FromSlash(f.ProjectDir))) {
		return true, nil
	}
	if f.ProjectMCPConfig != "" && fileExists(filepath.Join(contextDir, filepath.FromSlash(f.ProjectMCPConfig))) {
		return true, nil
	}
	if len(f.projectSkills(contextDir)) > 0 {
		return true, nil
	}
	// A framework may write an empty local-scope entry for every directory it
	// has run in, so only a non-empty server map counts.
	localEntries, err := f.LocalEntries(homeDir, contextDir)
	if err != nil {
		return false, err
	}
	return len(localEntries) > 0, nil
}

// projectSkills lists the project's skill directories: the subdirectories of
// the framework's skills directories that hold a SKILL.md.
func (f Framework) projectSkills(contextDir string) []string {
	var skills []string
	for _, skillDir := range f.SkillDirs() {
		root := filepath.Join(contextDir, filepath.FromSlash(skillDir))
		for _, name := range skillEntryNames(root) {
			skills = append(skills, skillDir+"/"+name)
		}
	}
	sort.Strings(skills)
	return skills
}

// userSkills lists the skills in the framework's user-scope skills directory,
// as absolute paths.
func (f Framework) userSkills() []string {
	root := f.UserSkillsDir()
	if root == "" {
		return nil
	}
	var skills []string
	for _, name := range skillEntryNames(root) {
		skills = append(skills, filepath.ToSlash(filepath.Join(root, name)))
	}
	return skills
}

// skillEntryNames returns the subdirectories of a skills directory that hold a
// SKILL.md.
func skillEntryNames(root string) []string {
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var names []string
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() && fileExists(filepath.Join(root, dirEntry.Name(), skill.FileName)) {
			names = append(names, dirEntry.Name())
		}
	}
	sort.Strings(names)
	return names
}

// projectRules lists the rule files in the framework's rules directories.
func (f Framework) projectRules(contextDir string) []string {
	var rules []string
	for _, ruleDir := range f.ProjectRules {
		root := filepath.Join(contextDir, filepath.FromSlash(ruleDir))
		dirEntries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, dirEntry := range dirEntries {
			if dirEntry.Type().IsRegular() {
				rules = append(rules, normalizePath(ruleDir)+"/"+dirEntry.Name())
			}
		}
	}
	sort.Strings(rules)
	return rules
}

// projectPrompts lists the framework's prompt files present in the project,
// and separately those kit will not pack: kit packs regular files and
// directories only, so a symlinked prompt does not travel.
func (f Framework) projectPrompts(contextDir string) (prompts, unpackable []string) {
	for _, prompt := range f.Prompts {
		path := filepath.Join(contextDir, filepath.FromSlash(prompt))
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		switch {
		case info.Mode().IsRegular() || info.IsDir():
			prompts = append(prompts, normalizePath(prompt))
		default:
			unpackable = append(unpackable, normalizePath(prompt))
		}
	}
	return prompts, unpackable
}

// nearMisses names agent-adjacent files present when no framework was selected,
// so a developer who expected an agent ModelKit learns why they did not get one.
func nearMisses(contextDir string) []string {
	var found []string
	for _, framework := range Frameworks {
		for _, prompt := range framework.Prompts {
			if _, err := os.Lstat(filepath.Join(contextDir, filepath.FromSlash(prompt))); err == nil {
				found = append(found, normalizePath(prompt))
			}
		}
	}
	sort.Strings(found)
	return slices.Compact(found)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// definitionBuilder accumulates modules and needs, keeping names unique and
// recording every place a literal was dropped.
type definitionBuilder struct {
	result      *ImportResult
	modules     []artifact.AgentModule
	needs       []artifact.AgentNeed
	moduleNames map[string]bool
	needIndex   map[string]int
}

func (b *definitionBuilder) addServer(entry ServerEntry) {
	if b.moduleNames == nil {
		b.moduleNames = map[string]bool{}
	}
	if b.moduleNames[entry.Name] {
		// An entry in a narrower scope already claimed this name; project beats
		// local beats user, and the entries arrive in that order.
		b.warnf("MCP server %q from %s scope is shadowed by an entry of the same name in a narrower scope; skipping", entry.Name, entry.Scope)
		return
	}

	module := artifact.AgentModule{Name: entry.Name, Type: artifact.AgentModuleMCP}
	switch {
	case entry.IsRemote():
		module.Source = &artifact.AgentSource{Endpoint: entry.URL}
		module.Transport = entry.Transport()
		module.Headers = b.headersFor(entry)
	case entry.Command != "":
		module.Source = &artifact.AgentSource{Command: entry.Command, Args: entry.Args}
		// A local server reads its own environment; its config's env map becomes
		// needs, which the resolver delivers to the launched process.
		b.needsFromEnv(entry)
		b.result.UnvendoredServers = append(b.result.UnvendoredServers, entry.Name)
	default:
		b.warnf("MCP server %q in %s declares neither a command nor a url; skipping", entry.Name, entry.Source)
		return
	}

	b.moduleNames[entry.Name] = true
	b.modules = append(b.modules, module)
}

// needsFromEnv converts a local server's env map into needs. A value that is
// plainly configuration is preserved as a literal; anything that could be a
// credential is replaced by an environment lookup and the literal dropped.
func (b *definitionBuilder) needsFromEnv(entry ServerEntry) {
	for _, name := range sortedKeys(entry.Env) {
		value := entry.Env[name]
		if !agentNeedName.MatchString(name) {
			b.warnf("environment variable %q on MCP server %q is not a valid need name; skipping", name, entry.Name)
			continue
		}
		if ref, ok := EnvVarReference(value); ok {
			b.addNeed(artifact.AgentNeed{Name: name, From: &artifact.AgentNeedFrom{EnvFrom: ref}}, entry.Name)
		} else if LooksLikeSecret(name, value) {
			b.addNeed(artifact.AgentNeed{Name: name, From: &artifact.AgentNeedFrom{EnvFrom: name}}, entry.Name)
			b.result.DroppedValues = append(b.result.DroppedValues,
				fmt.Sprintf("%s (env of MCP server %s)", name, entry.Name))
		} else {
			b.addNeed(artifact.AgentNeed{Name: name, From: &artifact.AgentNeedFrom{Text: value}}, entry.Name)
		}
	}
}

// headersFor rewrites a remote server's headers so that no credential survives
// as a literal. A dropped value becomes a ${needs.NAME} reference the resolver
// substitutes at deploy time.
func (b *definitionBuilder) headersFor(entry ServerEntry) map[string]string {
	if len(entry.Headers) == 0 {
		return nil
	}
	headers := map[string]string{}
	for _, header := range sortedKeys(entry.Headers) {
		value := entry.Headers[header]
		scheme, credential := SplitAuthScheme(value)

		if ref, ok := EnvVarReference(credential); ok {
			b.addNeed(artifact.AgentNeed{Name: ref, From: &artifact.AgentNeedFrom{EnvFrom: ref}}, entry.Name)
			headers[header] = scheme + "${needs." + ref + "}"
			continue
		}
		if strings.Contains(credential, "${needs.") {
			// Already a reference; keep it and make sure the need is declared.
			for _, name := range needsReferences(credential) {
				b.addNeed(artifact.AgentNeed{Name: name, From: &artifact.AgentNeedFrom{EnvFrom: name}}, entry.Name)
			}
			headers[header] = value
			continue
		}
		if IsCredentialName(header) || IsSecretValue(credential) {
			name := needNameForHeader(entry.Name, header)
			b.addNeed(artifact.AgentNeed{Name: name, From: &artifact.AgentNeedFrom{EnvFrom: name}}, entry.Name)
			headers[header] = scheme + "${needs." + name + "}"
			b.result.DroppedValues = append(b.result.DroppedValues,
				fmt.Sprintf("%s header of MCP server %s", header, entry.Name))
			continue
		}
		headers[header] = value
	}
	return headers
}

// addNeed records a need, keeping the first declaration when two servers want
// the same name. A conflicting second declaration is reported rather than
// silently overwriting the first.
func (b *definitionBuilder) addNeed(need artifact.AgentNeed, module string) {
	if b.needIndex == nil {
		b.needIndex = map[string]int{}
	}
	if idx, exists := b.needIndex[need.Name]; exists {
		existing := b.needs[idx]
		if !sameResolver(existing.From, need.From) {
			b.warnf("need %s is declared differently by more than one MCP server; keeping %s and ignoring the declaration from %q",
				need.Name, existing.Describe(), module)
		}
		return
	}
	b.needIndex[need.Name] = len(b.needs)
	b.needs = append(b.needs, need)
}

func (b *definitionBuilder) warnf(format string, args ...any) {
	b.result.Warnings = append(b.result.Warnings, fmt.Sprintf(format, args...))
}

var agentNeedName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var needsRefRegexp = regexp.MustCompile(`\$\{needs\.([A-Za-z_][A-Za-z0-9_]*)\}`)

func needsReferences(value string) []string {
	var names []string
	for _, match := range needsRefRegexp.FindAllStringSubmatch(value, -1) {
		names = append(names, match[1])
	}
	return names
}

// needNameForHeader derives a stable need name from the server and header. An
// Authorization header becomes <SERVER>_TOKEN, the name a developer would
// choose; any other header keeps its own name for traceability.
func needNameForHeader(server, header string) string {
	base := strings.ToUpper(nonNameChars.ReplaceAllString(server, "_"))
	suffix := "TOKEN"
	if !strings.EqualFold(header, "authorization") {
		suffix = strings.ToUpper(nonNameChars.ReplaceAllString(header, "_"))
	}
	name := strings.Trim(base+"_"+suffix, "_")
	if !agentNeedName.MatchString(name) {
		name = "NEED_" + name
	}
	return name
}

func sameResolver(a, b *artifact.AgentNeedFrom) bool {
	return reflect.DeepEqual(a, b)
}

// excludeByName drops entries whose name appears in other, so a config file
// that holds both user and project-bound scope does not report the same server
// twice.
func excludeByName(entries, other []ServerEntry) []ServerEntry {
	if len(other) == 0 {
		return entries
	}
	names := map[string]bool{}
	for _, entry := range other {
		names[entry.Name] = true
	}
	var kept []ServerEntry
	for _, entry := range entries {
		if !names[entry.Name] {
			kept = append(kept, entry)
		}
	}
	return kept
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// mcpBundles lists the .mcpb files in the working directory; each becomes an
// mcpServers entry.
func mcpBundles(contextDir string) ([]string, error) {
	dirEntries, err := os.ReadDir(contextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", contextDir, err)
	}
	var bundles []string
	for _, dirEntry := range dirEntries {
		if dirEntry.Type().IsRegular() && strings.HasSuffix(dirEntry.Name(), ".mcpb") {
			bundles = append(bundles, dirEntry.Name())
		}
	}
	sort.Strings(bundles)
	return bundles, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
