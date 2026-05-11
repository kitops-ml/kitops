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

package mlflow

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

// DownloadArtifacts downloads the given artifact files from the MLFlow tracking
// server into destDir, preserving the relative path structure returned by the
// artifacts list API.  maxConcurrency caps parallel downloads.
func DownloadArtifacts(
	ctx context.Context,
	client *Client,
	runID, destDir string,
	files []ArtifactFileInfo,
	maxConcurrency int,
) error {
	httpClient := &http.Client{
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

		artifactPath := f.Path
		localPath := filepath.FromSlash(artifactPath)
		if _, _, err := filesystem.VerifySubpath(destDir, localPath); err != nil {
			return fmt.Errorf("unsafe artifact path %q: %w", artifactPath, err)
		}
		destPath := filepath.Join(destDir, localPath)
		downloadURL, err := client.DownloadArtifactURL(runID, artifactPath)
		if err != nil {
			return fmt.Errorf("failed to build download URL for %s: %w", artifactPath, err)
		}
		fileSize := f.FileSize

		errs.Go(func() error {
			defer sem.Release(1)
			plog.Infof("Downloading artifact %s", artifactPath)
			return downloadArtifact(errCtx, httpClient, client.token, downloadURL, destPath, artifactPath, fileSize, progress, plog)
		})
	}

	if err := errs.Wait(); err != nil {
		return err
	}
	if semErr != nil {
		return fmt.Errorf("failed to acquire download semaphore: %w", semErr)
	}
	progress.Done()

	return nil
}

// ArtifactsToDirectoryListing converts a flat slice of ArtifactFileInfo into the
// DirectoryListing tree structure expected by the kitfile generator.
func ArtifactsToDirectoryListing(files []ArtifactFileInfo) *kfgen.DirectoryListing {
	root := &kfgen.DirectoryListing{
		Name: ".",
		Path: ".",
	}
	for _, f := range files {
		addToListing(root, f)
	}
	return root
}

func addToListing(root *kfgen.DirectoryListing, f ArtifactFileInfo) {
	parts := splitPath(f.Path)
	if len(parts) == 0 {
		return
	}

	cur := root
	for i, part := range parts {
		isLast := i == len(parts)-1
		if isLast {
			cur.Files = append(cur.Files, kfgen.FileListing{
				Name: part,
				Path: f.Path,
				Size: f.FileSize,
			})
		} else {
			cur = findOrCreateSubdir(cur, part, joinPath(parts[:i+1]))
		}
	}
}

func findOrCreateSubdir(parent *kfgen.DirectoryListing, name, path string) *kfgen.DirectoryListing {
	for i := range parent.Subdirs {
		if parent.Subdirs[i].Name == name {
			return &parent.Subdirs[i]
		}
	}
	parent.Subdirs = append(parent.Subdirs, kfgen.DirectoryListing{
		Name: name,
		Path: path,
	})
	return &parent.Subdirs[len(parent.Subdirs)-1]
}

// splitPath splits a slash-separated artifact path into its components.
func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	var parts []string
	for _, seg := range strings.Split(p, "/") {
		if seg != "" {
			parts = append(parts, seg)
		}
	}
	return parts
}

func joinPath(parts []string) string {
	return strings.Join(parts, "/")
}

func downloadArtifact(
	ctx context.Context,
	client *http.Client,
	token, srcURL, destPath, filename string,
	size int64,
	progress *output.DownloadProgressBar,
	plog *output.ProgressLogger,
) error {
	plog.Debugf("Downloading artifact from %s", srcURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build download request for %s: %w", filename, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading artifact %s: %w", filename, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			plog.Logf(output.LogLevelWarn, "Failed to close response body for %s: %s", filename, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %d when downloading artifact %s", resp.StatusCode, filename)
	}

	contentRC := progress.TrackDownload(resp.Body, filename, size)
	defer func() {
		if err := contentRC.Close(); err != nil {
			plog.Logf(output.LogLevelWarn, "Error closing progress tracker for %s: %s", filename, err)
		}
	}()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filename, err)
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", destPath, err)
	}
	defer func() {
		if err := f.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			plog.Logf(output.LogLevelError, "Error closing file %s: %s", destPath, err)
		}
	}()

	if _, err := io.Copy(f, contentRC); err != nil {
		return fmt.Errorf("failed to write artifact %s: %w", filename, err)
	}
	return nil
}
