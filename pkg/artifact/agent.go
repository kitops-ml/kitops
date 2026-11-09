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

package artifact

import (
	"fmt"
	"regexp"
	"strings"
)

// Agent is the Agent Definition carried inline in a Kitfile. It holds the same
// content as the spec: block of a standalone agent-definition.yaml document;
// apiVersion, kind and metadata are omitted because the Kitfile already
// supplies them (metadata.name comes from package.name).
type Agent struct {
	Framework string        `json:"framework,omitempty" yaml:"framework,omitempty"`
	Base      *AgentBase    `json:"base,omitempty" yaml:"base,omitempty"`
	Needs     []AgentNeed   `json:"needs,omitempty" yaml:"needs,omitempty"`
	Includes  []AgentModule `json:"includes,omitempty" yaml:"includes,omitempty"`
}

// AgentBase references another agent ModelKit to compose from.
type AgentBase struct {
	ModelKit string `json:"modelkit,omitempty" yaml:"modelkit,omitempty"`
}

// AgentNeed declares a runtime value the agent consumes, bound to a resolver.
type AgentNeed struct {
	Name      string         `json:"name,omitempty" yaml:"name,omitempty"`
	Target    string         `json:"target,omitempty" yaml:"target,omitempty"`
	MountPath string         `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	Optional  bool           `json:"optional,omitempty" yaml:"optional,omitempty"`
	From      *AgentNeedFrom `json:"from,omitempty" yaml:"from,omitempty"`
}

// AgentNeedFrom selects exactly one resolver kind for a need.
type AgentNeedFrom struct {
	EnvFrom  string             `json:"envFrom,omitempty" yaml:"envFrom,omitempty"`
	File     string             `json:"file,omitempty" yaml:"file,omitempty"`
	Text     string             `json:"text,omitempty" yaml:"text,omitempty"`
	Keychain *AgentKeychainFrom `json:"keychain,omitempty" yaml:"keychain,omitempty"`
	Vault    *AgentVaultFrom    `json:"vault,omitempty" yaml:"vault,omitempty"`
	OIDC     *AgentOIDCFrom     `json:"oidc,omitempty" yaml:"oidc,omitempty"`
}

type AgentKeychainFrom struct {
	Service string `json:"service,omitempty" yaml:"service,omitempty"`
	Account string `json:"account,omitempty" yaml:"account,omitempty"`
}

type AgentVaultFrom struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	Field string `json:"field,omitempty" yaml:"field,omitempty"`
}

type AgentOIDCFrom struct {
	Audience string `json:"audience,omitempty" yaml:"audience,omitempty"`
}

// AgentModule is one entry in the agent's includes list. A module with only a
// name is a removal entry, meaningful when this definition composes from a base.
type AgentModule struct {
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	Type      string            `json:"type,omitempty" yaml:"type,omitempty"`
	Source    *AgentSource      `json:"source,omitempty" yaml:"source,omitempty"`
	Transport string            `json:"transport,omitempty" yaml:"transport,omitempty"`
	Headers   map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// Config is opaque and passed through to the module. The same yaml caveats
	// as Model.Parameters apply.
	Config any `json:"config,omitempty" yaml:"config,omitempty"`
}

// AgentSource locates the thing a module refers to. Exactly one of Model,
// ModelKit, Endpoint or Command may be set.
type AgentSource struct {
	Model string `json:"model,omitempty" yaml:"model,omitempty"`
	// ModelKit is a reference to a ModelKit in a registry, or, when it begins
	// with RelativeModelKitPrefix, the name of an entry of this Kitfile: an
	// mcpServers entry for an mcp module, a prompts entry for a skill module,
	// the model for an llm module. The relative form declares content packed
	// in the same ModelKit as the agent.
	ModelKit string `json:"modelkit,omitempty" yaml:"modelkit,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// Command and Args describe a local stdio MCP server discovered from an
	// existing framework config. An explicit vendoring step replaces them with an
	// .mcpb bundle before the ModelKit leaves the machine.
	Command string   `json:"command,omitempty" yaml:"command,omitempty"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty,flow"`
}

