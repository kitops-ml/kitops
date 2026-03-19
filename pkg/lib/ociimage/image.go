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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"
)

const modelsDir = "models"

type BuildOptions struct {
	SourceDir string

	OutputDir string

	OS string

	Architecture string
}

func (o *BuildOptions) applyDefaults() {
	if o.OS == "" {
		o.OS = "linux"
	}
	if o.Architecture == "" {
		o.Architecture = "amd64"
	}
}

func BuildImage(ctx context.Context, opts BuildOptions) (ocispec.Descriptor, error) {
	opts.applyDefaults()

	if opts.SourceDir == "" {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("SourceDir must not be empty")
	}
	if opts.OutputDir == "" {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("OutputDir must not be empty")
	}
	if _, err := os.Stat(opts.SourceDir); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("source directory %s: %w", opts.SourceDir, err)
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to create output directory: %w", err)
	}

	store, err := oci.New(opts.OutputDir)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to create OCI layout store: %w", err)
	}

	layerDesc, diffID, err := buildLayer(ctx, store, opts.SourceDir)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to build image layer: %w", err)
	}

	configDesc, err := buildConfig(ctx, store, opts.OS, opts.Architecture, diffID)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to build image config: %w", err)
	}

	manifestDesc, err := buildManifest(ctx, store, configDesc, layerDesc)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to build image manifest: %w", err)
	}

	return manifestDesc, nil
}

func buildLayer(ctx context.Context, store *oci.Store, sourceDir string) (ocispec.Descriptor, digest.Digest, error) {
	pr, pw := io.Pipe()

	compressedDigester := digest.Canonical.Digester()
	uncompressedDigester := digest.Canonical.Digester()

	errCh := make(chan error, 1)

	go func() {
		defer pw.Close()

		mw := io.MultiWriter(pw, compressedDigester.Hash())
		gzw := gzip.NewWriter(mw)
		tw := tar.NewWriter(io.MultiWriter(gzw, uncompressedDigester.Hash()))

		err := writeDirToTar(tw, sourceDir)
		if closeErr := tw.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if closeErr := gzw.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		errCh <- err
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, pr); err != nil {
		return ocispec.DescriptorEmptyJSON, "", fmt.Errorf("failed to buffer layer: %w", err)
	}

	if err := <-errCh; err != nil {
		return ocispec.DescriptorEmptyJSON, "", fmt.Errorf("failed to write tar layer: %w", err)
	}

	layerBytes := buf.Bytes()
	layerDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    compressedDigester.Digest(),
		Size:      int64(len(layerBytes)),
	}

	if exists, err := store.Exists(ctx, layerDesc); err != nil {
		return ocispec.DescriptorEmptyJSON, "", err
	} else if !exists {
		if err := store.Push(ctx, layerDesc, bytes.NewReader(layerBytes)); err != nil {
			return ocispec.DescriptorEmptyJSON, "", fmt.Errorf("failed to push layer blob: %w", err)
		}
	}

	return layerDesc, uncompressedDigester.Digest(), nil
}

func writeDirToTar(tw *tar.Writer, sourceDir string) error {
	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     modelsDir + "/",
		Mode:     0o755,
	}); err != nil {
		return fmt.Errorf("failed to write %s/ header: %w", modelsDir, err)
	}

	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() && !info.Mode().IsDir() {
			return nil
		}

		tarName := filepath.ToSlash(filepath.Join(modelsDir, rel))
		if info.IsDir() {
			tarName += "/"
		}

		hdr := &tar.Header{
			Name:    tarName,
			Mode:    int64(info.Mode().Perm()),
			ModTime: time.Time{},
		}
		if info.IsDir() {
			hdr.Typeflag = tar.TypeDir
		} else {
			hdr.Typeflag = tar.TypeReg
			hdr.Size = info.Size()
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed to write header for %s: %w", tarName, err)
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("failed to write %s to tar: %w", path, err)
		}
		return nil
	})
}

type ociImageConfig struct {
	Architecture string          `json:"architecture"`
	OS           string          `json:"os"`
	RootFS       ociImageRootFS  `json:"rootfs"`
	Config       ociImageRunConf `json:"config,omitempty"`
}

type ociImageRootFS struct {
	Type    string          `json:"type"`
	DiffIDs []digest.Digest `json:"diff_ids"`
}

type ociImageRunConf struct {
	Cmd []string `json:"Cmd,omitempty"`
}

func buildConfig(ctx context.Context, store *oci.Store, osName, arch string, diffID digest.Digest) (ocispec.Descriptor, error) {
	cfg := ociImageConfig{
		Architecture: arch,
		OS:           osName,
		RootFS: ociImageRootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{diffID},
		},
		Config: ociImageRunConf{
			Cmd: []string{"sleep", "infinity"},
		},
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to marshal image config: %w", err)
	}

	cfgDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(cfgBytes),
		Size:      int64(len(cfgBytes)),
	}

	if exists, err := store.Exists(ctx, cfgDesc); err != nil {
		return ocispec.DescriptorEmptyJSON, err
	} else if !exists {
		if err := store.Push(ctx, cfgDesc, bytes.NewReader(cfgBytes)); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to push image config: %w", err)
		}
	}

	return cfgDesc, nil
}

func buildManifest(ctx context.Context, store *oci.Store, configDesc, layerDesc ocispec.Descriptor) (ocispec.Descriptor, error) {
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    configDesc,
		Layers:    []ocispec.Descriptor{layerDesc},
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}

	if exists, err := store.Exists(ctx, manifestDesc); err != nil {
		return ocispec.DescriptorEmptyJSON, err
	} else if !exists {
		if err := store.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes)); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to push manifest: %w", err)
		}
	}

	return manifestDesc, nil
}

func ModelsDir() string {
	return modelsDir
}

func ParseLayerEntries(r io.Reader) ([]string, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, hdr.Name)
	}
	return names, nil
}

func readBlobFromStore(store *oci.Store, desc ocispec.Descriptor) ([]byte, error) {
	rc, err := store.Fetch(context.Background(), desc)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

