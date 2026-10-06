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

package local

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry"
)

var testRef = &registry.Reference{Registry: "localhost", Repository: "test/repo"}

func testRepo(t *testing.T) LocalRepo {
	t.Helper()
	repo, err := NewLocalRepo(t.TempDir(), testRef)
	require.NoError(t, err)
	return repo
}

// testRepos returns ModelKit and index storage for the same repository, sharing blobs.
func testRepos(t *testing.T) (modelKits, indexes LocalRepo) {
	t.Helper()
	storagePath := t.TempDir()
	modelKits, err := NewLocalRepo(storagePath, testRef)
	require.NoError(t, err)
	indexes, err = NewLocalIndexRepo(storagePath, testRef)
	require.NoError(t, err)
	return modelKits, indexes
}

func pushBlob(t *testing.T, repo LocalRepo, mediaType string, content []byte) ocispec.Descriptor {
	t.Helper()
	desc := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest.FromBytes(content),
		Size:      int64(len(content)),
	}
	exists, err := repo.Exists(context.Background(), desc)
	require.NoError(t, err)
	if !exists {
		require.NoError(t, repo.Push(context.Background(), desc, bytes.NewReader(content)))
	}
	return desc
}

// pushModelKit stores a minimal ModelKit: a config blob, a layer blob, and the manifest
// referencing both.
func pushModelKit(t *testing.T, repo LocalRepo, tag string, layerContent string) ocispec.Descriptor {
	t.Helper()
	ctx := context.Background()

	configDesc := pushBlob(t, repo, mediatype.KitConfigMediaType.String(), []byte(`{"manifestVersion":"1.0.0"}`))
	layerDesc := pushBlob(t, repo, "application/vnd.kitops.modelkit.model.v1.tar", []byte(layerContent))

	manifest := ocispec.Manifest{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: mediatype.ArtifactTypeKitManifest,
		Config:       configDesc,
		Layers:       []ocispec.Descriptor{layerDesc},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)

	manifestDesc := pushBlob(t, repo, ocispec.MediaTypeImageManifest, manifestBytes)
	require.NoError(t, repo.Tag(ctx, manifestDesc, tag))
	return manifestDesc
}

func pushIndex(t *testing.T, repo LocalRepo, tag string, entries ...ocispec.Descriptor) ocispec.Descriptor {
	t.Helper()
	ctx := context.Background()

	index := ocispec.Index{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    ocispec.MediaTypeImageIndex,
		ArtifactType: mediatype.ArtifactTypeKitIndex,
		Manifests:    entries,
	}
	indexBytes, err := json.Marshal(index)
	require.NoError(t, err)

	indexDesc := pushBlob(t, repo, ocispec.MediaTypeImageIndex, indexBytes)
	require.NoError(t, repo.Tag(ctx, indexDesc, tag))
	return indexDesc
}

func TestIndexIsResolvableAndFetchable(t *testing.T) {
	ctx := context.Background()
	modelKits, indexes := testRepos(t)

	modelDesc := pushModelKit(t, modelKits, "q4_0", "weights-q4")
	indexDesc := pushIndex(t, indexes, "all", modelDesc)

	resolved, err := indexes.Resolve(ctx, "all")
	require.NoError(t, err)
	assert.Equal(t, indexDesc.Digest, resolved.Digest)
	assert.Equal(t, ocispec.MediaTypeImageIndex, resolved.MediaType)

	exists, err := indexes.Exists(ctx, indexDesc)
	require.NoError(t, err)
	assert.True(t, exists)

	assert.Equal(t, []string{"all"}, indexes.GetTags(indexDesc))
}

