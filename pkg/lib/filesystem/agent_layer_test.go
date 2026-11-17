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

package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAgentDefinitionLayer(t *testing.T) {
	dir := t.TempDir()
	definition := "apiVersion: kitops.org/v1\nkind: AgentDefinition\nspec:\n  framework: claude-code\n"
	write := func(name, contents string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	plain := write("agent-definition.yaml", definition)
	hidden := write(".agent-definition.yaml", definition)
	otherKind := write("sub/agent-definition.yaml", "kind: Other\n")
	otherName := write("definition.yaml", definition)
	broken := write("broken/agent-definition.yaml", "kind: [\n")
	link := filepath.Join(dir, "linked", "agent-definition.yaml")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(plain, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{name: "definition file", path: plain, want: true},
		{name: "hidden definition file", path: hidden, want: true},
		{name: "directory holding a definition", path: dir},
		{name: "definition name, other kind", path: otherKind},
		{name: "definition document, other name", path: otherName},
		{name: "symlink to a definition", path: link},
		{name: "unparseable", path: broken, wantErr: true},
		{name: "missing", path: filepath.Join(dir, "missing", "agent-definition.yaml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isAgentDefinitionLayer(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
