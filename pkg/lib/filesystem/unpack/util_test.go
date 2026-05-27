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

package unpack

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mtModel     = "application/vnd.kitops.modelkit.model.v1.tar+gzip"
	mtModelPart = "application/vnd.kitops.modelkit.modelpart.v1.tar+gzip"
	mtDataset   = "application/vnd.kitops.modelkit.dataset.v1.tar+gzip"
	mtCode      = "application/vnd.kitops.modelkit.code.v1.tar+gzip"
	mtDocs      = "application/vnd.kitops.modelkit.docs.v1.tar+gzip"
)

// fakeDigest builds a syntactically valid sha256 digest from a short id so test
// fixtures stay readable.
func fakeDigest(id string) string {
	d := digest.FromString(id)
	return d.String()
}

// layerDesc constructs an OCI descriptor for use in a synthetic manifest.
func layerDesc(id, mt string) ocispec.Descriptor {
	return ocispec.Descriptor{
		Digest:    digest.Digest(fakeDigest(id)),
		MediaType: mt,
		Size:      1,
	}
}

func layerInfo(id string) *artifact.LayerInfo {
	return &artifact.LayerInfo{Digest: fakeDigest(id)}
}

func TestGenerateUnpackPlan_EmptyKitfile(t *testing.T) {
	manifest := &ocispec.Manifest{}
	kitfile := &artifact.KitFile{ManifestVersion: "1.0.0"}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	assert.Empty(t, steps, "empty kitfile should yield no steps")
}

func TestGenerateUnpackPlan_ModelOnly(t *testing.T) {
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{layerDesc("model", mtModel)},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
		},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, fakeDigest("model"), steps[0].desc.Digest.String())
	assert.Equal(t, mediatype.ModelBaseType, steps[0].mediatype.Base())
	assert.Equal(t, "Unpacking model llama to model.bin", steps[0].userMessage)
}

func TestGenerateUnpackPlan_ModelWithParts(t *testing.T) {
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("model", mtModel),
			layerDesc("part-a", mtModelPart),
			layerDesc("part-b", mtModelPart),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
			Parts: []artifact.ModelPart{
				{Name: "tokenizer", Path: "tokenizer.json", LayerInfo: layerInfo("part-a")},
				{Name: "weights", Path: "weights.bin", LayerInfo: layerInfo("part-b")},
			},
		},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 3)
	assert.Equal(t, "Unpacking model llama to model.bin", steps[0].userMessage)
	assert.Equal(t, "Unpacking modelpart tokenizer to tokenizer.json", steps[1].userMessage)
	assert.Equal(t, "Unpacking modelpart weights to weights.bin", steps[2].userMessage)
}

func TestGenerateUnpackPlan_AllLayerTypesOrder(t *testing.T) {
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("model", mtModel),
			layerDesc("part", mtModelPart),
			layerDesc("ds1", mtDataset),
			layerDesc("ds2", mtDataset),
			layerDesc("code", mtCode),
			layerDesc("docs", mtDocs),
			layerDesc("prompt", mtCode),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "m",
			Path:      "m",
			LayerInfo: layerInfo("model"),
			Parts: []artifact.ModelPart{
				{Name: "p", Path: "p", LayerInfo: layerInfo("part")},
			},
		},
		DataSets: []artifact.DataSet{
			{Name: "train", Path: "data/train", LayerInfo: layerInfo("ds1")},
			{Name: "val", Path: "data/val", LayerInfo: layerInfo("ds2")},
		},
		Code:    []artifact.Code{{Path: "src", LayerInfo: layerInfo("code")}},
		Docs:    []artifact.Docs{{Path: "docs", LayerInfo: layerInfo("docs")}},
		Prompts: []artifact.Prompt{{Name: "sys", Path: "prompts/sys.txt", LayerInfo: layerInfo("prompt")}},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 7)

	// Steps must follow the order: model, modelparts, datasets, code, docs, prompts.
	wantMessages := []string{
		"Unpacking model m to m",
		"Unpacking modelpart p to p",
		"Unpacking dataset train to data/train",
		"Unpacking dataset val to data/val",
		"Unpacking code to src",
		"Unpacking docs to docs",
		"Unpacking prompt sys to prompts/sys.txt",
	}
	for i, want := range wantMessages {
		assert.Equal(t, want, steps[i].userMessage, "step %d", i)
	}
}

func TestGenerateUnpackPlan_MessageFormatting(t *testing.T) {
	// Code and Docs have no Name field so the message should drop the "name" segment.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("code", mtCode),
			layerDesc("docs", mtDocs),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Code:            []artifact.Code{{Path: "src/", LayerInfo: layerInfo("code")}},
		Docs:            []artifact.Docs{{Path: "README.md", LayerInfo: layerInfo("docs")}},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 2)
	assert.Equal(t, "Unpacking code to src/", steps[0].userMessage)
	assert.Equal(t, "Unpacking docs to README.md", steps[1].userMessage)
}

