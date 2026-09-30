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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

const testDigest = "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"

func testContext(t *testing.T) context.Context {
	t.Helper()
	return context.WithValue(context.Background(), constants.ConfigKey{}, t.TempDir())
}

func TestAddCompleteRejectsCrossRepository(t *testing.T) {
	opts := &addOptions{}
	err := opts.complete(testContext(t), []string{"my-org/my-model:all", "other-org/other-model:q4_0"})
	assert.ErrorContains(t, err, "not in the same repository")
}

func pushRejection(status int, code string) error {
	return &errcode.ErrorResponse{
		Method:     http.MethodPut,
		URL:        &url.URL{Scheme: "https", Host: "registry.example.com", Path: "/v2/my-org/my-model/manifests/all"},
		StatusCode: status,
		Errors:     errcode.Errors{{Code: code, Message: "blob unknown to registry"}},
	}
}

func testDestRef() *registry.Reference {
	return &registry.Reference{Registry: "registry.example.com", Repository: "my-org/my-model", Reference: "all"}
}

// Registries agree on the MANIFEST_BLOB_UNKNOWN code for an index referencing manifests they
// do not have, but not on the status code: distribution answers 400 rather than 404.
func TestDescribePushErrorExplainsMissingModelKits(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusNotFound} {
		err := describePushError(pushRejection(status, errcode.ErrorCodeManifestBlobUnknown), testDestRef())
		assert.ErrorContains(t, err, "does not contain every ModelKit")
		assert.ErrorContains(t, err, "registry.example.com/my-org/my-model")
	}
}

func TestDescribePushErrorPassesThroughOtherFailures(t *testing.T) {
	nameUnknown := pushRejection(http.StatusNotFound, errcode.ErrorCodeNameUnknown)
	assert.Equal(t, nameUnknown, describePushError(nameUnknown, testDestRef()))

	unauthorized := pushRejection(http.StatusUnauthorized, errcode.ErrorCodeUnauthorized)
	assert.Equal(t, unauthorized, describePushError(unauthorized, testDestRef()))

	refused := errors.New("connection refused")
	assert.Equal(t, refused, describePushError(refused, testDestRef()))
}

func TestDeleteAll(t *testing.T) {
	ctx := context.Background()
	configHome := t.TempDir()
	ref := &registry.Reference{Registry: "localhost", Repository: "test/model", Reference: "all"}
	openRepo := func() local.LocalRepo {
		repo, err := local.NewLocalIndexRepo(constants.StoragePath(configHome), ref)
		require.NoError(t, err)
		return repo
	}

	tagged, err := writeIndex(ctx, openRepo(), libindex.CreateIndex(nil), ref, ocispec.Descriptor{})
	require.NoError(t, err)
	oldRef := *ref
	oldRef.Reference = "old"
	entry := libindex.ModelKitIndexDescriptor{
		Descriptor: ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: digest.Digest(testDigest), Size: 649},
	}
	untagged, err := writeIndex(ctx, openRepo(), libindex.CreateIndex([]libindex.ModelKitIndexDescriptor{entry}), &oldRef, ocispec.Descriptor{})
	require.NoError(t, err)
	require.NoError(t, openRepo().Untag(ctx, "old"))
	exists, err := openRepo().Exists(ctx, untagged)
	require.NoError(t, err)
	require.True(t, exists, "untagging should keep the index")

	require.NoError(t, runDeleteAll(ctx, &deleteOptions{configHome: configHome}))
	exists, err = openRepo().Exists(ctx, untagged)
	require.NoError(t, err)
	assert.False(t, exists, "untagged index should be deleted")
	exists, err = openRepo().Exists(ctx, tagged)
	require.NoError(t, err)
	assert.True(t, exists, "tagged index should be kept without --force")

	require.NoError(t, runDeleteAll(ctx, &deleteOptions{configHome: configHome, forceDelete: true}))
	exists, err = openRepo().Exists(ctx, tagged)
	require.NoError(t, err)
	assert.False(t, exists, "tagged index should be deleted with --force")
}

const testDigestB = "sha256:b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"

func testIndexEntry(dgst string, labels libindex.ModelMetadata) libindex.ModelKitIndexDescriptor {
	return libindex.ModelKitIndexDescriptor{
		Descriptor: ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: digest.Digest(dgst), Size: 649},
		ModelMeta:  labels,
	}
}

