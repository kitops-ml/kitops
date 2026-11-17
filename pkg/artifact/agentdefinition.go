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
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

const (
	// AgentDefinitionFileName is the conventional name of a standalone Agent
	// Definition file, packed in its own code layer.
	AgentDefinitionFileName = "agent-definition.yaml"
	// HiddenAgentDefinitionFileName is the dot-prefixed spelling of the same
	// file, for projects that keep tool configuration out of the directory listing.
	HiddenAgentDefinitionFileName = "." + AgentDefinitionFileName

	agentDefinitionKind = "AgentDefinition"
)

// IsAgentDefinitionFileName reports whether name is one of the two accepted
// spellings of a standalone Agent Definition file.
func IsAgentDefinitionFileName(name string) bool {
	return name == AgentDefinitionFileName || name == HiddenAgentDefinitionFileName
}

// DetectAgentDefinition reports whether the YAML read from r holds an
// AgentDefinition document, and that document's framework. A file may hold
// several documents; the first AgentDefinition is the one a resolver consumes.
func DetectAgentDefinition(r io.Reader) (isDefinition bool, framework string, err error) {
	decoder := yaml.NewDecoder(r)
	for {
		var doc struct {
			Kind string `yaml:"kind"`
			Spec struct {
				Framework string `yaml:"framework"`
			} `yaml:"spec"`
		}
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return false, "", nil
			}
			return false, "", fmt.Errorf("failed to parse agent definition: %w", err)
		}
		if doc.Kind == agentDefinitionKind {
			return true, doc.Spec.Framework, nil
		}
	}
}
