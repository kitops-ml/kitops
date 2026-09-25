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

package index

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLabelEdits(t *testing.T) {
	edits, err := ParseLabelEdits([]string{
		"quantization=q4_0",
		"source=https://example.com/?a=b",
		"empty=",
		"activeParameters:=7000000000",
		`gpuArchs:=[ "sm_80", "sm_90" ]`,
		"stale-",
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]json.RawMessage{
		"quantization":     json.RawMessage(`"q4_0"`),
		"source":           json.RawMessage(`"https://example.com/?a=b"`),
		"empty":            json.RawMessage(`""`),
		"activeParameters": json.RawMessage(`7000000000`),
		"gpuArchs":         json.RawMessage(`["sm_80","sm_90"]`),
	}, edits.Set)
	assert.Equal(t, []string{"stale"}, edits.Remove)
}

func TestParseLabelEditsRejectsBadInput(t *testing.T) {
	for arg, msg := range map[string]string{
		"quantization": "expected key=value to set, or key- to remove",
		"=value":       "key cannot be empty",
		"-":            "key cannot be empty",
		"has space=x":  "invalid label key",
		"gpuArchs:=[":  "value is not valid JSON",
	} {
		_, err := ParseLabelEdits([]string{arg})
		assert.ErrorContains(t, err, msg, arg)
	}
	_, err := ParseLabelEdits([]string{"z=1", "z-"})
	assert.ErrorContains(t, err, "each key can be specified only once")
}

func TestParseAnnotationEdits(t *testing.T) {
	edits, err := ParseAnnotationEdits([]string{"org.example.tested=true", "org.example.old-"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"org.example.tested": "true"}, edits.Set)
	assert.Equal(t, []string{"org.example.old"}, edits.Remove)

	_, err = ParseAnnotationEdits([]string{"org.example.count:=1"})
	assert.ErrorContains(t, err, "annotation values are strings")
}

func TestApplyEdits(t *testing.T) {
	existing := map[string]string{"org.example.old": "x", "org.example.keep": "y"}
	merged := ApplyEdits(existing, KeyValueEdits[string]{
		Set:    map[string]string{"org.example.keep": "z", "org.example.new": "w"},
		Remove: []string{"org.example.old"},
	})
	assert.Equal(t, map[string]string{"org.example.keep": "z", "org.example.new": "w"}, merged)
	assert.Equal(t, map[string]string{"org.example.old": "x", "org.example.keep": "y"}, existing, "existing values must not be mutated in place")

	assert.Nil(t, ApplyEdits(existing, KeyValueEdits[string]{Remove: []string{"org.example.old", "org.example.keep"}}),
		"an emptied map must be nil so it is omitted from the index")
}