// TestDeleteIndexPreservesReferencedModelKits covers the oras AutoGC cascade: the store
// treats a manifest tagged only by its own digest as untagged, so deleting an index with
// garbage collection enabled would take every ModelKit it references with it.
func TestDeleteIndexPreservesReferencedModelKits(t *testing.T) {
	ctx := context.Background()
	modelKits, indexes := testRepos(t)

	modelA := pushModelKit(t, modelKits, "q4_0", "weights-q4")
	modelB := pushModelKit(t, modelKits, "q8_0", "weights-q8")
	indexDesc := pushIndex(t, indexes, "all", modelA, modelB)

	require.NoError(t, indexes.Delete(ctx, indexDesc))

	for _, model := range []ocispec.Descriptor{modelA, modelB} {
		exists, err := modelKits.Exists(ctx, model)
		require.NoError(t, err)
		assert.True(t, exists, "ModelKit %s must survive deleting the index", model.Digest)

		manifest, err := getManifest(t, modelKits, model)
		require.NoError(t, err, "ModelKit %s manifest must still be readable", model.Digest)

		configExists, err := modelKits.Exists(ctx, manifest.Config)
		require.NoError(t, err)
		assert.True(t, configExists, "config blob for %s must survive", model.Digest)

		layerExists, err := modelKits.Exists(ctx, manifest.Layers[0])
		require.NoError(t, err)
		assert.True(t, layerExists, "layer blob for %s must survive", model.Digest)
	}

	indexExists, err := indexes.Exists(ctx, indexDesc)
	require.NoError(t, err)
	assert.False(t, indexExists, "the index itself must be gone")
}

func getManifest(t *testing.T, repo LocalRepo, desc ocispec.Descriptor) (*ocispec.Manifest, error) {
	t.Helper()
	reader, err := repo.Fetch(context.Background(), desc)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	manifest := &ocispec.Manifest{}
	if err := json.NewDecoder(reader).Decode(manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func TestDeletingModelKitStillCollectsItsBlobs(t *testing.T) {
	ctx := context.Background()
	repo := testRepo(t)

	modelDesc := pushModelKit(t, repo, "q4_0", "weights-q4")
	manifest, err := getManifest(t, repo, modelDesc)
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, modelDesc))

	layerExists, err := repo.Exists(ctx, manifest.Layers[0])
	require.NoError(t, err)
	assert.False(t, layerExists, "deleting a ModelKit must still collect its layers")
}

func TestIndexAndModelKitTagsAreSeparate(t *testing.T) {
	ctx := context.Background()
	modelKits, indexes := testRepos(t)

	modelDesc := pushModelKit(t, modelKits, "all", "weights-q4")
	indexDesc := pushIndex(t, indexes, "all", modelDesc)

	resolvedModel, err := modelKits.Resolve(ctx, "all")
	require.NoError(t, err)
	assert.Equal(t, modelDesc.Digest, resolvedModel.Digest)

	resolvedIndex, err := indexes.Resolve(ctx, "all")
	require.NoError(t, err)
	assert.Equal(t, indexDesc.Digest, resolvedIndex.Digest)

	assert.Len(t, modelKits.GetAllModels(), 1)
	assert.Len(t, indexes.GetAllModels(), 1)
}

// TestDeleteIndexKeepsBlobSharedWithOtherRepository covers identical indexes in two
// repositories, such as empty ones, which share a blob in the common store.
func TestDeleteIndexKeepsBlobSharedWithOtherRepository(t *testing.T) {
	ctx := context.Background()
	storagePath := t.TempDir()
	repoA, err := NewLocalIndexRepo(storagePath, &registry.Reference{Registry: "localhost", Repository: "test/a"})
	require.NoError(t, err)
	repoB, err := NewLocalIndexRepo(storagePath, &registry.Reference{Registry: "localhost", Repository: "test/b"})
	require.NoError(t, err)

	indexDesc := pushIndex(t, repoA, "all")
	pushIndex(t, repoB, "all")

	require.NoError(t, repoA.Delete(ctx, indexDesc))

	_, err = getManifest(t, repoB, indexDesc)
	assert.NoError(t, err, "the other repository's index must still be readable")
}