const (
	AgentModuleLLM    = "llm"
	AgentModuleMCP    = "mcp"
	AgentModulePolicy = "policy"
	AgentModuleSkill  = "skill"

	AgentNeedTargetEnv  = "env"
	AgentNeedTargetFile = "file"

	// RelativeModelKitPrefix marks a source.modelkit that names an entry of
	// this Kitfile rather than a ModelKit in a registry. No registry reference
	// can begin with it, since a repository name starts with an alphanumeric.
	RelativeModelKitPrefix = "./"
)

// RelativeModelKitName returns the entry name a relative source.modelkit
// refers to, and whether the reference is relative at all.
func RelativeModelKitName(ref string) (name string, relative bool) {
	return strings.CutPrefix(ref, RelativeModelKitPrefix)
}

// agentNeedNameRegexp matches names safe to use as environment variable names.
var agentNeedNameRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// agentNeedsRefRegexp matches the ${needs.NAME} interpolation permitted in
// remote MCP header values.
var agentNeedsRefRegexp = regexp.MustCompile(`\$\{([^}]*)\}`)

// validateAgent appends errors for the Kitfile's agent block, if any.
func (kf *KitFile) validateAgent(addErr func(format string, a ...any)) {
	agent := kf.Agent
	if agent == nil {
		return
	}

	if agent.Base != nil && agent.Base.ModelKit == "" {
		addErr("agent base must specify a modelkit reference")
	}
	if agent.Framework == "" && agent.Base == nil {
		addErr("agent must specify a framework, or a base to inherit one from")
	}

	needNames := map[string]bool{}
	for _, need := range agent.Needs {
		switch {
		case need.Name == "":
			addErr("agent need must have a name")
			continue
		case !agentNeedNameRegexp.MatchString(need.Name):
			addErr("agent need %s has an invalid name: names must match [A-Za-z_][A-Za-z0-9_]*", need.Name)
		case needNames[need.Name]:
			addErr("duplicate agent need %s: names must be unique", need.Name)
		}
		needNames[need.Name] = true

		switch need.Target {
		case "", AgentNeedTargetEnv:
			if need.MountPath != "" {
				addErr("agent need %s must not set mountPath: mountPath applies to target 'file'", need.Name)
			}
		case AgentNeedTargetFile:
			if need.MountPath == "" {
				addErr("agent need %s must set mountPath when target is 'file'", need.Name)
			}
		default:
			addErr("agent need %s has invalid target %q: must be 'env' or 'file'", need.Name, need.Target)
		}

		if n := need.From.count(); n > 1 {
			addErr("agent need %s declares more than one resolver: exactly one of envFrom, file, text, keychain, vault or oidc is permitted", need.Name)
		} else if n == 0 && need.From != nil {
			addErr("agent need %s declares an empty from: block", need.Name)
		}
		if need.From != nil && need.From.OIDC != nil && need.Target != AgentNeedTargetFile {
			addErr("agent need %s uses the oidc resolver, which requires target 'file'", need.Name)
		}
	}

	moduleNames := map[string]bool{}
	for _, module := range agent.Includes {
		if module.Name == "" {
			addErr("agent module must have a name")
			continue
		}
		if moduleNames[module.Name] {
			addErr("duplicate agent module %s: names must be unique", module.Name)
		}
		moduleNames[module.Name] = true

		if module.Type == "" {
			if agent.Base == nil {
				addErr("agent module %s must have a type", module.Name)
			} else if module.Source != nil || module.Transport != "" || len(module.Headers) > 0 || module.Config != nil {
				addErr("agent module %s has no type but declares other fields: a name-only entry removes an inherited module", module.Name)
			}
			continue
		}

		switch module.Type {
		case AgentModuleLLM:
			kf.validateAgentLLM(module, addErr)
		case AgentModuleMCP:
			kf.validateAgentMCP(module, needNames, addErr)
		case AgentModulePolicy:
			if module.Source == nil || module.Source.ModelKit == "" {
				addErr("agent module %s (policy) must specify source.modelkit", module.Name)
			} else if _, relative := RelativeModelKitName(module.Source.ModelKit); relative {
				addErr("agent module %s (policy) has a relative source.modelkit %q: a Kitfile has no policy section to refer to", module.Name, module.Source.ModelKit)
			}
			if module.Source != nil && (module.Source.Model != "" || module.Source.Endpoint != "" || module.Source.Command != "") {
				addErr("agent module %s (policy) may only specify source.modelkit", module.Name)
			}
		case AgentModuleSkill:
			kf.validateAgentSkill(module, addErr)
		default:
			addErr("agent module %s has invalid type %q: must be one of llm, mcp, policy, skill", module.Name, module.Type)
		}

		if module.Type != AgentModuleMCP {
			if module.Transport != "" {
				addErr("agent module %s (%s) must not specify transport", module.Name, module.Type)
			}
			if len(module.Headers) > 0 {
				addErr("agent module %s (%s) must not specify headers", module.Name, module.Type)
			}
		}
	}
}

