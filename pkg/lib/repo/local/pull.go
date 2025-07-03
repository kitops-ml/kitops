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
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

// downloadConfig holds all dynamically determined download configuration parameters
type downloadConfig struct {
	copyBufferSize        int
	largeLayerThreshold   int64
	chunkSize             int64
	chunkConcurrency      int64
	layerConcurrency      int
	adaptiveBufferEnabled bool
}

// getSystemMemory returns the total system memory in bytes using cross-platform approach
func getSystemMemory() int64 {
	// Try to get memory info from runtime stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// If we can get system memory info, use a conservative estimate
	// based on current heap + available memory heuristics
	if m.Sys > 0 {
		// Estimate total system memory as roughly 4-8x the current heap size
		// This is a rough heuristic but works across platforms
		estimatedTotal := int64(m.Sys) * 8

		// Clamp to reasonable bounds (1GB to 128GB)
		if estimatedTotal < 1*1024*1024*1024 {
			estimatedTotal = 8 * 1024 * 1024 * 1024 // Default to 8GB
		}
		if estimatedTotal > 128*1024*1024*1024 {
			estimatedTotal = 128 * 1024 * 1024 * 1024 // Cap at 128GB
		}

		return estimatedTotal
	}

	// Fallback default
	return 8 * 1024 * 1024 * 1024 // 8 GB
}

// determineOptimalConfig dynamically determines optimal download parameters
// based on available system resources
func determineOptimalConfig() downloadConfig {
	cpus := runtime.NumCPU()
	mem := getSystemMemory()

	// Basic heuristics - actual values will be adjusted based on file size and network conditions
	config := downloadConfig{
		adaptiveBufferEnabled: true,
	}

	// Buffer size: Scale with available memory (0.1% of RAM, with min/max bounds)
	// Larger buffers reduce syscall overhead but use more memory
	memoryFraction := mem / 1000         // 0.1% of total memory
	minBuffer := int64(1 * 1024 * 1024)  // 1 MB minimum
	maxBuffer := int64(16 * 1024 * 1024) // 16 MB maximum

	config.copyBufferSize = int(clampInt64(memoryFraction, minBuffer, maxBuffer))

	// Large layer threshold: Files above this size use chunked downloads
	// Scale with memory (smaller threshold on systems with more RAM)
	config.largeLayerThreshold = clampInt64(mem/200, 10*1024*1024, 100*1024*1024) // 0.5% of RAM, 10MB-100MB range

	// Chunk size: Scale with memory and CPUs
	// Larger chunks reduce overhead but use more memory per download
	basedOnMemory := mem / 100                                                                       // 1% of RAM per chunk
	basedOnCPUs := int64(16 * 1024 * 1024 * cpus)                                                    // Scale with CPU count
	config.chunkSize = clampInt64(minInt64(basedOnMemory, basedOnCPUs), 10*1024*1024, 200*1024*1024) // 10MB-200MB range

	// Chunk concurrency: Scale with CPUs and memory
	// Memory-bound for I/O operations
	memoryBasedConcurrency := mem / (200 * 1024 * 1024) // Assume each download might use up to 200MB
	cpuBasedConcurrency := int64(cpus * 4)              // 4 chunks per CPU is reasonable for I/O bound tasks
	config.chunkConcurrency = clampInt64(maxInt64(memoryBasedConcurrency, cpuBasedConcurrency), 4, 32)

	// Layer concurrency: How many files to download in parallel
	// Primarily limited by memory and network connection capacity
	memoryBasedLayerConcurrency := int(mem / (1024 * 1024 * 1024)) // Assume 1GB per layer
	config.layerConcurrency = clampInt(maxInt(memoryBasedLayerConcurrency, cpus*2), 4, 16)

	return config
}

