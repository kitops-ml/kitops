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

package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSinkCommitWritesTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")

	sink, err := OpenDataSink(target)
	require.NoError(t, err)
	defer sink.Close()

	_, err = sink.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, sink.Commit())

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(got))
}

func TestFileSinkAbortPreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")
	require.NoError(t, os.WriteFile(target, []byte("original"), 0o644))

	sink, err := OpenDataSink(target)
	require.NoError(t, err)
	// Write something, then close without Commit (simulating importer failure
	// after open / before commit).
	_, err = sink.Write([]byte("would-be-new"))
	require.NoError(t, err)
	require.NoError(t, sink.Close())

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "original", string(got), "Close without Commit must leave the target untouched")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "temp file must be cleaned up; got %v", entries)
	assert.Equal(t, "out.json", entries[0].Name())
}

func TestFileSinkCommitIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")

	sink, err := OpenDataSink(target)
	require.NoError(t, err)
	defer sink.Close()

	_, err = sink.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, sink.Commit())
	require.NoError(t, sink.Commit())
	require.NoError(t, sink.Close())
}