func (kf *KitFile) validateAgentLLM(module AgentModule, addErr func(format string, a ...any)) {
	src := module.Source
	if src == nil || (src.Model == "" && src.ModelKit == "") {
		addErr("agent module %s (llm) must specify source.model or source.modelkit", module.Name)
		return
	}
	if src.Model != "" && src.ModelKit != "" {
		addErr("agent module %s (llm) must specify exactly one of source.model or source.modelkit", module.Name)
	}
	if src.Command != "" {
		addErr("agent module %s (llm) must not specify source.command", module.Name)
	}
	if src.ModelKit != "" && src.Endpoint != "" {
		addErr("agent module %s (llm) must not specify source.endpoint with a local source.modelkit", module.Name)
	}
	if name, relative := RelativeModelKitName(src.ModelKit); relative {
		if name == "" {
			addErr("agent module %s (llm) has an empty relative source.modelkit: name this Kitfile's model, as ./<name>", module.Name)
		} else if kf.Model == nil || kf.Model.Name != name {
			addErr("agent module %s (llm) has relative source.modelkit %q, which is not the name of this Kitfile's model", module.Name, src.ModelKit)
		}
	}
}

func (kf *KitFile) validateAgentSkill(module AgentModule, addErr func(format string, a ...any)) {
	src := module.Source
	if src == nil || src.ModelKit == "" {
		addErr("agent module %s (skill) must specify source.modelkit", module.Name)
		return
	}
	if src.Model != "" || src.Endpoint != "" || src.Command != "" {
		addErr("agent module %s (skill) may only specify source.modelkit", module.Name)
	}
	if name, relative := RelativeModelKitName(src.ModelKit); relative {
		switch {
		case name == "":
			addErr("agent module %s (skill) has an empty relative source.modelkit: name a prompts entry, as ./<name>", module.Name)
		case kf.countPrompts(name) == 0:
			addErr("agent module %s (skill) has relative source.modelkit %q, which is not an entry of prompts", module.Name, src.ModelKit)
		case kf.countPrompts(name) > 1:
			addErr("agent module %s (skill) has relative source.modelkit %q, which names more than one entry of prompts", module.Name, src.ModelKit)
		}
	}
}

