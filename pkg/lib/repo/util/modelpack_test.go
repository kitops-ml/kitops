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

package util

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestGenerateKitfileForModelPack(t *testing.T) {
	manifest := &ocispec.Manifest{
		ArtifactType: mediatype.ArtifactTypeModelManifest,
		Config: ocispec.Descriptor{
			MediaType: mediatype.ModelPackConfigMediaType.String(),
		},
		Layers: []ocispec.Descriptor{
			{
				MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.ModelBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
				Annotations: map[string]string{
					modelspecv1.AnnotationFilepath: "models/main.gguf",
				},
			},
			{
				MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.CodeBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
				Annotations: map[string]string{
					modelspecv1.AnnotationFilepath: "src/app.py",
				},
			},
			{
				MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.DatasetBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
				Annotations: map[string]string{
					modelspecv1.AnnotationFilepath: "data/train.csv",
				},
			},
			{
				MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.DocsBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
				Annotations: map[string]string{
					modelspecv1.AnnotationFilepath: "docs/readme.md",
				},
			},
		},
	}

	kf, err := GenerateKitfileForModelPack(manifest)
	require.NoError(t, err)
	require.NotNil(t, kf.Model)
	require.Equal(t, "models/main.gguf", kf.Model.Path)
	require.Len(t, kf.Code, 1)
	require.Equal(t, "src/app.py", kf.Code[0].Path)
	require.Len(t, kf.DataSets, 1)
	require.Equal(t, "data/train.csv", kf.DataSets[0].Path)
	require.Len(t, kf.Docs, 1)
	require.Equal(t, "docs/readme.md", kf.Docs[0].Path)
}

func TestGenerateKitfileForModelPackMissingPathAnnotation(t *testing.T) {
	manifest := &ocispec.Manifest{
		ArtifactType: mediatype.ArtifactTypeModelManifest,
		Config: ocispec.Descriptor{
			MediaType: mediatype.ModelPackConfigMediaType.String(),
		},
		Layers: []ocispec.Descriptor{
			{
				MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.ModelBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
			},
		},
	}

	_, err := GenerateKitfileForModelPack(manifest)
	require.Error(t, err)
	require.Contains(t, err.Error(), modelspecv1.AnnotationFilepath)
}
