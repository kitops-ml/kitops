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

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEntry(dgst string) ModelKitIndexDescriptor {
	return ModelKitIndexDescriptor{
		Descriptor: ocispec.Descriptor{
			MediaType: ocispec.MediaTypeImageManifest,
			Digest:    digest.Digest(dgst),
			Size:      649,
		},
	}
}

const (
	digestA = "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
	digestB = "sha256:da799e0122c7ed9ee7a934ef25c5e5b7a9bf896252e7215d9dfa9c8894b86bfc"
)

func TestCreateIndexHasEmptyManifests(t *testing.T) {
	_, indexBytes, err := CreateIndex(nil).Marshal()
	require.NoError(t, err)
	assert.Contains(t, string(indexBytes), `"manifests":[]`)
	assert.NotContains(t, string(indexBytes), `"manifests":null`)
}

func TestCreateIndexMediaTypes(t *testing.T) {
	idx := CreateIndex(nil)
	assert.Equal(t, "application/vnd.kitops.modelkit.index.v1+json", idx.ArtifactType)
	desc, _, err := idx.Marshal()
	require.NoError(t, err)
	assert.Equal(t, ocispec.MediaTypeImageIndex, desc.MediaType)
}

func TestAddEntryDedupesByDigest(t *testing.T) {
	idx := CreateIndex(nil)
	idx.AddEntry(testEntry(digestA))
	idx.AddEntry(testEntry(digestB))
	idx.AddEntry(testEntry(digestA))
	assert.Len(t, idx.Manifests, 2)
}

func TestAddEntryReplacesMetadata(t *testing.T) {
	idx := CreateIndex(nil)
	idx.AddEntry(testEntry(digestA))

	updated := testEntry(digestA)
	updated.ModelMeta = ModelMetadata{"quantization": json.RawMessage(`"q4_0"`)}
	idx.AddEntry(updated)

	require.Len(t, idx.Manifests, 1)
	assert.Equal(t, updated.ModelMeta, idx.Manifests[0].ModelMeta)
}

func TestRemoveEntry(t *testing.T) {
	idx := CreateIndex(nil)
	idx.AddEntry(testEntry(digestA))
	idx.AddEntry(testEntry(digestB))

	assert.True(t, idx.RemoveEntry(digest.Digest(digestA)))
	assert.False(t, idx.RemoveEntry(digest.Digest(digestA)))
	require.Len(t, idx.Manifests, 1)
	assert.Equal(t, digest.Digest(digestB), idx.Manifests[0].Digest)
}

func TestRoundTripPreservesMetadata(t *testing.T) {
	entry := testEntry(digestA)
	entry.ModelMeta = ModelMetadata{
		"quantization":      json.RawMessage(`"q4_0"`),
		"totalParameters":   json.RawMessage(`7000000000`),
		"vramRequirement":   json.RawMessage(`{"bytes":6442450944}`),
		"targetAccelerator": json.RawMessage(`["cuda","metal"]`),
	}
	idx := CreateIndex([]ModelKitIndexDescriptor{entry})

	desc, indexBytes, err := idx.Marshal()
	require.NoError(t, err)

	parsed, err := ParseIndex(indexBytes)
	require.NoError(t, err)
	require.Len(t, parsed.Manifests, 1)
	assert.Equal(t, entry.ModelMeta, parsed.Manifests[0].ModelMeta)

	reDesc, reBytes, err := parsed.Marshal()
	require.NoError(t, err)
	assert.Equal(t, desc.Digest, reDesc.Digest, "re-marshalling must be digest-stable")
	assert.Contains(t, string(reBytes), "7000000000",
		"numbers must survive verbatim; decoding into any would rewrite this as 7e+09")
}

func TestLabelPairsFormatsValues(t *testing.T) {
	pairs := LabelPairs(ModelMetadata{
		"quantization":      json.RawMessage(`"q4_0"`),
		"activeParameters":  json.RawMessage(`7000000000`),
		"targetAccelerator": json.RawMessage(`["cuda","metal"]`),
	})

	assert.Equal(t, []string{
		"activeParameters=7000000000",
		`quantization=q4_0`,
		`targetAccelerator=["cuda","metal"]`,
	}, pairs, "keys are sorted and string values are unquoted")
}

func TestParseIndexRejectsOtherArtifacts(t *testing.T) {
	kitManifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json",` +
		`"artifactType":"application/vnd.kitops.modelkit.manifest.v1+json","layers":[]}`)
	_, err := ParseIndex(kitManifest)
	assert.ErrorIs(t, err, ErrNotAKitIndex)

	plainIndex := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}`)
	_, err = ParseIndex(plainIndex)
	assert.ErrorIs(t, err, ErrNotAKitIndex)
}

func TestParseIndexRejectsMalformedJSON(t *testing.T) {
	_, err := ParseIndex([]byte(`{"schemaVersion":`))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotAKitIndex)
}
