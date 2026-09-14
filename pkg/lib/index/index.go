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
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var ErrNotAKitIndex = errors.New("artifact is not a ModelKit index")

type ModelKitIndex struct {
	specs.Versioned

	MediaType    string                    `json:"mediaType,omitempty"`
	ArtifactType string                    `json:"artifactType,omitempty"`
	Manifests    []ModelKitIndexDescriptor `json:"manifests"`
	Subject      *ocispec.Descriptor       `json:"subject,omitempty"`
	Annotations  map[string]string         `json:"annotations,omitempty"`
}

type ModelKitIndexDescriptor struct {
	ocispec.Descriptor
	ModelMeta ModelMetadata `json:"modelMetadata,omitempty"`
}

// ModelMetadata is an open set of metadata about a ModelKit. Keys and values are not
// interpreted by KitOps; values are stored and returned as they were written.
type ModelMetadata map[string]json.RawMessage

func CreateIndex(metas []ModelKitIndexDescriptor) *ModelKitIndex {
	return &ModelKitIndex{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    ocispec.MediaTypeImageIndex,
		ArtifactType: mediatype.ArtifactTypeKitIndex,
		Manifests:    metas,
	}
}

// AddEntry adds an entry to the index, replacing any existing entry with the same digest.
func (idx *ModelKitIndex) AddEntry(entry ModelKitIndexDescriptor) {
	for i, existing := range idx.Manifests {
		if existing.Digest == entry.Digest {
			idx.Manifests[i] = entry
			return
		}
	}
	idx.Manifests = append(idx.Manifests, entry)
}

// RemoveEntry removes the entry with the given digest, reporting whether it was present.
func (idx *ModelKitIndex) RemoveEntry(dgst digest.Digest) bool {
	for i, existing := range idx.Manifests {
		if existing.Digest == dgst {
			idx.Manifests = append(idx.Manifests[:i], idx.Manifests[i+1:]...)
			return true
		}
	}
	return false
}

func (idx *ModelKitIndex) GetEntry(dgst digest.Digest) (ModelKitIndexDescriptor, bool) {
	for _, existing := range idx.Manifests {
		if existing.Digest == dgst {
			return existing, true
		}
	}
	return ModelKitIndexDescriptor{}, false
}

// Marshal serializes the index and returns the descriptor for the resulting bytes.
func (idx *ModelKitIndex) Marshal() (ocispec.Descriptor, []byte, error) {
	if idx.Manifests == nil {
		idx.Manifests = []ModelKitIndexDescriptor{}
	}
	indexBytes, err := json.Marshal(idx)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, nil, fmt.Errorf("failed to marshal index: %w", err)
	}
	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageIndex,
		Digest:    digest.FromBytes(indexBytes),
		Size:      int64(len(indexBytes)),
	}
	return desc, indexBytes, nil
}

// ParseIndex deserializes a ModelKit index, returning ErrNotAKitIndex if the content
// does not describe one.
func ParseIndex(indexBytes []byte) (*ModelKitIndex, error) {
	idx := &ModelKitIndex{}
	if err := json.Unmarshal(indexBytes, idx); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}
	if !mediatype.IsKitIndex(idx.MediaType, idx.ArtifactType) {
		return nil, ErrNotAKitIndex
	}
	return idx, nil
}
