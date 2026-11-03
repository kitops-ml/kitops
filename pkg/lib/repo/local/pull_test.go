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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

func TestPullReloadsIndexes(t *testing.T) {
	for _, sameRepo := range []bool{false, true} {
		name := "different repositories"
		if sameRepo {
			name = "same repository"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			src := testRepo(t)
			first := pushModelKit(t, src, "v1", "first layer")
			second := pushModelKit(t, src, "v2", "second layer")
			storage := t.TempDir()
			refA := &registry.Reference{Registry: "localhost", Repository: "a", Reference: "v1"}
			refB := &registry.Reference{Registry: "localhost", Repository: "b", Reference: "v2"}
			if sameRepo {
				refB.Repository = refA.Repository
			}
			a, err := NewLocalRepo(storage, refA)
			require.NoError(t, err)
			b, err := NewLocalRepo(storage, refB)
			require.NoError(t, err)
			_, err = a.PullModel(ctx, src, *refA, &options.NetworkOptions{Concurrency: 2})
			require.NoError(t, err)
			_, err = b.PullModel(ctx, src, *refB, &options.NetworkOptions{Concurrency: 2})
			require.NoError(t, err)
			for _, item := range []struct {
				ref  *registry.Reference
				desc ocispec.Descriptor
			}{{refA, first}, {refB, second}} {
				reopened, err := NewLocalRepo(storage, item.ref)
				require.NoError(t, err)
				resolved, err := reopened.Resolve(ctx, item.ref.Reference)
				require.NoError(t, err)
				require.Equal(t, item.desc.Digest, resolved.Digest)
			}
			data, err := os.ReadFile(filepath.Join(storage, "index.json"))
			require.NoError(t, err)
			var index ocispec.Index
			require.NoError(t, json.Unmarshal(data, &index))
			require.Len(t, index.Manifests, 2)
		})
	}
}

type blockedSource struct {
	resumable bool
	oras.ReadOnlyTarget
	digest  digest.Digest
	entered chan struct{}
	release chan struct{}
}

func (s *blockedSource) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	reader, err := s.ReadOnlyTarget.Fetch(ctx, desc)
	if err != nil || desc.Digest != s.digest {
		return reader, err
	}
	if s.resumable {
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, err
		}
		seek := &seekBlob{Reader: bytes.NewReader(data)}
		return &blockedSeekReader{blockedReader: &blockedReader{ReadCloser: seek, entered: s.entered, release: s.release}, Seeker: seek}, nil
	}
	return &blockedReader{ReadCloser: reader, entered: s.entered, release: s.release}, nil
}

type blockedSeekReader struct {
	*blockedReader
	io.Seeker
}

type blockedReader struct {
	io.ReadCloser
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func (r *blockedReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.entered); <-r.release })
	return r.ReadCloser.Read(p)
}

func TestPullWaitsWithoutDeletingActiveDownload(t *testing.T) {
	for _, resumable := range []bool{false, true} {
		name := "non-resumable"
		if resumable {
			name = "resumable"
		}
		t.Run(name, func(t *testing.T) { testPullWaitsWithoutDeletingActiveDownload(t, resumable) })
	}
}

