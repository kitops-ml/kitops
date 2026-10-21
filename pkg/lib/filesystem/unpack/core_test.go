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

package unpack

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
)

// fixedBlobStore returns blob for every Fetch, regardless of the descriptor requested.
type fixedBlobStore struct {
	blob []byte
}

func (s *fixedBlobStore) Fetch(context.Context, ocispec.Descriptor) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.blob)), nil
}

func (s *fixedBlobStore) Exists(context.Context, ocispec.Descriptor) (bool, error) {
	return true, nil
}

func (s *fixedBlobStore) Resolve(context.Context, string) (ocispec.Descriptor, error) {
	return ocispec.Descriptor{}, errdef.ErrNotFound
}

func tarBlob(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(data)), Typeflag: tar.TypeReg}))
	_, err := tw.Write(data)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	return buf.Bytes()
}

func TestUnpackLayer_VerifiesDigest(t *testing.T) {
	data := []byte("a,b\n1,2\n")
	tests := []struct {
		name      string
		mediaType string
		blob      []byte
	}{
		{name: "raw", mediaType: "application/vnd.kitops.modelkit.dataset.v1.raw", blob: data},
		{name: "tar", mediaType: "application/vnd.kitops.modelkit.dataset.v1.tar", blob: tarBlob(t, "train.csv", data)},
	}
	for _, tt := range tests {
		mt, err := mediatype.ParseMediaType(tt.mediaType)
		require.NoError(t, err)
		desc := ocispec.Descriptor{
			MediaType:   tt.mediaType,
			Digest:      digest.FromBytes(tt.blob),
			Size:        int64(len(tt.blob)),
			Annotations: map[string]string{modelspecv1.AnnotationFilepath: "train.csv"},
		}
		tampered := bytes.Replace(tt.blob, []byte("1,2"), []byte("9,2"), 1)
		require.Len(t, tampered, len(tt.blob))

		t.Run(tt.name+"/valid", func(t *testing.T) {
			t.Chdir(t.TempDir())
			err := unpackLayer(t.Context(), &fixedBlobStore{blob: tt.blob}, desc, "", false, false, mt)
			assert.NoError(t, err)
			assert.FileExists(t, "train.csv")
		})
		t.Run(tt.name+"/tampered", func(t *testing.T) {
			t.Chdir(t.TempDir())
			err := unpackLayer(t.Context(), &fixedBlobStore{blob: tampered}, desc, "", false, false, mt)
			assert.ErrorIs(t, err, content.ErrMismatchedDigest)
			assert.NoFileExists(t, "train.csv")
		})
	}
}
