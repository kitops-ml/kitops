// Copyright 2025 The KitOps Authors.
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

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "https://user:token@github.com/org/repo", want: "https://github.com/org/repo"},
		{in: "https://user@github.com/org/repo", want: "https://github.com/org/repo"},
		{in: "https://github.com/org/repo", want: "https://github.com/org/repo"},
		{in: "https://huggingface.co/org/repo", want: "https://huggingface.co/org/repo"},
		// '@' in the path portion is not userinfo and must be preserved.
		{in: "https://example.com/path@v1", want: "https://example.com/path@v1"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeURL(tt.in))
		})
	}
}

func TestResolveHead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}

	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		require.NoError(t, cmd.Run(), "git %v failed", args)
	}

	run("init", "-q")
	run("config", "commit.gpgsign", "false")
	run("config", "tag.gpgsign", "false")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0o644))
	run("add", ".")
	run("commit", "-q", "-m", "init")

	head, err := ResolveHead(dir)
	require.NoError(t, err)
	assert.Len(t, head, 40)
}