func testPullWaitsWithoutDeletingActiveDownload(t *testing.T, resumable bool) {
	ctx := context.Background()
	src := testRepo(t)
	first := pushModelKit(t, src, "v1", "first layer")
	second := pushModelKit(t, src, "v2", "second layer")
	manifest, err := getManifest(t, src, first)
	require.NoError(t, err)
	refA := &registry.Reference{Registry: "localhost", Repository: "a", Reference: "v1"}
	refB := &registry.Reference{Registry: "localhost", Repository: "b", Reference: "v2"}
	storage := t.TempDir()
	a, err := NewLocalRepo(storage, refA)
	require.NoError(t, err)
	b, err := NewLocalRepo(storage, refB)
	require.NoError(t, err)
	other, err := NewLocalRepo(t.TempDir(), refB)
	require.NoError(t, err)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	blocked := &blockedSource{resumable: resumable, ReadOnlyTarget: src, digest: manifest.Layers[0].Digest, entered: make(chan struct{}), release: release}
	done := make(chan error, 1)
	go func() {
		_, err := a.PullModel(ctx, blocked, *refA, &options.NetworkOptions{Concurrency: 1})
		done <- err
	}()
	select {
	case <-blocked.entered:
	case <-time.After(10 * time.Second):
		t.Fatal("first pull did not start its layer")
	}
	files, err := os.ReadDir(constants.IngestPath(storage))
	require.NoError(t, err)
	require.Len(t, files, 1)
	waiting, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	_, err = b.PullModel(waiting, src, *refB, &options.NetworkOptions{Concurrency: 1})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	_, err = os.Stat(filepath.Join(constants.IngestPath(storage), files[0].Name()))
	require.NoError(t, err)
	independent, cancelIndependent := context.WithTimeout(ctx, 5*time.Second)
	defer cancelIndependent()
	_, err = other.PullModel(independent, src, *refB, &options.NetworkOptions{Concurrency: 1})
	require.NoError(t, err)
	unblock()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("first pull did not complete")
	}
	_, err = b.PullModel(ctx, src, *refB, &options.NetworkOptions{Concurrency: 1})
	require.NoError(t, err)
	for _, desc := range []ocispec.Descriptor{first, second} {
		require.FileExists(t, a.BlobPath(desc))
	}
}

type interruptedBlob struct{ reader *bytes.Reader }

func (b *interruptedBlob) Close() error { return nil }
func (b *interruptedBlob) Seek(offset int64, whence int) (int64, error) {
	return b.reader.Seek(offset, whence)
}
func (b *interruptedBlob) Read(p []byte) (int, error) {
	if b.reader.Len() <= 4 {
		return 0, io.ErrUnexpectedEOF
	}
	if len(p) > 4 {
		p = p[:4]
	}
	return b.reader.Read(p)
}

func TestPullResumesAfterInterruptedDownload(t *testing.T) {
	repo := testRepo(t).(*localRepo)
	require.NoError(t, repo.ensurePullDirs())
	content := []byte("download content")
	desc := ocispec.Descriptor{Digest: digest.FromBytes(content), Size: int64(len(content))}
	progress := output.NewPullProgress(context.Background())
	err := repo.resumeAndDownloadFile(desc, &interruptedBlob{bytes.NewReader(content)}, progress)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	ingest := filepath.Join(constants.IngestPath(repo.storagePath), desc.Digest.Encoded())
	partial, err := os.ReadFile(ingest)
	require.NoError(t, err)
	require.NotEmpty(t, partial)
	require.Less(t, len(partial), len(content))
	reader := &seekBlob{Reader: bytes.NewReader(content)}
	require.NoError(t, repo.resumeAndDownloadFile(desc, reader, progress))
	require.Equal(t, int64(len(partial)), reader.offset)
	data, err := os.ReadFile(repo.BlobPath(desc))
	require.NoError(t, err)
	require.Equal(t, content, data)
}

type seekBlob struct {
	*bytes.Reader
	offset int64
}

func (b *seekBlob) Close() error { return nil }
func (b *seekBlob) Seek(offset int64, whence int) (int64, error) {
	b.offset = offset
	return b.Reader.Seek(offset, whence)
}

func TestStorageLockReleasedAfterProcessExit(t *testing.T) {
	if storage := os.Getenv("KITOPS_TEST_LOCK_STORAGE"); storage != "" {
		lock, err := lockStorage(context.Background(), storage)
		require.NoError(t, err)
		defer lock.Close()
		fmt.Println("locked")
		_, _ = os.Stdin.Read(make([]byte, 1))
		return
	}
	storage := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStorageLockReleasedAfterProcessExit$")
	cmd.Env = append(os.Environ(), "KITOPS_TEST_LOCK_STORAGE="+storage)
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	line, err := bufio.NewReader(stdout).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "locked\n", line)
	waiting, cancelWait := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancelWait()
	_, err = lockStorage(waiting, storage)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NoError(t, cmd.Process.Kill())
	require.Error(t, cmd.Wait())
	lock, err := lockStorage(ctx, storage)
	require.NoError(t, err)
	require.NoError(t, lock.Close())
}

