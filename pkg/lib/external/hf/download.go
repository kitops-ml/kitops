// Copyright 2025 The KitOps Authors.
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

package hf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/output"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	modelResolveURLFmt   = "https://huggingface.co/%s/resolve/%s/%s"
	datasetResolveURLFmt = "https://huggingface.co/datasets/%s/resolve/%s/%s"
)

// DownloadFiles fetches each entry of files into destDir. The caller is
// expected to pass an immutable commit SHA as repoRef (via PinCommit) so the
// snapshot is consistent across all downloads.
func DownloadFiles(
	ctx context.Context,
	modelRepo, repoRef, destDir string,
	files []kfgen.FileListing,
	token string,
	maxConcurrency int,
	repoType RepositoryType) error {

	if len(files) == 0 {
		return nil
	}

	// Reject any path that would escape destDir before doing any network or
	// filesystem work. HF tree responses are not authoritative; a malicious
	// or compromised repo could embed "../" segments to write files outside
	// the import temp directory.
	for _, f := range files {
		if _, _, err := filesystem.VerifySubpath(destDir, f.Path); err != nil {
			return fmt.Errorf("rejecting unsafe path %q from HuggingFace API: %w", f.Path, err)
		}
	}

	var resolveURLFmt string
	if repoType == RepoTypeDataset {
		resolveURLFmt = datasetResolveURLFmt
	} else {
		resolveURLFmt = modelResolveURLFmt
	}

	client := &http.Client{
		Timeout: 1 * time.Hour,
	}
	sem := semaphore.NewWeighted(int64(maxConcurrency))
	errs, errCtx := errgroup.WithContext(ctx)
	var semErr error

	progress, plog := output.NewDownloadProgress()

	for _, f := range files {
		if err := sem.Acquire(errCtx, 1); err != nil {
			semErr = err
			break
		}

		fileURL := fmt.Sprintf(resolveURLFmt, modelRepo, repoRef, f.Path)
		destPath := filepath.Join(destDir, f.Path)
		errs.Go(func() error {
			defer sem.Release(1)
			plog.Infof("Downloading file %s", f.Path)
			return downloadFile(errCtx, client, token, fileURL, destPath, f.Path, f.Size, progress, plog)
		})
	}

	if err := errs.Wait(); err != nil {
		return err
	}
	if semErr != nil {
		return fmt.Errorf("failed to acquire lock: %w", semErr)
	}
	progress.Done()

	return nil
}

func downloadFile(
	ctx context.Context,
	client *http.Client,
	token, srcURL, destPath, filename string,
	size int64,
	progress *output.DownloadProgressBar,
	plog *output.ProgressLogger) error {

	plog.Debugf("Downloading from %s", srcURL)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	resumeOffset, complete, err := downloadResumeOffset(destPath, size)
	if err != nil {
		return err
	}
	if complete {
		plog.Debugf("Using cached file %s", destPath)
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return fmt.Errorf("failed to resolve URL: %w", err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			plog.Logf(output.LogLevelWarn, "Failed to close response body: %s", err)
		}
	}()

	if resumeOffset > 0 && resp.StatusCode == http.StatusOK {
		plog.Debugf("Server ignored range request for %s; restarting download", filename)
		resumeOffset = 0
	}
	if resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent {
		expectedPrefix := fmt.Sprintf("bytes %d-", resumeOffset)
		if contentRange := resp.Header.Get("Content-Range"); contentRange != "" && !strings.HasPrefix(contentRange, expectedPrefix) {
			return fmt.Errorf("unexpected content range %q when resuming file %s", contentRange, filename)
		}
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status code %d when downloading file %s from %s", resp.StatusCode, filename, srcURL)
	}

	contentRC := progress.TrackDownload(resp.Body, filename, remainingDownloadSize(size, resumeOffset, resp.ContentLength))
	defer func() {
		if err := contentRC.Close(); err != nil {
			plog.Logf(output.LogLevelWarn, "Failed to close download stream for %s: %s", filename, err)
		}
	}()

	flags := os.O_CREATE | os.O_WRONLY
	if resumeOffset > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(destPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			plog.Logf(output.LogLevelError, "Error closing file %s: %s", destPath, err)
		}
	}()

	n, err := io.Copy(f, contentRC)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if resp.ContentLength > 0 && n != resp.ContentLength {
		return fmt.Errorf("mismatched file size: expected %d but got %d", resp.ContentLength, n)
	}
	if size > 0 {
		finalSize := n
		if resumeOffset > 0 {
			finalSize += resumeOffset
		}
		if finalSize != size {
			return fmt.Errorf("mismatched file size: expected %d but got %d", size, finalSize)
		}
	}

	return nil
}

func downloadResumeOffset(destPath string, size int64) (offset int64, complete bool, err error) {
	info, err := os.Stat(destPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to examine existing file: %w", err)
	}
	if info.IsDir() {
		return 0, false, fmt.Errorf("download destination %s is a directory", destPath)
	}

	existingSize := info.Size()
	if size == 0 {
		if existingSize == 0 {
			return 0, true, nil
		}
		if err := os.Remove(destPath); err != nil {
			return 0, false, fmt.Errorf("failed to remove oversized cached file: %w", err)
		}
		return 0, false, nil
	}
	if existingSize == size {
		return 0, true, nil
	}
	if existingSize > size {
		if err := os.Remove(destPath); err != nil {
			return 0, false, fmt.Errorf("failed to remove oversized cached file: %w", err)
		}
		return 0, false, nil
	}
	return existingSize, false, nil
}

func remainingDownloadSize(size, resumeOffset, contentLength int64) int64 {
	if size > 0 && resumeOffset > 0 {
		return size - resumeOffset
	}
	if size > 0 {
		return size
	}
	if contentLength > 0 {
		return contentLength
	}
	return 0
}