// getNetworkAdjustedConfig monitors initial download speed and adjusts parameters
// to optimize for the current network conditions
func (l *localRepo) getNetworkAdjustedConfig(ctx context.Context, src oras.ReadOnlyTarget, initialConfig downloadConfig, desc ocispec.Descriptor, p *output.PullProgress) downloadConfig {
	config := initialConfig

	// Create a test download to measure network speed
	start := time.Now()
	testSize := int64(1024 * 1024) // 1MB test download

	// Only run the test if the file is large enough
	if desc.Size <= testSize*2 {
		return config // File too small to bother with network testing
	}

	// Get a small sample to measure network speed
	rc, err := l.fetchAndSeek(ctx, src, desc, 0, testSize)
	if err != nil {
		p.Logf(output.LogLevelDebug, "Skipping network speed test: %v", err)
		return config
	}
	defer rc.Close()

	// Discard the data, we just care about speed
	buf := make([]byte, 32*1024)
	n, err := io.CopyBuffer(io.Discard, rc, buf)
	elapsed := time.Since(start)

	if err != nil || n < testSize/2 || elapsed > 5*time.Second {
		// Network seems slow or unstable
		p.Logf(output.LogLevelDebug, "Network appears slow or unstable, optimizing for reliability")
		// Reduce concurrency and chunk size for more reliable downloads
		config.chunkConcurrency = maxInt64(4, config.chunkConcurrency/2)
		config.layerConcurrency = maxInt(2, config.layerConcurrency/2)
		config.chunkSize = maxInt64(5*1024*1024, config.chunkSize/2)
		return config
	}

	// Calculate speed in MB/s
	mbps := float64(n) / (1024 * 1024) / elapsed.Seconds()
	p.Logf(output.LogLevelDebug, "Network speed test: %.2f MB/s", mbps)

	// Adjust based on network speed
	if mbps > 50 {
		// Fast network - increase concurrency and chunk size
		config.chunkConcurrency = minInt64(config.chunkConcurrency*2, 64)
		config.layerConcurrency = minInt(config.layerConcurrency*2, 32)
		config.chunkSize = minInt64(config.chunkSize*2, 400*1024*1024)
	} else if mbps < 5 {
		// Slow network - reduce overhead
		config.chunkConcurrency = maxInt64(2, config.chunkConcurrency/2)
		config.layerConcurrency = maxInt(1, config.layerConcurrency/2)
		config.chunkSize = maxInt64(5*1024*1024, config.chunkSize/2)
	}

	return config
}

