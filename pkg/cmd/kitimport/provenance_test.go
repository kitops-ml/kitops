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
	"encoding/json"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPredicateShape(t *testing.T) {
	prov := &ProvenanceData{
		KitfileConfigDigest: digest.NewDigestFromHex("sha256", "deadbeef"),
		ManifestDigest:      digest.NewDigestFromHex("sha256", "cafebabe"),
		SourceURI:           "hf://org/repo",
		SourceCommitSHA:     "abcdef0123456789",
		InvocationID:        "00000000-0000-4000-8000-000000000000",
		StartedOn:           time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		FinishedOn:          time.Date(2026, 5, 1, 12, 0, 1, 0, time.UTC),
		KitVersion:          "test",
		GoVersion:           "go1.99",
	}

	pred := BuildPredicate(prov)
	b, err := json.Marshal(pred)
	require.NoError(t, err)

	// Round-trip via map to assert the wire shape is exactly the predicate body
	// (no _type/subject/predicateType wrapper).
	var raw map[string]any
	require.NoError(t, json.Unmarshal(b, &raw))
	require.Contains(t, raw, "buildDefinition")
	require.Contains(t, raw, "runDetails")
	assert.NotContains(t, raw, "_type")
	assert.NotContains(t, raw, "subject")
	assert.NotContains(t, raw, "predicateType")

	bd := raw["buildDefinition"].(map[string]any)
	assert.Equal(t, BuildType, bd["buildType"])

	ext := bd["externalParameters"].(map[string]any)
	src := ext["source"].(map[string]any)
	assert.Equal(t, "hf://org/repo", src["uri"])
	srcDig := src["digest"].(map[string]any)
	assert.Equal(t, "abcdef0123456789", srcDig["gitCommit"])
	_, hasSha256 := srcDig["sha256"]
	assert.False(t, hasSha256, "source digest must use gitCommit, not sha256, for git commit IDs")

	kf := ext["kitfile"].(map[string]any)
	kfDig := kf["digest"].(map[string]any)
	assert.Equal(t, "deadbeef", kfDig["sha256"])

	// resolvedDependencies is intentionally empty: the source commit digest
	// pins the input tree and the manifest digest pins the output bytes.
	deps := bd["resolvedDependencies"].([]any)
	assert.Empty(t, deps)

	rd := raw["runDetails"].(map[string]any)
	builder := rd["builder"].(map[string]any)
	assert.Equal(t, BuilderID, builder["id"])
	version := builder["version"].(map[string]any)
	assert.Equal(t, "test", version["kit"])
	assert.Equal(t, "go1.99", version["go"])

	meta := rd["metadata"].(map[string]any)
	assert.Equal(t, "00000000-0000-4000-8000-000000000000", meta["invocationId"])
	assert.Equal(t, "2026-05-01T12:00:00Z", meta["startedOn"])
	assert.Equal(t, "2026-05-01T12:00:01Z", meta["finishedOn"])
}

func TestProvenanceValidate(t *testing.T) {
	good := func() *ProvenanceData {
		return &ProvenanceData{
			SourceURI:           "hf://org/repo",
			SourceCommitSHA:     "abc123",
			KitfileConfigDigest: digest.NewDigestFromHex("sha256", "deadbeef"),
			ManifestDigest:      digest.NewDigestFromHex("sha256", "cafebabe"),
		}
	}

	require.NoError(t, good().validate())

	tests := []struct {
		name    string
		mutate  func(p *ProvenanceData)
		errLike string
	}{
		{"empty source URI", func(p *ProvenanceData) { p.SourceURI = "" }, "source URI"},
		{"empty source commit", func(p *ProvenanceData) { p.SourceCommitSHA = "" }, "source commit SHA"},
		{"empty kitfile digest", func(p *ProvenanceData) { p.KitfileConfigDigest = "" }, "Kitfile config digest"},
		{"empty manifest digest", func(p *ProvenanceData) { p.ManifestDigest = "" }, "manifest digest"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := good()
			tt.mutate(p)
			err := p.validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errLike)
		})
	}
}

func TestNewInvocationIDFormat(t *testing.T) {
	id := newInvocationID()
	assert.Len(t, id, 36)
	// Versions: 8-4-4-4-12 hex segments separated by dashes; UUIDv4 has '4' at idx 14, variant 8/9/a/b at idx 19.
	assert.Equal(t, byte('4'), id[14])
	assert.Contains(t, "89ab", string(id[19]))
}