func TestGenerateUnpackPlan_FilterExcludesNonMatchingLayers(t *testing.T) {
	// A datasets-only filter should drop the model, modelpart, code, docs,
	// and prompt steps, leaving just the dataset.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("model", mtModel),
			layerDesc("part", mtModelPart),
			layerDesc("ds", mtDataset),
			layerDesc("code", mtCode),
			layerDesc("docs", mtDocs),
			layerDesc("prompt", mtCode),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
			Parts: []artifact.ModelPart{
				{Name: "tok", Path: "tok", LayerInfo: layerInfo("part")},
			},
		},
		DataSets: []artifact.DataSet{
			{Name: "train", Path: "train", LayerInfo: layerInfo("ds")},
		},
		Code:    []artifact.Code{{Path: "src", LayerInfo: layerInfo("code")}},
		Docs:    []artifact.Docs{{Path: "README.md", LayerInfo: layerInfo("docs")}},
		Prompts: []artifact.Prompt{{Name: "sys", Path: "prompts/sys.txt", LayerInfo: layerInfo("prompt")}},
	}

	filters := []kfutils.FilterConf{{BaseTypes: []kfutils.BaseType{kfutils.BaseTypeDatasets}}}

	steps, err := generateUnpackPlan(manifest, kitfile, filters)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "Unpacking dataset train to train", steps[0].userMessage)
}

func TestGenerateUnpackPlan_FilterMultipleBaseTypes(t *testing.T) {
	// A filter targeting both model and code base types should keep the model,
	// its parts, and the code layer — but drop datasets, docs, and prompts.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("model", mtModel),
			layerDesc("part", mtModelPart),
			layerDesc("ds", mtDataset),
			layerDesc("code", mtCode),
			layerDesc("docs", mtDocs),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
			Parts: []artifact.ModelPart{
				{Name: "tok", Path: "tok", LayerInfo: layerInfo("part")},
			},
		},
		DataSets: []artifact.DataSet{{Name: "train", Path: "train", LayerInfo: layerInfo("ds")}},
		Code:     []artifact.Code{{Path: "src", LayerInfo: layerInfo("code")}},
		Docs:     []artifact.Docs{{Path: "README.md", LayerInfo: layerInfo("docs")}},
	}

	filters := []kfutils.FilterConf{{
		BaseTypes: []kfutils.BaseType{kfutils.BaseTypeModel, kfutils.BaseTypeCode},
	}}

	steps, err := generateUnpackPlan(manifest, kitfile, filters)
	require.NoError(t, err)
	require.Len(t, steps, 3)
	assert.Equal(t, "Unpacking model llama to model.bin", steps[0].userMessage)
	assert.Equal(t, "Unpacking modelpart tok to tok", steps[1].userMessage)
	assert.Equal(t, "Unpacking code to src", steps[2].userMessage)
}

func TestGenerateUnpackPlan_FilterByFieldName(t *testing.T) {
	// A filter with Filters set matches against name/path of model and modelpart;
	// only the matching part should remain in the plan.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			layerDesc("model", mtModel),
			layerDesc("part-a", mtModelPart),
			layerDesc("part-b", mtModelPart),
		},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
			Parts: []artifact.ModelPart{
				{Name: "tokenizer", Path: "tok", LayerInfo: layerInfo("part-a")},
				{Name: "weights", Path: "weights.bin", LayerInfo: layerInfo("part-b")},
			},
		},
	}

	filters := []kfutils.FilterConf{{
		BaseTypes: []kfutils.BaseType{kfutils.BaseTypeModel},
		Filters:   []string{"tokenizer"},
	}}

	steps, err := generateUnpackPlan(manifest, kitfile, filters)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "Unpacking modelpart tokenizer to tok", steps[0].userMessage)
}

func TestGenerateUnpackPlan_FilterMatchingModel(t *testing.T) {
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{layerDesc("model", mtModel)},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
		},
	}
	filters := []kfutils.FilterConf{{BaseTypes: []kfutils.BaseType{kfutils.BaseTypeModel}}}

	steps, err := generateUnpackPlan(manifest, kitfile, filters)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "Unpacking model llama to model.bin", steps[0].userMessage)
}

func TestGenerateUnpackPlan_SkipsRemoteModelReference(t *testing.T) {
	// A model whose Path is a ModelKit reference has no layer in the manifest
	// and must be skipped — it shouldn't trigger a "digest not found" error.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{layerDesc("code", mtCode)},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model:           &artifact.Model{Name: "llama", Path: "registry.io/myorg/model:v1"},
		Code:            []artifact.Code{{Path: "src", LayerInfo: layerInfo("code")}},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "Unpacking code to src", steps[0].userMessage)
}

func TestGenerateUnpackPlan_SkipsRemoteDataset(t *testing.T) {
	// Datasets with a RemotePath (S3 or ModelKit ref) are resolved separately
	// and must be skipped here.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{layerDesc("ds", mtDataset)},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		DataSets: []artifact.DataSet{
			{Name: "s3", Path: "s3data", RemotePath: "s3://bucket/key", RemoteHash: "sha256:abc"},
			{Name: "ref", Path: "refdata", RemotePath: "registry.io/myorg/dataset:v1"},
			{Name: "local", Path: "data", LayerInfo: layerInfo("ds")},
		},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "Unpacking dataset local to data", steps[0].userMessage)
}

func TestGenerateUnpackPlan_DigestNotInManifest(t *testing.T) {
	// Manifest is missing the layer the kitfile references.
	manifest := &ocispec.Manifest{Layers: []ocispec.Descriptor{}}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
		},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	assert.Nil(t, steps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in manifest")
}

func TestGenerateUnpackPlan_InvalidMediaType(t *testing.T) {
	// Manifest layer matches the kitfile digest but its media type is unparseable.
	manifest := &ocispec.Manifest{
		Layers: []ocispec.Descriptor{layerDesc("model", "application/octet-stream")},
	}
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Model: &artifact.Model{
			Name:      "llama",
			Path:      "model.bin",
			LayerInfo: layerInfo("model"),
		},
	}

	steps, err := generateUnpackPlan(manifest, kitfile, nil)
	assert.Nil(t, steps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized media type")
}
