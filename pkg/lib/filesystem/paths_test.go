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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySubpath(t *testing.T) {
	contextDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(contextDir, "data..v2"), 0700))

	tests := []struct {
		subDir  string
		relPath string
		wantErr bool
	}{
		{subDir: ".", relPath: "."},
		{subDir: "data", relPath: "data"},
		{subDir: "data..v2", relPath: "data..v2"},
		{subDir: "weights..bin", relPath: "weights..bin"},
		{subDir: "dir/notes...txt", relPath: filepath.Join("dir", "notes...txt")},
		{subDir: "..", wantErr: true},
		{subDir: "../outside", wantErr: true},
		{subDir: "data/../../outside", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.subDir, func(t *testing.T) {
			_, relPath, err := VerifySubpath(contextDir, tt.subDir)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.relPath, relPath)
		})
	}
}