// Helper function to fetch and seek a remote resource
func (l *localRepo) fetchAndSeek(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor, offset, length int64) (io.ReadCloser, error) {
	rc, err := src.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}

	seeker, ok := rc.(io.ReadSeeker)
	if !ok {
		rc.Close()
		return nil, fmt.Errorf("remote does not support seeking")
	}

	if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
		rc.Close()
		return nil, err
	}

	return io.NopCloser(io.LimitReader(rc, length)), nil
}

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

	// Determine optimal configuration based on system resources
	config := determineOptimalConfig()
	progress.Logf(output.LogLevelDebug, "Dynamic config: buffer=%dKB, chunk=%dMB, concurrency=%d/%d",
		config.copyBufferSize/1024, config.chunkSize/(1024*1024),
		config.layerConcurrency, config.chunkConcurrency)

	// If concurrency wasn't explicitly set, use the dynamically determined value
	if opts.Concurrency <= 0 {
		opts.Concurrency = config.layerConcurrency
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
		return fmt.Errorf("failed to get %s layer: %w", constants.FormatMediaTypeForUser(desc.MediaType), err)
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
			return fmtErr(pullDesc, l.pullNode(errCtx, src, pullDesc, progress, config))
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

func (l *localRepo) pullNode(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor, p *output.PullProgress, config downloadConfig) error {
	if exists, err := l.Exists(ctx, desc); err != nil {
		return fmt.Errorf("failed to check local storage: %w", err)
	} else if exists {
		return nil
	}

	// For larger files, try to adjust configuration based on a quick network speed test
	if desc.Size > config.largeLayerThreshold {
		config = l.getNetworkAdjustedConfig(ctx, src, config, desc, p)
	}

	blob, err := src.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	if seekBlob, ok := blob.(io.ReadSeekCloser); ok {
		// For large layers where the remote supports seeking (which implies support for
		// HTTP Range requests), we use a new parallel chunking strategy to speed up
		// the download of the single layer.
		if desc.Size > config.largeLayerThreshold {
			p.Logf(output.LogLevelTrace, "Layer %s is large (%d bytes), using parallel chunk download", desc.Digest, desc.Size)
			// Close the initially fetched blob; the chunking function will manage its own fetches.
			seekBlob.Close()
			return l.downloadFileInChunks(ctx, src, desc, p, config)
		}

		// For smaller layers that support seeking, continue with the original resumable download logic.
		p.Logf(output.LogLevelTrace, "Remote supports range requests, using resumable download")
		return l.resumeAndDownloadFile(desc, seekBlob, p, config)
	} else {
		// If the remote does not support seeking, fall back to a simple, non-resumable download.
		return l.downloadFile(desc, blob, p, config)
	}
}

func (l *localRepo) resumeAndDownloadFile(desc ocispec.Descriptor, blob io.ReadSeekCloser, p *output.PullProgress, config downloadConfig) error {
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

	// Use io.CopyBuffer with a dynamically sized buffer
	buf := make([]byte, config.copyBufferSize)
	if _, err := io.CopyBuffer(mw, blob, buf); err != nil {
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

func (l *localRepo) downloadFile(desc ocispec.Descriptor, blob io.ReadCloser, p *output.PullProgress, config downloadConfig) (ingestErr error) {
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

	// Use io.CopyBuffer with dynamically sized buffer
	buf := make([]byte, config.copyBufferSize)
	if _, err := io.CopyBuffer(mw, blob, buf); err != nil {
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

// offsetWriter is a helper that implements io.Writer but writes at a specific offset.
type offsetWriter struct {
	w      io.WriterAt
	offset int64
}

func (ow *offsetWriter) Write(p []byte) (n int, err error) {
	n, err = ow.w.WriteAt(p, ow.offset)
	ow.offset += int64(n)
	return
}

// downloadFileInChunks implements a parallel download strategy for large files. It splits
// the file into chunks and downloads them concurrently. This is particularly effective
// for maximizing bandwidth utilization on high-speed networks.
func (l *localRepo) downloadFileInChunks(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor, p *output.PullProgress, config downloadConfig) (err error) {
	ingestDir := constants.IngestPath(l.storagePath)
	ingestFile, err := os.CreateTemp(ingestDir, desc.Digest.Encoded()+"_chunked_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary ingest file: %w", err)
	}
	ingestFilename := ingestFile.Name()
	defer func() {
		ingestFile.Close()
		if err != nil {
			os.Remove(ingestFilename)
		}
	}()

	// Pre-allocate the file to its full size
	if err := ingestFile.Truncate(desc.Size); err != nil {
		return fmt.Errorf("failed to pre-allocate file space: %w", err)
	}

	// Dynamically adjust chunk size based on file size
	chunkSize := config.chunkSize
	if desc.Size > 10*1024*1024*1024 { // 10 GB
		// For very large files, use larger chunks
		chunkSize = minInt64(chunkSize*2, 400*1024*1024) // Up to 400MB for huge files
	} else if desc.Size < 500*1024*1024 { // 500 MB
		// For smaller files, use smaller chunks
		chunkSize = maxInt64(chunkSize/2, 5*1024*1024) // At least 5MB
	}

	numChunks := int(math.Ceil(float64(desc.Size) / float64(chunkSize)))

	// Scale concurrency based on file size and available resources
	concurrency := config.chunkConcurrency
	if numChunks < int(concurrency) {
		concurrency = int64(numChunks)
	}

	p.Logf(output.LogLevelDebug, "Downloading layer %s in %d chunks with %d concurrent workers (chunk size: %d MB)",
		desc.Digest, numChunks, concurrency, chunkSize/(1024*1024))

	g, gCtx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(concurrency)

	// This ProxyWriter will be used concurrently to report progress
	pwriter := p.ProxyWriter(io.Discard, desc.Digest.Encoded(), desc.Size, 0)

	// Create a channel to monitor download speeds of the first few chunks
	speedChan := make(chan float64, 5)
	adaptiveConfig := config.adaptiveBufferEnabled && numChunks > 5

	for i := 0; i < numChunks; i++ {
		if err := sem.Acquire(gCtx, 1); err != nil {
			break // Context was cancelled
		}

		chunkIndex := i
		g.Go(func() error {
			defer sem.Release(1)

			start := int64(chunkIndex) * chunkSize
			length := chunkSize
			if start+length > desc.Size {
				length = desc.Size - start
			}

			// Time this chunk download for adaptive configuration
			chunkStart := time.Now()

			// Fetch a new reader for each chunk
			rc, fetchErr := src.Fetch(gCtx, desc)
			if fetchErr != nil {
				return fmt.Errorf("chunk %d: failed to fetch: %w", chunkIndex, fetchErr)
			}
			defer rc.Close()

			seeker, ok := rc.(io.ReadSeeker)
			if !ok {
				return fmt.Errorf("chunk %d: remote does not support seek, cannot download in chunks", chunkIndex)
			}
			if _, seekErr := seeker.Seek(start, io.SeekStart); seekErr != nil {
				return fmt.Errorf("chunk %d: failed to seek to offset %d: %w", chunkIndex, start, seekErr)
			}

			// Use a LimitedReader to ensure we don't read past the chunk boundary
			limitedReader := io.LimitReader(rc, length)

			// Determine buffer size - may be dynamically adjusted based on early chunk performance
			bufSize := config.copyBufferSize
			if adaptiveConfig && chunkIndex > 3 && len(speedChan) > 0 {
				// Calculate average speed from first chunks
				var totalSpeed float64
				var count int
				for len(speedChan) > 0 && count < 3 {
					totalSpeed += <-speedChan
					count++
				}

				if count > 0 {
					avgSpeed := totalSpeed / float64(count)

					// Adjust buffer size based on observed speed
					if avgSpeed > 20*1024*1024 { // 20 MB/s
						// Fast connection - use larger buffers
						bufSize = minInt(bufSize*2, 32*1024*1024) // Up to 32MB
					} else if avgSpeed < 1*1024*1024 { // 1 MB/s
						// Slow connection - use smaller buffers
						bufSize = maxInt(bufSize/2, 32*1024) // At least 32KB
					}
				}
			}

			buf := make([]byte, bufSize)
			n, err := io.CopyBuffer(
				io.MultiWriter(pwriter, &offsetWriter{w: ingestFile, offset: start}),
				limitedReader,
				buf,
			)

			// If this is one of the first few chunks, report its speed for adaptive configuration
			if adaptiveConfig && chunkIndex < 3 && err == nil && n > 0 {
				elapsed := time.Since(chunkStart).Seconds()
				if elapsed > 0 {
					speed := float64(n) / elapsed // bytes per second
					select {
					case speedChan <- speed:
					default:
						// Channel full, just continue
					}
				}
			}

			if err != nil {
				return fmt.Errorf("chunk %d: failed to write to file: %w", chunkIndex, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Verify the integrity of the complete file
	if _, err := ingestFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to start of ingest file for verification: %w", err)
	}
	verifier := desc.Digest.Verifier()
	if _, err := io.Copy(verifier, ingestFile); err != nil {
		return fmt.Errorf("failed to verify downloaded file: %w", err)
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

// Helper functions for min/max and clamping values - fixed to avoid conflicts
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func clampInt64(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Integer versions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
