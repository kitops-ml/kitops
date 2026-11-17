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

import "testing"

func TestLooksLikeSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
		why   string
	}{
		// Credential-shaped names, whatever the value.
		{"GITHUB_TOKEN", "abc", true, "name says token"},
		{"apiKey", "x", true, "name says key"},
		{"DB_PASSWORD", "hunter2", true, "name says password"},
		{"DB_PASS", "correcthorsebatterystaple", true, "name says pass; the value alone would not"},
		{"AWS_SECRET_ACCESS_KEY", "y", true, "name says secret"},
		{"AUTH_HEADER", "z", true, "name says auth"},

		// Known issuer prefixes.
		{"GH", "ghp_A1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8", true, "github token prefix"},
		{"OPENAI", "sk-proj-A1b2C3d4E5f6G7h8I9j0", true, "openai prefix"},
		{"SLACK", "xoxb-1234-5678-abcdefgh", true, "slack prefix"},

		// Credentials embedded in a connection string, with an innocuous name.
		{"POSTGRES_URL", "postgres://app:hunter2@localhost:5432/app", true, "userinfo carries a password"},
		{"MONGO_URI", "mongodb+srv://user:p4ss@cluster0.example.net/db", true, "userinfo carries a password"},

		// Plain configuration that must survive as a literal.
		{"ALLOWED_PATHS", "/workspace", false, "a path"},
		{"LOG_LEVEL", "debug", false, "a word"},
		{"NODE_ENV", "production", false, "a word"},
		{"PGDATABASE", "analytics", false, "a word"},
		{"HOME_DIR", "/Users/someone/workspaces/a-long-project-path", false, "a long path is still a path"},
		{"API_URL", "https://api.example.com/v1/endpoint", false, "a url without userinfo"},
		{"SERVER_NAME", "modelcontextprotocol-server-filesystem", false, "no digits, so not encoded"},
		{"PORT", "5432", false, "too short"},
		{"DESCRIPTION", "a sentence with several words in it", false, "prose"},
		{"PATHS", "/workspace,/tmp,/var/data", false, "a list"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"="+tt.value, func(t *testing.T) {
			if got := LooksLikeSecret(tt.name, tt.value); got != tt.want {
				t.Errorf("LooksLikeSecret(%q, %q) = %v, want %v (%s)", tt.name, tt.value, got, tt.want, tt.why)
			}
		})
	}
}

func TestEnvVarReference(t *testing.T) {
	tests := []struct {
		value string
		want  string
		ok    bool
	}{
		{"${GITHUB_TOKEN}", "GITHUB_TOKEN", true},
		{"$GITHUB_TOKEN", "GITHUB_TOKEN", true},
		{"  ${API_KEY}  ", "API_KEY", true},
		{"Bearer ${TOKEN}", "", false},
		{"literal", "", false},
		{"${needs.TOKEN}", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, ok := EnvVarReference(tt.value)
			if ok != tt.ok || got != tt.want {
				t.Errorf("EnvVarReference(%q) = %q, %v; want %q, %v", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestSplitAuthScheme(t *testing.T) {
	scheme, credential := SplitAuthScheme("Bearer ghp_secret")
	if scheme != "Bearer " || credential != "ghp_secret" {
		t.Errorf("SplitAuthScheme(Bearer ghp_secret) = %q, %q", scheme, credential)
	}
	scheme, credential = SplitAuthScheme("raw-value")
	if scheme != "" || credential != "raw-value" {
		t.Errorf("SplitAuthScheme(raw-value) = %q, %q", scheme, credential)
	}
}
