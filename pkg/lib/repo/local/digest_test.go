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

package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestBlobPathForManifest(t *testing.T) {
	tests := []struct {
		name          string
		digest        digest.Digest
		wantAlgorithm string
	}{
		{
			name:          "SHA-256 digest",
			digest:        "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantAlgorithm: "sha256",
		},
		{
			name:          "SHA-512 digest",
			digest:        "sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			wantAlgorithm: "sha512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &LocalStore{storePath: "/tmp/test-store"}
			desc := ocispec.Descriptor{Digest: tt.digest}

			got := BlobPathForManifest(store, desc)
			wantSuffix := filepath.Join(ocispec.ImageBlobsDir, tt.wantAlgorithm, tt.digest.Encoded())
			if !strings.HasSuffix(got, wantSuffix) {
				t.Errorf("BlobPathForManifest() = %q, want suffix %q", got, wantSuffix)
			}
		})
	}
}

func TestBlobPath(t *testing.T) {
	tests := []struct {
		name          string
		digest        digest.Digest
		wantAlgorithm string
	}{
		{
			name:          "SHA-256 digest",
			digest:        "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantAlgorithm: "sha256",
		},
		{
			name:          "SHA-512 digest",
			digest:        "sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			wantAlgorithm: "sha512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &localRepo{storagePath: "/tmp/test-repo"}
			desc := ocispec.Descriptor{Digest: tt.digest}

			got := repo.BlobPath(desc)
			wantSuffix := filepath.Join(ocispec.ImageBlobsDir, tt.wantAlgorithm, tt.digest.Encoded())
			if !strings.HasSuffix(got, wantSuffix) {
				t.Errorf("BlobPath() = %q, want suffix %q", got, wantSuffix)
			}
		})
	}
}

func TestEnsureDirs(t *testing.T) {
	tests := []struct {
		name          string
		digest        digest.Digest
		wantAlgorithm string
	}{
		{
			name:          "SHA-256 creates sha256 directory",
			digest:        "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantAlgorithm: "sha256",
		},
		{
			name:          "SHA-512 creates sha512 directory",
			digest:        "sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			wantAlgorithm: "sha512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			repo := &localRepo{storagePath: tmpDir}
			desc := ocispec.Descriptor{Digest: tt.digest}

			if err := repo.EnsureDirs(desc); err != nil {
				t.Fatalf("EnsureDirs() error = %v", err)
			}

			expectedDir := filepath.Join(tmpDir, ocispec.ImageBlobsDir, tt.wantAlgorithm)
			if !dirExists(expectedDir) {
				t.Errorf("EnsureDirs() did not create directory %q", expectedDir)
			}
		})
	}
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
