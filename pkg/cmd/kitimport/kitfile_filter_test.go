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

package kitimport

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKitfilePathFilterCatchallExcludesKitfile(t *testing.T) {
	// Catchall code layer that would otherwise sweep up everything.
	kf := &artifact.KitFile{
		Code: []artifact.Code{{Path: "."}},
	}
	filter, err := newKitfilePathFilter(kf)
	require.NoError(t, err)

	// Regular files in the catchall are packed.
	got, err := filter.Matches("README.md")
	require.NoError(t, err)
	assert.True(t, got, "README.md should be packed by the catchall layer")

	// Kitfile and .kitignore are always ignore-pattern-matched and must NOT
	// appear in the packed set even with a catchall layer.
	for _, name := range []string{"Kitfile", ".kitignore"} {
		got, err := filter.Matches(name)
		require.NoError(t, err, name)
		assert.False(t, got, "%s should never be reported as packed (it's always ignored)", name)
	}
}

func TestKitfilePathFilterSpecificLayer(t *testing.T) {
	kf := &artifact.KitFile{
		Model: &artifact.Model{Path: "model_dir"},
	}
	filter, err := newKitfilePathFilter(kf)
	require.NoError(t, err)

	in, err := filter.Matches("model_dir/weights.bin")
	require.NoError(t, err)
	assert.True(t, in, "files under model_dir belong to the model layer")

	out, err := filter.Matches("README.md")
	require.NoError(t, err)
	assert.False(t, out, "files outside any layer must not be reported as packed")

	// Boundary: "model_directory" must not be confused with "model_dir".
	sibling, err := filter.Matches("model_directory/foo")
	require.NoError(t, err)
	assert.False(t, sibling, "sibling-of-prefix paths must not match the layer")
}

func TestPathUnderLayerNormalizesSeparators(t *testing.T) {
	// On Windows, repoutil.LayerPathsFromKitfile (filepath.Clean) rewrites
	// nested layer paths to use backslashes, while file paths reach the filter
	// as forward-slash. pathUnderLayer must normalize both before comparing or
	// non-catchall imports break on Windows.
	tests := []struct {
		path, layer string
		want        bool
	}{
		// Windows-cleaned layer with forward-slash file paths.
		{path: "model_dir/sub/weights.bin", layer: "model_dir\\sub", want: true},
		{path: "other_dir/file.bin", layer: "model_dir\\sub", want: false},
		// Mixed-separator layer.
		{path: "a/b/c.bin", layer: "a\\b", want: true},
		// Boundary check survives normalization.
		{path: "model_directory/file", layer: "model_dir", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.path+" under "+tt.layer, func(t *testing.T) {
			assert.Equal(t, tt.want, pathUnderLayer(tt.path, tt.layer))
		})
	}
}

func TestKitfilePathFilterMultipleLayersWithCatchall(t *testing.T) {
	// model_dir owns its files; "." catchall picks up everything else.
	kf := &artifact.KitFile{
		Model: &artifact.Model{Path: "model_dir"},
		Code:  []artifact.Code{{Path: "."}},
	}
	filter, err := newKitfilePathFilter(kf)
	require.NoError(t, err)

	for _, p := range []string{"model_dir/weights.bin", "src/main.py", "README.md"} {
		got, err := filter.Matches(p)
		require.NoError(t, err, p)
		assert.True(t, got, "%s should be packed (catchall + model layer cover everything)", p)
	}

	// Kitfile / .kitignore still excluded by ignore patterns despite the catchall.
	for _, name := range []string{"Kitfile", ".kitignore"} {
		got, err := filter.Matches(name)
		require.NoError(t, err, name)
		assert.False(t, got, "%s must remain excluded", name)
	}
}
