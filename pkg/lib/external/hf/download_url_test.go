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

package hf

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscapeFilePath(t *testing.T) {
	tests := []string{
		"config.json",
		"dir/sub/model.safetensors",
		"dir/with space.txt",
		"notes#1.md",
		"what?.txt",
		"100%.bin",
		"data/a+b;c=d.csv",
		"données/modèle.bin",
	}
	for _, filePath := range tests {
		t.Run(filePath, func(t *testing.T) {
			fileURL := fmt.Sprintf(modelResolveURLFmt, "org/repo", "abc123", escapeFilePath(filePath))
			req, err := http.NewRequest(http.MethodGet, fileURL, nil)
			require.NoError(t, err)
			assert.Equal(t, "/org/repo/resolve/abc123/"+filePath, req.URL.Path)
			assert.Empty(t, req.URL.RawQuery)
			assert.Empty(t, req.URL.Fragment)
		})
	}
}
