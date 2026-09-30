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

	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry"
)

var testIndexRef = &registry.Reference{Registry: "registry.example.com", Repository: "my-org/my-model", Reference: "all"}

func labelledEntry(dgst string, labels ModelMetadata) ModelKitIndexDescriptor {
	return ModelKitIndexDescriptor{
		Descriptor: ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: digest.Digest(dgst), Size: 649},
		ModelMeta:  labels,
	}
}

func TestParseLabelSelector(t *testing.T) {
	selector, err := ParseLabelSelector([]string{"quantization=q4_0", "vram=6GB"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"quantization": "q4_0", "vram": "6GB"}, selector)

	// The edit syntax from 'kit index add' is not a selector.
	_, err = ParseLabelSelector([]string{"quantization-"})
	assert.ErrorContains(t, err, "expected key=value")

	_, err = ParseLabelSelector([]string{"=q4_0"})
	assert.ErrorContains(t, err, "expected key=value")
}

func TestSelectEntryMatchesOne(t *testing.T) {
	idx := CreateIndex([]ModelKitIndexDescriptor{
		labelledEntry(digestA, ModelMetadata{"quantization": json.RawMessage(`"q4_0"`), "vram": json.RawMessage(`"6GB"`)}),
		labelledEntry(digestB, ModelMetadata{"quantization": json.RawMessage(`"q8_0"`)}),
	})

	entry, err := SelectEntry(idx, map[string]string{"quantization": "q8_0"}, testIndexRef)
	require.NoError(t, err)
	assert.Equal(t, digest.Digest(digestB), entry.Digest)

	// Every label in the selector must be present on the entry.
	entry, err = SelectEntry(idx, map[string]string{"quantization": "q4_0", "vram": "6GB"}, testIndexRef)
	require.NoError(t, err)
	assert.Equal(t, digest.Digest(digestA), entry.Digest)

	_, err = SelectEntry(idx, map[string]string{"quantization": "q4_0", "vram": "24GB"}, testIndexRef)
	assert.ErrorContains(t, err, "no ModelKit in the index has quantization=q4_0, vram=24GB")
}

func TestSelectEntryRequiresUniqueMatch(t *testing.T) {
	idx := CreateIndex([]ModelKitIndexDescriptor{
		labelledEntry(digestA, ModelMetadata{"quantization": json.RawMessage(`"q4_0"`)}),
		labelledEntry(digestB, ModelMetadata{"quantization": json.RawMessage(`"q8_0"`)}),
	})

	_, err := SelectEntry(idx, nil, testIndexRef)
	assert.ErrorContains(t, err, "2 ModelKits in the index match")
	assert.ErrorContains(t, err, "quantization=q4_0")
	assert.ErrorContains(t, err, "quantization=q8_0")

	// A single-entry index needs no selector.
	single := CreateIndex([]ModelKitIndexDescriptor{labelledEntry(digestA, nil)})
	entry, err := SelectEntry(single, nil, testIndexRef)
	require.NoError(t, err)
	assert.Equal(t, digest.Digest(digestA), entry.Digest)

	_, err = SelectEntry(CreateIndex(nil), nil, testIndexRef)
	assert.ErrorContains(t, err, "index contains no ModelKits")
}

func TestSelectEntryReportsUnlabelledEntries(t *testing.T) {
	idx := CreateIndex([]ModelKitIndexDescriptor{
		labelledEntry(digestA, nil),
		labelledEntry(digestB, nil),
	})

	_, err := SelectEntry(idx, map[string]string{"quantization": "q4_0"}, testIndexRef)
	assert.ErrorContains(t, err, "no labels")
}

func TestDescribeEntriesNamesReferences(t *testing.T) {
	withTag := labelledEntry(digestA, ModelMetadata{"quantization": json.RawMessage(`"q4_0"`)})
	withTag.Annotations = map[string]string{constants.OriginalTagAnnotation: "q4_0"}

	described := DescribeEntries([]ModelKitIndexDescriptor{withTag, labelledEntry(digestB, nil)}, testIndexRef)
	assert.Contains(t, described, "registry.example.com/my-org/my-model:q4_0  (quantization=q4_0)",
		"the original tag is shown in the index's repository")
	assert.Contains(t, described, digestB+"  (no labels)", "entries without a tag fall back to their digest")
}
