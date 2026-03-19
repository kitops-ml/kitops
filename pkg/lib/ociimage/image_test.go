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

package ociimage

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content/oci"
)

func makeSourceDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
	return dir
}

func TestBuildImageValidOptions(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		os          string
		arch        string
		wantOS      string
		wantArch    string
		wantEntries []string
	}{
		{
			name: "single file uses defaults",
			files: map[string]string{
				"weights.bin": "fake model data",
			},
			wantOS:   "linux",
			wantArch: "amd64",
			wantEntries: []string{
				"models/",
				"models/weights.bin",
			},
		},
		{
			name: "nested directory structure",
			files: map[string]string{
				"config.json":         `{"model":"test"}`,
				"weights/shard-0.bin": "shard0",
				"weights/shard-1.bin": "shard1",
			},
			wantOS:   "linux",
			wantArch: "amd64",
			wantEntries: []string{
				"models/",
				"models/config.json",
				"models/weights/",
				"models/weights/shard-0.bin",
				"models/weights/shard-1.bin",
			},
		},
		{
			name: "custom os and arch",
			files: map[string]string{
				"model.gguf": "gguf data",
			},
			os:       "linux",
			arch:     "arm64",
			wantOS:   "linux",
			wantArch: "arm64",
			wantEntries: []string{
				"models/",
				"models/model.gguf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcDir := makeSourceDir(t, tt.files)
			outDir := t.TempDir()

			desc, err := BuildImage(context.Background(), BuildOptions{
				SourceDir:    srcDir,
				OutputDir:    outDir,
				OS:           tt.os,
				Architecture: tt.arch,
			})
			require.NoError(t, err)

			assert.Equal(t, ocispec.MediaTypeImageManifest, desc.MediaType)
			assert.NotEmpty(t, desc.Digest)
			assert.Greater(t, desc.Size, int64(0))

			store, err := oci.New(outDir)
			require.NoError(t, err)

			manifestBytes, err := readBlobFromStore(store, desc)
			require.NoError(t, err)

			var manifest ocispec.Manifest
			require.NoError(t, json.Unmarshal(manifestBytes, &manifest))

			assert.Equal(t, ocispec.MediaTypeImageConfig, manifest.Config.MediaType)
			assert.Len(t, manifest.Layers, 1)
			assert.Equal(t, ocispec.MediaTypeImageLayerGzip, manifest.Layers[0].MediaType)

			cfgBytes, err := readBlobFromStore(store, manifest.Config)
			require.NoError(t, err)

			var cfg ociImageConfig
			require.NoError(t, json.Unmarshal(cfgBytes, &cfg))

			assert.Equal(t, tt.wantOS, cfg.OS)
			assert.Equal(t, tt.wantArch, cfg.Architecture)
			assert.Equal(t, "layers", cfg.RootFS.Type)
			assert.Len(t, cfg.RootFS.DiffIDs, 1)
			assert.NotEmpty(t, cfg.RootFS.DiffIDs[0])

			layerBytes, err := readBlobFromStore(store, manifest.Layers[0])
			require.NoError(t, err)

			entries, err := ParseLayerEntries(bytes.NewReader(layerBytes))
			require.NoError(t, err)

			assert.ElementsMatch(t, tt.wantEntries, entries)
		})
	}
}

func TestBuildImageErrors(t *testing.T) {
	tests := []struct {
		name            string
		opts            BuildOptions
		wantErrContains string
	}{
		{
			name:            "empty SourceDir",
			opts:            BuildOptions{OutputDir: t.TempDir()},
			wantErrContains: "SourceDir must not be empty",
		},
		{
			name:            "empty OutputDir",
			opts:            BuildOptions{SourceDir: t.TempDir()},
			wantErrContains: "OutputDir must not be empty",
		},
		{
			name: "non-existent SourceDir",
			opts: BuildOptions{
				SourceDir: "/this/does/not/exist",
				OutputDir: t.TempDir(),
			},
			wantErrContains: "source directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildImage(context.Background(), tt.opts)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}

func TestBuildImageIdempotent(t *testing.T) {
	srcDir := makeSourceDir(t, map[string]string{
		"model.bin": "data",
	})
	outDir := t.TempDir()

	desc1, err := BuildImage(context.Background(), BuildOptions{
		SourceDir: srcDir,
		OutputDir: outDir,
	})
	require.NoError(t, err)

	desc2, err := BuildImage(context.Background(), BuildOptions{
		SourceDir: srcDir,
		OutputDir: outDir,
	})
	require.NoError(t, err)

	assert.Equal(t, desc1.Digest, desc2.Digest, "repeated builds must be deterministic")
}

func TestModelsDir(t *testing.T) {
	assert.Equal(t, "models", ModelsDir())
}

func TestParseLayerEntries(t *testing.T) {
	srcDir := makeSourceDir(t, map[string]string{
		"a.bin": "aaa",
		"b.bin": "bbb",
	})
	outDir := t.TempDir()

	desc, err := BuildImage(context.Background(), BuildOptions{
		SourceDir: srcDir,
		OutputDir: outDir,
	})
	require.NoError(t, err)

	store, err := oci.New(outDir)
	require.NoError(t, err)

	var manifest ocispec.Manifest
	manifestBytes, err := readBlobFromStore(store, desc)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))

	layerBytes, err := readBlobFromStore(store, manifest.Layers[0])
	require.NoError(t, err)

	entries, err := ParseLayerEntries(bytes.NewReader(layerBytes))
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"models/", "models/a.bin", "models/b.bin"}, entries)
}

