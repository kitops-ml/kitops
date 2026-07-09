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

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"

	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content/memory"
)

// TestSaveModelToNonLocalTarget verifies that SaveModel can write directly to an arbitrary
// oras.Target (such as a remote registry) without requiring a local.LocalRepo, which is what
// allows `kit pack --push` to stream a modelkit straight to a registry without persisting it
// to local storage first.
func TestSaveModelToNonLocalTarget(t *testing.T) {
	tmpDir := t.TempDir()
	codeDir := filepath.Join(tmpDir, "code")
	require.NoError(t, os.Mkdir(codeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(codeDir, "main.py"), []byte("print('hi')"), 0644))

	curDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(curDir) }()
	require.NoError(t, os.Chdir(tmpDir))

	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Code:            []artifact.Code{{Path: "code"}},
	}
	ignorePaths, err := ignore.NewFromContext(tmpDir, kitfile)
	require.NoError(t, err)

	target := memory.New()
	ctx := context.Background()
	manifestDesc, err := SaveModel(ctx, target, kitfile, ignorePaths, &SaveModelOptions{
		ModelFormat: mediatype.KitFormat,
		Compression: mediatype.NoneCompression,
		LayerFormat: mediatype.TarFormat,
	})
	require.NoError(t, err)

	exists, err := target.Exists(ctx, *manifestDesc)
	require.NoError(t, err)
	require.True(t, exists, "manifest should be pushed directly to the target")
}
