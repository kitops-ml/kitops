// Copyright 2024 The KitOps Authors.
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

package list

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestModelInfoFillNormal(t *testing.T) {
	info := modelInfo{}
	manifest := genTestManifest(100, 200, 300)
	kitfile := &artifact.KitFile{
		Package: artifact.Package{
			Name:    "testmodelkit",
			Authors: []string{"testauthor1", "testauthor2"},
		},
	}
	info.fill(manifest, kitfile)
	// Not testing size formatting here
	assert.Equal(t, info.Size, "600 B")
	// Should use first author in list
	assert.Equal(t, info.Author, "testauthor1")
	assert.Equal(t, info.ModelName, "testmodelkit")
}

func TestModelInfoFillEmptyKitfile(t *testing.T) {
	info := modelInfo{}
	manifest := genTestManifest(100, 200, 300)
	kitfile := &artifact.KitFile{}
	info.fill(manifest, kitfile)
	// Not testing size formatting here
	assert.Equal(t, info.Size, "600 B")
	// Should use first author in list
	assert.Equal(t, info.Author, "<none>")
	assert.Equal(t, info.ModelName, "<none>")
}

func TestModelInfoFillNilKitfile(t *testing.T) {
	info := modelInfo{}
	manifest := genTestManifest(100, 200, 300)
	info.fill(manifest, nil)
	// Not testing size formatting here
	assert.Equal(t, info.Size, "600 B")
	// Should use first author in list
	assert.Equal(t, info.Author, "<none>")
	assert.Equal(t, info.ModelName, "<none>")
}

func genTestManifest(layersizes ...int64) *ocispec.Manifest {
	var layers []ocispec.Descriptor
	for _, s := range layersizes {
		layers = append(layers, ocispec.Descriptor{
			// Empty media type is sufficient for now as we don't care about mediatypes
			// when filling
			MediaType: ocispec.DescriptorEmptyJSON.MediaType,
			Digest:    ocispec.DescriptorEmptyJSON.Digest,
			Size:      s,
		})
	}
	return &ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    ocispec.DescriptorEmptyJSON,
		Layers:    layers,
	}
}
