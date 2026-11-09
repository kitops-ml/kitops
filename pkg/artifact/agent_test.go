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
	"io"
	"strings"
	"testing"
)

func loadKitfile(t *testing.T, contents string) (*KitFile, error) {
	t.Helper()
	kf := &KitFile{}
	err := kf.LoadModel(io.NopCloser(strings.NewReader(contents)))
	return kf, err
}

func TestAgentBlockValid(t *testing.T) {
	kf, err := loadKitfile(t, `
manifestVersion: 1.0.0
package:
  name: my-agent
prompts:
  - path: CLAUDE.md
  - name: review
    path: .claude/skills/review/
mcpServers:
  - name: postgres
    path: postgres.mcpb
model:
  name: llama3-8b
  path: models/llama3
agent:
  framework: claude-code
  needs:
    - name: GITHUB_TOKEN
      from:
        envFrom: GITHUB_TOKEN
    - name: LOG_LEVEL
      from:
        text: debug
    - name: AWS_WEB_IDENTITY_TOKEN_FILE
      target: file
      mountPath: /run/creds/token
      from:
        oidc:
          audience: sts.amazonaws.com
  includes:
    - name: github
      type: mcp
      source:
        endpoint: https://api.githubcopilot.com/mcp/
      transport: http
      headers:
        Authorization: Bearer ${needs.GITHUB_TOKEN}
    - name: filesystem
      type: mcp
      source:
        command: npx
        args: [-y, server-filesystem]
    - name: postgres
      type: mcp
      source:
        modelkit: ./postgres
    - name: primary-llm
      type: llm
      source:
        model: anthropic/claude-opus-4-5
      config:
        maxTokens: "8192"
    - name: strict-tools
      type: policy
      source:
        modelkit: jozu.com/policies/strict-tools:v1
    - name: review
      type: skill
      source:
        modelkit: ./review
    - name: local-llm
      type: llm
      source:
        modelkit: ./llama3-8b
`)
	if err != nil {
		t.Fatalf("expected a valid Kitfile, got: %s", err)
	}
	if kf.Agent == nil || kf.Agent.Framework != "claude-code" {
		t.Fatalf("agent block did not parse: %+v", kf.Agent)
	}
	if len(kf.Agent.Includes) != 7 || len(kf.Agent.Needs) != 3 {
		t.Errorf("got %d modules and %d needs", len(kf.Agent.Includes), len(kf.Agent.Needs))
	}
}

func TestAgentBlockValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		agent    string
		wantErr  string
		describe string
	}{
		{
			name:     "no framework and no base",
			agent:    "agent:\n  needs: []\n",
			wantErr:  "must specify a framework",
			describe: "a root definition has to name a framework",
		},
		{
			name:     "invalid need name",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: not-a-valid-name\n",
			wantErr:  "invalid name",
			describe: "need names are used as environment variable names",
		},
		{
			name:     "duplicate need",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: TOKEN\n    - name: TOKEN\n",
			wantErr:  "duplicate agent need",
			describe: "names must be unique",
		},
		{
			name:     "mountPath on env target",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: TOKEN\n      mountPath: /run/token\n",
			wantErr:  "must not set mountPath",
			describe: "mountPath belongs to file targets",
		},
		{
			name:     "file target without mountPath",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: TOKEN\n      target: file\n",
			wantErr:  "must set mountPath",
			describe: "a file target needs somewhere to land",
		},
		{
			name:     "two resolvers on one need",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: TOKEN\n      from:\n        envFrom: A\n        text: b\n",
			wantErr:  "more than one resolver",
			describe: "exactly one resolver kind per need",
		},
		{
			name:     "oidc without file target",
			agent:    "agent:\n  framework: claude-code\n  needs:\n    - name: TOKEN\n      from:\n        oidc:\n          audience: sts\n",
			wantErr:  "requires target 'file'",
			describe: "oidc writes a JWT to a path",
		},
		{
			name:     "unknown module type",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: widget\n",
			wantErr:  "invalid type",
			describe: "four module types are defined",
		},
		{
			name:     "remote mcp without transport",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        endpoint: https://e.example.com\n",
			wantErr:  "must specify transport",
			describe: "a remote endpoint needs a transport",
		},
		{
			name:     "local mcp with transport",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        command: npx\n      transport: http\n",
			wantErr:  "must not specify transport",
			describe: "local servers are always stdio",
		},
		{
			name:     "local mcp with headers",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        command: npx\n      headers:\n        A: b\n",
			wantErr:  "must not specify headers",
			describe: "headers apply to remote servers only",
		},
		{
			name:     "two mcp sources",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        command: npx\n        endpoint: https://e.example.com\n",
			wantErr:  "exactly one of source.modelkit, source.endpoint or source.command",
			describe: "a module refers to one thing",
		},
		{
			name:     "relative mcp reference names no entry",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        modelkit: ./missing\n",
			wantErr:  "is not an entry of mcpServers",
			describe: "the bundle has to be packed in this ModelKit",
		},
		{
			name:     "empty relative reference",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        modelkit: ./\n",
			wantErr:  "empty relative source.modelkit",
			describe: "the prefix alone names nothing",
		},
		{
			name:     "relative reference on a policy",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: policy\n      source:\n        modelkit: ./strict\n",
			wantErr:  "has no policy section",
			describe: "a Kitfile packs no policies",
		},
		{
			name:     "relative skill reference names no entry",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: skill\n      source:\n        modelkit: ./missing\n",
			wantErr:  "is not an entry of prompts",
			describe: "the skill has to be packed in this ModelKit",
		},
		{
			name:     "relative skill reference names two entries",
			agent:    "prompts:\n  - name: dup\n    path: a/\n  - name: dup\n    path: b/\nagent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: skill\n      source:\n        modelkit: ./dup\n",
			wantErr:  "names more than one entry of prompts",
			describe: "a module names one entry",
		},
		{
			name:     "skill with an endpoint",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: skill\n      source:\n        modelkit: a/b:v1\n        endpoint: https://e.example.com\n",
			wantErr:  "may only specify source.modelkit",
			describe: "skills come from ModelKits",
		},
		{
			name:     "relative llm reference with no model",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: llm\n      source:\n        modelkit: ./llama\n",
			wantErr:  "is not the name of this Kitfile's model",
			describe: "the model has to be packed in this ModelKit",
		},
		{
			name:     "relative llm reference with another name",
			agent:    "model:\n  name: other\n  path: m/\nagent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: llm\n      source:\n        modelkit: ./llama\n",
			wantErr:  "is not the name of this Kitfile's model",
			describe: "the reference is by the model's name",
		},
		{
			name:     "relative llm reference with an endpoint",
			agent:    "model:\n  name: llama\n  path: m/\nagent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: llm\n      source:\n        modelkit: ./llama\n        endpoint: https://e.example.com\n",
			wantErr:  "must not specify source.endpoint with a local",
			describe: "a served model has no remote endpoint",
		},
		{
			name:     "header references unknown need",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        endpoint: https://e.example.com\n      transport: http\n      headers:\n        Authorization: Bearer ${needs.MISSING}\n",
			wantErr:  "references unknown need",
			describe: "a reference must resolve",
		},
		{
			name:     "header with a non-needs template",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: mcp\n      source:\n        endpoint: https://e.example.com\n      transport: http\n      headers:\n        Authorization: Bearer ${SOME_VAR}\n",
			wantErr:  "only ${needs.NAME} references are permitted",
			describe: "the only interpolation is needs",
		},
		{
			name:     "policy with an endpoint",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: policy\n      source:\n        endpoint: https://e.example.com\n",
			wantErr:  "must specify source.modelkit",
			describe: "policies come from ModelKits",
		},
		{
			name:     "duplicate module name",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: x\n      type: policy\n      source:\n        modelkit: a/b:v1\n    - name: x\n      type: skill\n      source:\n        modelkit: a/c:v1\n",
			wantErr:  "duplicate agent module",
			describe: "names must be unique",
		},
		{
			name:     "removal entry without a base",
			agent:    "agent:\n  framework: claude-code\n  includes:\n    - name: github\n",
			wantErr:  "must have a type",
			describe: "removing an inherited module needs something to inherit from",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadKitfile(t, "manifestVersion: 1.0.0\npackage:\n  name: a\n"+tt.agent)
			if err == nil {
				t.Fatalf("expected an error (%s)", tt.describe)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %q (%s)", err, tt.wantErr, tt.describe)
			}
		})
	}
}

func TestAgentRemovalEntryWithBase(t *testing.T) {
	// A name-only entry is how a derived definition drops an inherited module.
	_, err := loadKitfile(t, `
manifestVersion: 1.0.0
package:
  name: derived
agent:
  base:
    modelkit: registry.example.com/org/base-agent:v1
  includes:
    - name: github
`)
	if err != nil {
		t.Fatalf("expected a valid composed definition, got: %s", err)
	}
}

func TestKitfileWithoutAgentBlockUnchanged(t *testing.T) {
	kf, err := loadKitfile(t, "manifestVersion: 1.0.0\npackage:\n  name: a\ncode:\n  - path: ./src\n")
	if err != nil {
		t.Fatal(err)
	}
	if kf.Agent != nil {
		t.Errorf("agent should be nil when no block is present, got %+v", kf.Agent)
	}
	data, err := kf.MarshalToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "agent") {
		t.Errorf("an agent key was emitted for a Kitfile that has none:\n%s", data)
	}
}
