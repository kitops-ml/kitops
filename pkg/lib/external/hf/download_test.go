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

package hf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/stretchr/testify/require"
)

func TestDownloadFileResumesPartialFile(t *testing.T) {
	var gotRange string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRange = r.Header.Get("Range")
		w.Header().Set("Content-Range", "bytes 3-5/6")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("def"))
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "model.bin")
	require.NoError(t, os.WriteFile(destPath, []byte("abc"), 0600))

	progress, plog := output.NewDownloadProgress()
	err := downloadFile(context.Background(), srv.Client(), "", srv.URL, destPath, "model.bin", 6, progress, plog)
	progress.Done()
	require.NoError(t, err)

	require.Equal(t, "bytes=3-", gotRange)
	contents, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Equal(t, []byte("abcdef"), contents)
}

func TestDownloadFileRestartsWhenRangeIsIgnored(t *testing.T) {
	var gotRange string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRange = r.Header.Get("Range")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "model.bin")
	require.NoError(t, os.WriteFile(destPath, []byte("abc"), 0600))

	progress, plog := output.NewDownloadProgress()
	err := downloadFile(context.Background(), srv.Client(), "", srv.URL, destPath, "model.bin", 6, progress, plog)
	progress.Done()
	require.NoError(t, err)

	require.Equal(t, "bytes=3-", gotRange)
	contents, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Equal(t, []byte("abcdef"), contents)
}

func TestDownloadFileSkipsCompleteFile(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "model.bin")
	require.NoError(t, os.WriteFile(destPath, []byte("abcdef"), 0600))

	progress, plog := output.NewDownloadProgress()
	err := downloadFile(context.Background(), srv.Client(), "", srv.URL, destPath, "model.bin", 6, progress, plog)
	progress.Done()
	require.NoError(t, err)

	require.False(t, called)
	contents, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Equal(t, []byte("abcdef"), contents)
}
