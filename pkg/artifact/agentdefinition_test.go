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
	"strings"
	"testing"
)

func TestDetectAgentDefinition(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		wantFound     bool
		wantFramework string
		wantErr       bool
	}{
		{
			name:          "definition",
			yaml:          "apiVersion: kitops.org/v1\nkind: AgentDefinition\nspec:\n  framework: cursor\n",
			wantFound:     true,
			wantFramework: "cursor",
		},
		{
			name:          "second document is the definition",
			yaml:          "kind: Other\n---\nkind: AgentDefinition\nspec:\n  framework: claude-code\n",
			wantFound:     true,
			wantFramework: "claude-code",
		},
		{
			name: "other kind",
			yaml: "kind: Deployment\n",
		},
		{
			name: "empty",
			yaml: "",
		},
		{
			name:    "unparseable",
			yaml:    "kind: [\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, framework, err := DetectAgentDefinition(strings.NewReader(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if found != tt.wantFound || framework != tt.wantFramework {
				t.Errorf("got (%v, %q), want (%v, %q)", found, framework, tt.wantFound, tt.wantFramework)
			}
		})
	}
}