func TestOpenExistingRepoDuringPull(t *testing.T) {
	storage := t.TempDir()
	_, err := NewLocalRepo(storage, testRef)
	require.NoError(t, err)
	lock, err := lockStorage(context.Background(), storage)
	require.NoError(t, err)
	defer lock.Close()
	result := make(chan error, 1)
	go func() { _, err := NewLocalRepo(storage, testRef); result <- err }()
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("opening an existing repository waited for the pull lock")
	}
}

func TestOpenReadOnlyRepoWithoutLockFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permissions")
	}
	storage := t.TempDir()
	_, err := NewLocalRepo(storage, testRef)
	require.NoError(t, err)
	require.NoError(t, os.Remove(filepath.Join(storage, ".pull.lock")))
	require.NoError(t, os.Chmod(storage, 0555))
	defer os.Chmod(storage, 0755)
	// Root can write despite the directory mode.
	probe := filepath.Join(storage, "permission-probe")
	if err := os.WriteFile(probe, nil, 0600); err == nil {
		os.Remove(probe)
		t.Skip("directory permissions are not enforced for this user")
	}
	_, err = NewLocalRepo(storage, testRef)
	require.NoError(t, err)
}

func TestFetchDuringPull(t *testing.T) {
	ctx := context.Background()
	src := testRepo(t)
	desc := pushModelKit(t, src, "v1", "first layer")
	manifest, err := getManifest(t, src, desc)
	require.NoError(t, err)
	storage := t.TempDir()
	ref := &registry.Reference{Registry: "localhost", Repository: "a", Reference: "v1"}
	dst, err := NewLocalRepo(storage, ref)
	require.NoError(t, err)
	_, err = dst.PullModel(ctx, src, *ref, &options.NetworkOptions{Concurrency: 2})
	require.NoError(t, err)
	var stop atomic.Bool
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		for !stop.Load() {
			if _, err := dst.Resolve(ctx, "v1"); err != nil {
				done <- err
				return
			}
			_ = dst.GetAllModels()
			_ = dst.GetTags(desc)
			if _, err := dst.(*localRepo).Predecessors(ctx, manifest.Layers[0]); err != nil {
				done <- err
				return
			}
			r, err := dst.Fetch(ctx, manifest.Layers[0])
			if err != nil {
				done <- err
				return
			}
			_, err = io.Copy(io.Discard, r)
			r.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	<-started
	defer func() { stop.Store(true); require.NoError(t, <-done) }()
	pulls := make(chan error, 2)
	for range 2 {
		go func() {
			for range 10 {
				if _, err := dst.PullModel(ctx, src, *ref, &options.NetworkOptions{Concurrency: 2}); err != nil {
					pulls <- err
					return
				}
			}
			pulls <- nil
		}()
	}
	for range 2 {
		require.NoError(t, <-pulls)
	}
}

func TestFailedPullReleasesStorageLock(t *testing.T) {
	src := testRepo(t)
	desc := pushModelKit(t, src, "v1", "layer")
	ref := &registry.Reference{Registry: "localhost", Repository: "a", Reference: "missing"}
	dst, err := NewLocalRepo(t.TempDir(), ref)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = dst.PullModel(ctx, src, *ref, &options.NetworkOptions{Concurrency: 1})
	require.Error(t, err)
	ref.Reference = "v1"
	got, err := dst.PullModel(ctx, src, *ref, &options.NetworkOptions{Concurrency: 1})
	require.NoError(t, err)
	require.Equal(t, desc.Digest, got.Digest)
}
