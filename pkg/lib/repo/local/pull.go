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
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

func (l *localRepo) PullModel(ctx context.Context, src oras.ReadOnlyTarget, ref registry.Reference, opts *options.NetworkOptions) (ocispec.Descriptor, error) {
	// Only support pulling image manifests
	desc, err := src.Resolve(ctx, ref.Reference)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	if desc.MediaType != ocispec.MediaTypeImageManifest {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("expected manifest for pull but got %s", desc.MediaType)
	}

	if err := l.ensurePullDirs(); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to set up directories for pull: %w", err)
	}

	progress := output.NewPullProgress(ctx)

	manifest, err := util.GetManifest(ctx, src, desc)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}

	toPull := []ocispec.Descriptor{manifest.Config}
	toPull = append(toPull, manifest.Layers...)
	toPull = append(toPull, desc)
	sem := semaphore.NewWeighted(int64(opts.Concurrency))
	errs, errCtx := errgroup.WithContext(ctx)
	fmtErr := func(desc ocispec.Descriptor, err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("failed to get %s layer: %w", mediatype.FormatMediaTypeForUser(desc.MediaType), err)
	}
	var semErr error
	// In some cases, manifests can contain duplicate digests. If we try to concurrently pull the same digest
	// twice, a race condition will cause the pull the fail.
	pulledDigests := map[string]bool{}
	for _, pullDesc := range toPull {
		pullDesc := pullDesc
		digest := pullDesc.Digest.String()
		if pulledDigests[digest] {
			continue
		}
		pulledDigests[digest] = true
		if err := sem.Acquire(errCtx, 1); err != nil {
			// Save error and break to get the _actual_ error
			semErr = err
			break
		}
		errs.Go(func() error {
			defer sem.Release(1)
			return fmtErr(pullDesc, l.pullNode(errCtx, src, pullDesc, progress))
		})
	}
	if err := errs.Wait(); err != nil {
		return ocispec.DescriptorEmptyJSON, err
	}
	if semErr != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to acquire lock: %w", semErr)
	}

	// Special handling to make sure local (scoped) repo contains the just-pulled manifest
	if err := l.localIndex.addManifest(desc); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to add manifest to index: %w", err)
	}
	// This is a workaround to add the manifest to the main index as well; this is necessary for garbage collection to work
	if err := l.Store.Tag(ctx, desc, desc.Digest.String()); err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to add manifest to shared index: %w", err)
	}

	if !util.ReferenceIsDigest(ref.Reference) {
		if err := l.localIndex.tag(desc, ref.Reference); err != nil {
			return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to save tag: %w", err)
		}
	}
	progress.Done()

	if err := l.cleanupIngestDir(); err != nil {
		output.Logln(output.LogLevelWarn, err)
	}

	return desc, nil
}

func (l *localRepo) pullNode(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor, p *output.PullProgress) error {
	if exists, err := l.Exists(ctx, desc); err != nil {
		return fmt.Errorf("failed to check local storage: %w", err)
	} else if exists {
		return nil
	}

	blob, err := src.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	if seekBlob, ok := blob.(io.ReadSeekCloser); ok {
		p.Logf(output.LogLevelTrace, "Remote supports range requests, using resumable download")
		return l.resumeAndDownloadFile(desc, seekBlob, p)
	} else {
		return l.downloadFile(desc, blob, p)
	}
}

func (l *localRepo) resumeAndDownloadFile(desc ocispec.Descriptor, blob io.ReadSeekCloser, p *output.PullProgress) error {
	ingestDir := constants.IngestPath(l.storagePath)
	ingestFilename := filepath.Join(ingestDir, desc.Digest.Encoded())
	ingestFile, err := os.OpenFile(ingestFilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ingest file for writing: %w", err)
	}
	defer func() {
		if err := ingestFile.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			p.Logf(output.LogLevelError, "Error closing temporary ingest file: %s", err)
		}
	}()

	verifier := desc.Digest.Verifier()
	var offset int64 = 0
	if stat, err := ingestFile.Stat(); err != nil {
		return fmt.Errorf("failed to stat ingest file: %w", err)
	} else if stat.Size() != 0 {
		p.Debugf("Resuming download for digest %s", desc.Digest.String())
		numBytes, err := io.Copy(verifier, ingestFile)
		if err != nil {
			return fmt.Errorf("failed to resume download: %w", err)
		}
		p.Logf(output.LogLevelTrace, "Updating offset to %d bytes", numBytes)
		offset = numBytes
	}
	if _, err := blob.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek in remote resource: %w", err)
	}

	pwriter := p.ProxyWriter(ingestFile, desc.Digest.Encoded(), desc.Size, offset)
	mw := io.MultiWriter(pwriter, verifier)
	if _, err := io.Copy(mw, blob); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if !verifier.Verified() {
		return fmt.Errorf("downloaded file hash does not match descriptor")
	}
	if err := ingestFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary ingest file: %w", err)
	}
	blobPath := l.BlobPath(desc)
	if err := os.Rename(ingestFilename, blobPath); err != nil {
		return fmt.Errorf("failed to move downloaded file into storage: %w", err)
	}
	if err := os.Chmod(blobPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions on blob: %w", err)
	}

	return nil
}

func (l *localRepo) downloadFile(desc ocispec.Descriptor, blob io.ReadCloser, p *output.PullProgress) (ingestErr error) {
	ingestDir := constants.IngestPath(l.storagePath)
	ingestFile, err := os.CreateTemp(ingestDir, desc.Digest.Encoded()+"_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary ingest file: %w", err)
	}

	ingestFilename := ingestFile.Name()
	// If we return an error anywhere after this point, we want to delete the ingest file we're
	// working on, since it will never be reused.
	defer func() {
		if err := ingestFile.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			p.Logf(output.LogLevelError, "Error closing temporary ingest file: %s", err)
		}
		if ingestErr != nil {
			os.Remove(ingestFilename)
		}
	}()

	verifier := desc.Digest.Verifier()
	pwriter := p.ProxyWriter(ingestFile, desc.Digest.Encoded(), desc.Size, 0)
	mw := io.MultiWriter(pwriter, verifier)
	if _, err := io.Copy(mw, blob); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if !verifier.Verified() {
		return fmt.Errorf("downloaded file hash does not match descriptor")
	}
	if err := ingestFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary ingest file: %w", err)
	}

	blobPath := l.BlobPath(desc)
	if err := os.Rename(ingestFilename, blobPath); err != nil {
		return fmt.Errorf("failed to move downloaded file into storage: %w", err)
	}
	if err := os.Chmod(blobPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions on blob: %w", err)
	}

	return nil
}

func (l *localRepo) ensurePullDirs() error {
	blobsPath := filepath.Join(l.storagePath, ocispec.ImageBlobsDir, "sha256")
	if err := os.MkdirAll(blobsPath, 0755); err != nil {
		return err
	}
	ingestPath := constants.IngestPath(l.storagePath)
	return os.MkdirAll(ingestPath, 0755)
}

func (l *localRepo) cleanupIngestDir() error {
	ingestPath := constants.IngestPath(l.storagePath)
	err := filepath.WalkDir(ingestPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to clean up ingest directory: %w", err)
	}
	return nil
}