func (kf *KitFile) validateAgentMCP(module AgentModule, needNames map[string]bool, addErr func(format string, a ...any)) {
	src := module.Source
	if src == nil {
		addErr("agent module %s (mcp) must specify a source", module.Name)
		return
	}
	switch countSet(src.ModelKit, src.Endpoint, src.Command) {
	case 0:
		addErr("agent module %s (mcp) must specify one of source.modelkit, source.endpoint or source.command", module.Name)
		return
	case 1:
	default:
		addErr("agent module %s (mcp) must specify exactly one of source.modelkit, source.endpoint or source.command", module.Name)
	}
	if src.Model != "" {
		addErr("agent module %s (mcp) must not specify source.model", module.Name)
	}
	if name, relative := RelativeModelKitName(src.ModelKit); relative {
		if name == "" {
			addErr("agent module %s (mcp) has an empty relative source.modelkit: name an mcpServers entry, as ./<name>", module.Name)
		} else if !kf.hasMCPServer(name) {
			addErr("agent module %s (mcp) has relative source.modelkit %q, which is not an entry of mcpServers", module.Name, src.ModelKit)
		}
	}
	if len(src.Args) > 0 && src.Command == "" {
		addErr("agent module %s (mcp) specifies source.args without source.command", module.Name)
	}

	isRemote := src.Endpoint != ""
	if isRemote {
		switch module.Transport {
		case "http", "sse":
		case "":
			addErr("agent module %s (mcp) must specify transport ('http' or 'sse') with a remote source.endpoint", module.Name)
		default:
			addErr("agent module %s (mcp) has invalid transport %q: must be 'http' or 'sse'", module.Name, module.Transport)
		}
	} else if module.Transport != "" {
		addErr("agent module %s (mcp) must not specify transport: local servers always use stdio", module.Name)
	}

	if len(module.Headers) > 0 && !isRemote {
		addErr("agent module %s (mcp) must not specify headers: headers apply to remote servers only", module.Name)
	}
	for header, value := range module.Headers {
		for _, match := range agentNeedsRefRegexp.FindAllStringSubmatch(value, -1) {
			name, found := strings.CutPrefix(match[1], "needs.")
			if !found {
				addErr("agent module %s header %s contains %s: only ${needs.NAME} references are permitted", module.Name, header, match[0])
				continue
			}
			if !needNames[name] {
				addErr("agent module %s header %s references unknown need %s", module.Name, header, name)
			}
		}
	}
}

// hasMCPServer reports whether the Kitfile's mcpServers has an entry by name.
func (kf *KitFile) hasMCPServer(name string) bool {
	for _, server := range kf.MCPServers {
		if server.Name == name {
			return true
		}
	}
	return false
}

// countPrompts returns how many of the Kitfile's prompts carry the name.
func (kf *KitFile) countPrompts(name string) int {
	n := 0
	for _, prompt := range kf.Prompts {
		if prompt.Name == name {
			n++
		}
	}
	return n
}

// countSet returns how many of the strings are non-empty.
func countSet(values ...string) int {
	n := 0
	for _, value := range values {
		if value != "" {
			n++
		}
	}
	return n
}

// count returns the number of resolver kinds set on the need.
func (from *AgentNeedFrom) count() int {
	if from == nil {
		return 0
	}
	n := 0
	for _, set := range []bool{
		from.EnvFrom != "", from.File != "", from.Text != "",
		from.Keychain != nil, from.Vault != nil, from.OIDC != nil,
	} {
		if set {
			n++
		}
	}
	return n
}

// Describe returns a short human-readable description of where a need's value
// comes from. It never includes the value itself except for text, which is
// non-secret by construction.
func (need AgentNeed) Describe() string {
	switch {
	case need.From == nil:
		return fmt.Sprintf("environment variable %s", need.Name)
	case need.From.EnvFrom != "":
		return fmt.Sprintf("environment variable %s", need.From.EnvFrom)
	case need.From.File != "":
		return fmt.Sprintf("file %s", need.From.File)
	case need.From.Text != "":
		return "literal value"
	case need.From.Keychain != nil:
		return fmt.Sprintf("keychain %s/%s", need.From.Keychain.Service, need.From.Keychain.Account)
	case need.From.Vault != nil:
		return fmt.Sprintf("vault %s#%s", need.From.Vault.Path, need.From.Vault.Field)
	case need.From.OIDC != nil:
		return fmt.Sprintf("oidc token for %s", need.From.OIDC.Audience)
	default:
		return "unknown resolver"
	}
}