// pushTestModelKit stores a ModelKit manifest whose layers total layerSize bytes, returning
// its descriptor.
func pushTestModelKit(t *testing.T, store *memory.Store, layerSize int64) ocispec.Descriptor {
	t.Helper()
	manifest := ocispec.Manifest{
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: mediatype.ArtifactTypeKitManifest,
		Config:       ocispec.Descriptor{Digest: digest.Digest(testDigest), Size: 100},
		Layers: []ocispec.Descriptor{
			{Digest: digest.Digest(testDigestB), Size: layerSize / 2},
			{Digest: digest.Digest(testDigest), Size: layerSize / 2},
		},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	require.NoError(t, store.Push(context.Background(), desc, bytes.NewReader(manifestBytes)))
	return desc
}

func TestPrintIndexInfoTable(t *testing.T) {
	store := memory.New()
	modelDesc := pushTestModelKit(t, store, 4096)

	entry := libindex.ModelKitIndexDescriptor{
		Descriptor: modelDesc,
		ModelMeta: libindex.ModelMetadata{
			"quantization": json.RawMessage(`"q4_0"`),
			"vram":         json.RawMessage(`"6GB"`),
		},
	}
	entry.Annotations = map[string]string{
		constants.OriginalTagAnnotation: "q4_0",
		"org.example.tested":            "true",
	}
	idx := libindex.CreateIndex([]libindex.ModelKitIndexDescriptor{
		entry,
		testIndexEntry(testDigestB, nil),
	})
	ref := &registry.Reference{Registry: "registry.example.com", Repository: "my-org/my-model", Reference: "all"}

	var buf bytes.Buffer
	printIndexInfo(context.Background(), &buf, store, ocispec.Descriptor{Digest: digest.Digest(testDigest)}, idx, ref)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

	require.Len(t, lines, 6, "header, two rows for the labelled entry, one for the other")
	assert.Equal(t, "2 ModelKits:", lines[1])
	assert.Regexp(t, `^ORIGINAL REFERENCE\s+SIZE\s+LABELS\s+ANNOTATIONS\s+DIGEST$`, lines[2])
	assert.Regexp(t, `^registry\.example\.com/my-org/my-model:q4_0\s+4.0 KiB\s+quantization=q4_0\s+org\.example\.tested=true\s+`+modelDesc.Digest.String()+`$`, lines[3],
		"the original tag is shown in the index's repository; size is the ModelKit's layers, not the manifest blob")
	assert.Regexp(t, `^\s+vram=6GB$`, lines[4], "extra labels continue on their own row, with no trailing padding")
	assert.Regexp(t, `^<none>\s+<none>\s+<none>\s+<none>\s+`+testDigestB+`$`, lines[5],
		"a ModelKit that is not available has an unknown size")
	assert.NotContains(t, buf.String(), constants.OriginalTagAnnotation, "the original tag annotation has its own column")
}

func TestPrintIndexInfoJSON(t *testing.T) {
	store := memory.New()
	modelDesc := pushTestModelKit(t, store, 4096)

	entry := libindex.ModelKitIndexDescriptor{
		Descriptor: modelDesc,
		ModelMeta: libindex.ModelMetadata{
			"quantization": json.RawMessage(`"q4_0"`),
			"vramBytes":    json.RawMessage(`6442450944`),
		},
	}
	entry.Annotations = map[string]string{
		constants.OriginalTagAnnotation: "q4_0",
		"org.example.tested":            "true",
	}
	idx := libindex.CreateIndex([]libindex.ModelKitIndexDescriptor{
		entry,
		testIndexEntry(testDigestB, nil),
	})
	ref := &registry.Reference{Registry: "registry.example.com", Repository: "my-org/my-model", Reference: "all"}

	var buf bytes.Buffer
	require.NoError(t, printIndexInfoJSON(context.Background(), &buf, store, ocispec.Descriptor{Digest: digest.Digest(testDigest)}, idx, ref))

	assert.JSONEq(t, `{
		"reference": "registry.example.com/my-org/my-model:all",
		"digest": "`+testDigest+`",
		"modelKits": [
			{
				"reference": "registry.example.com/my-org/my-model:q4_0",
				"digest": "`+modelDesc.Digest.String()+`",
				"size": 4096,
				"labels": {"quantization": "q4_0", "vramBytes": 6442450944},
				"annotations": {"org.example.tested": "true"}
			},
			{
				"digest": "`+testDigestB+`",
				"size": null,
				"labels": {},
				"annotations": {}
			}
		]
	}`, buf.String())
}

func TestPrintIndexListJSON(t *testing.T) {
	count := 3
	listings := []indexListing{
		{Repo: "registry.example.com/my-org/my-model", Digest: testDigest, Tags: []string{"all", "latest"}, ModelKits: &count},
		{Repo: "registry.example.com/my-org/my-model", Digest: testDigestB, Tags: []string{}},
	}

	var buf bytes.Buffer
	require.NoError(t, printIndexListJSON(&buf, listings))

	assert.JSONEq(t, `[
		{"repo": "registry.example.com/my-org/my-model", "digest": "`+testDigest+`", "tags": ["all", "latest"], "modelKits": 3},
		{"repo": "registry.example.com/my-org/my-model", "digest": "`+testDigestB+`", "tags": [], "modelKits": null}
	]`, buf.String())
}
