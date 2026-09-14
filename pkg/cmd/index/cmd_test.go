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
	libindex "github.com/kitops-ml/kitops/pkg/lib/index"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPrintModelMetadataFormatsValues(t *testing.T) {
	var buf bytes.Buffer
	printModelMetadata(&buf, libindex.ModelMetadata{
		"quantization":      json.RawMessage(`"q4_0"`),
		"activeParameters":  json.RawMessage(`7000000000`),
		"targetAccelerator": json.RawMessage(`["cuda","metal"]`),
	})

	out := buf.String()
	assert.Contains(t, out, "quantization:")
	assert.Contains(t, out, "q4_0")
	assert.NotContains(t, out, `"q4_0"`, "string values are printed unquoted")
	assert.Contains(t, out, "7000000000")
	assert.Contains(t, out, `["cuda","metal"]`)
	assert.Less(t, strings.Index(out, "activeParameters"), strings.Index(out, "quantization"),
		"keys are printed in sorted order")
}
