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

package remote

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewRepositoryUploadChunkSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		repositoryOpts []RepositoryOption
		wantSize       int64
		wantFormat     uploadFormat
		wantErr        bool
	}{
		{name: "default", wantSize: 100 << 20, wantFormat: uploadMonolithicPut},
		{name: "16 MiB", repositoryOpts: []RepositoryOption{WithUploadChunkSize(16 << 20)}, wantSize: 16 << 20, wantFormat: uploadChunkedPatch},
		{name: "32 MiB", repositoryOpts: []RepositoryOption{WithUploadChunkSize(32 << 20)}, wantSize: 32 << 20, wantFormat: uploadChunkedPatch},
		{name: "100 MiB", repositoryOpts: []RepositoryOption{WithUploadChunkSize(100 << 20)}, wantSize: 100 << 20, wantFormat: uploadMonolithicPut},
		{name: "128 MiB", repositoryOpts: []RepositoryOption{WithUploadChunkSize(128 << 20)}, wantSize: 128 << 20, wantFormat: uploadMonolithicPut},
		{name: "zero", repositoryOpts: []RepositoryOption{WithUploadChunkSize(0)}, wantErr: true},
		{name: "negative", repositoryOpts: []RepositoryOption{WithUploadChunkSize(-1)}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewRepository(t.Context(), "registry.example.com", "test/repo", options.DefaultNetworkOptions(t.TempDir()), tt.repositoryOpts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, repo)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			configuredRepo := repo.(*Repository)
			assert.Equal(t, tt.wantSize, configuredRepo.uploadChunkSize)
			assert.Equal(t, tt.wantFormat, getUploadFormat("registry.example.com", 50<<20, configuredRepo.uploadChunkSize))
		})
	}
}

func TestRepositoryPushWithUploadChunkSize(t *testing.T) {
	t.Parallel()

	const uploadURL = "https://registry.example.com/v2/test/repo/blobs/uploads/test"
	const blobURL = "https://registry.example.com/v2/test/repo/blobs/test"
	content := []byte("hello world")
	expected := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayer,
		Digest:    digest.FromBytes(content),
		Size:      int64(len(content)),
	}
	repo, err := NewRepository(t.Context(), "registry.example.com", "test/repo", options.DefaultNetworkOptions(t.TempDir()), WithUploadChunkSize(5))
	if err != nil {
		t.Fatal(err)
	}

	responses := []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			return makeResponse(req, http.StatusAccepted, "", map[string]string{"Location": uploadURL})
		},
	}
	for _, chunk := range []struct{ body, contentRange, responseRange string }{
		{"hello", "0-4", "0-4"},
		{" worl", "5-9", "0-9"},
		{"d", "10-10", "0-10"},
	} {
		responses = append(responses, func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPatch, req.Method)
			assert.Equal(t, uploadURL, req.URL.String())
			assert.Equal(t, int64(len(chunk.body)), req.ContentLength)
			assert.Equal(t, chunk.contentRange, req.Header.Get("Content-Range"))
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body.Close()
			assert.Equal(t, chunk.body, string(body))
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": uploadURL,
				"Range":    chunk.responseRange,
			})
		})
	}
	responses = append(responses, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Zero(t, req.ContentLength)
		assert.Equal(t, expected.Digest.String(), req.URL.Query().Get("digest"))
		return makeResponse(req, http.StatusCreated, "", map[string]string{"Location": blobURL})
	})
	client := &testClient{responses: responses}
	repo.(*Repository).Client = client
	assert.NoError(t, repo.Push(t.Context(), expected, bytes.NewReader(content)))
	assert.Equal(t, len(responses), client.responseIdx)
}
