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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Scope is where an MCP entry was configured, which decides whether kit init
// imports it. The line between imported and ignored is scope, not file
// location: Claude Code's local scope lives in the user's home directory but is
// bound to one project, so it is imported.
type Scope string

const (
	// ScopeProject is configuration committed with the project.
	ScopeProject Scope = "project"
	// ScopeLocal is configuration bound to this project but stored in the user's
	// home directory.
	ScopeLocal Scope = "local"
	// ScopeUser describes the developer's personal environment, not the agent.
	ScopeUser Scope = "user"
)

// ServerEntry is one entry from a framework's mcpServers map.
type ServerEntry struct {
	Name  string
	Scope Scope
	// Source is the file the entry was read from, for reporting.
	Source string

	// Local (stdio) server fields.
	Command string
	Args    []string
	Env     map[string]string

	// Remote server fields.
	URL     string
	Headers map[string]string

	// Type is the transport declared in the config, if any.
	Type string
}

// IsRemote reports whether the entry describes a remote server.
func (e ServerEntry) IsRemote() bool {
	return e.URL != ""
}

// Transport returns the transport to record for a remote entry, honoring an
// explicitly declared type and otherwise inferring from the URL.
func (e ServerEntry) Transport() string {
	switch strings.ToLower(e.Type) {
	case "http", "streamable-http", "streamablehttp":
		return "http"
	case "sse":
		return "sse"
	}
	if strings.HasSuffix(strings.TrimSuffix(e.URL, "/"), "/sse") {
		return "sse"
	}
	return "http"
}

// rawServerEntry is the on-disk shape of an mcpServers entry. Both the stdio
// and remote shapes are decoded; which fields are set decides the kind.
type rawServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Type    string            `json:"type"`
}

// mcpConfigFile is the shape kit init reads from a framework MCP config. Files
// that carry more than this (Claude Code's ~/.claude.json holds conversation
// history as well) decode the parts named here and discard the rest.
type mcpConfigFile struct {
	MCPServers map[string]rawServerEntry `json:"mcpServers"`
	Projects   map[string]struct {
		MCPServers map[string]rawServerEntry `json:"mcpServers"`
	} `json:"projects"`
}

// readMCPConfig parses an MCP config file. A file that does not exist yields no
// entries and no error; a file that exists but cannot be parsed is an error,
// because silently importing nothing from a config the developer wrote would be
// worse than failing.
func readMCPConfig(path string) (*mcpConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	config := &mcpConfigFile{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return config, nil
}

// entriesFromMap converts a decoded mcpServers map into entries, sorted by name
// so that repeated runs produce the same Kitfile.
func entriesFromMap(servers map[string]rawServerEntry, scope Scope, source string) []ServerEntry {
	entries := make([]ServerEntry, 0, len(servers))
	for name, raw := range servers {
		entries = append(entries, ServerEntry{
			Name:    name,
			Scope:   scope,
			Source:  source,
			Command: raw.Command,
			Args:    raw.Args,
			Env:     raw.Env,
			URL:     raw.URL,
			Headers: raw.Headers,
			Type:    raw.Type,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

// ProjectEntries reads the framework's project-scoped MCP config from contextDir.
func (f Framework) ProjectEntries(contextDir string) ([]ServerEntry, error) {
	if f.ProjectMCPConfig == "" {
		return nil, nil
	}
	path := filepath.Join(contextDir, filepath.FromSlash(f.ProjectMCPConfig))
	config, err := readMCPConfig(path)
	if err != nil || config == nil {
		return nil, err
	}
	return entriesFromMap(config.MCPServers, ScopeProject, f.ProjectMCPConfig), nil
}

// LocalEntries reads the entries in the framework's home-directory config that
// are bound to contextDir. These are project-bound despite living in the user's
// home, so they are imported.
func (f Framework) LocalEntries(homeDir, contextDir string) ([]ServerEntry, error) {
	if f.LocalMCPConfig == "" {
		return nil, nil
	}
	path := filepath.Join(homeDir, filepath.FromSlash(f.LocalMCPConfig))
	config, err := readMCPConfig(path)
	if err != nil || config == nil {
		return nil, err
	}
	for key, project := range config.Projects {
		if !samePath(key, contextDir) {
			continue
		}
		return entriesFromMap(project.MCPServers, ScopeLocal, path), nil
	}
	return nil, nil
}

// UserEntries reads the framework's user-scope entries. They describe the
// developer's personal environment rather than the agent, so kit init lists
// them and imports them only when asked.
func (f Framework) UserEntries(homeDir string) ([]ServerEntry, error) {
	if f.UserMCPConfig == "" {
		return nil, nil
	}
	path := filepath.Join(homeDir, filepath.FromSlash(f.UserMCPConfig))
	config, err := readMCPConfig(path)
	if err != nil || config == nil {
		return nil, err
	}
	return entriesFromMap(config.MCPServers, ScopeUser, path), nil
}

// samePath reports whether two paths name the same directory. The comparison
// resolves symlinks and, on platforms with case-insensitive filesystems by
// default, ignores case: a project key written by another tool may differ from
// the path kit was invoked with while naming the same directory.
func samePath(a, b string) bool {
	clean := func(p string) string {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			p = resolved
		}
		return filepath.Clean(p)
	}
	cleanA, cleanB := clean(a), clean(b)
	if cleanA == cleanB {
		return true
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return strings.EqualFold(cleanA, cleanB)
	}
	return false
}
