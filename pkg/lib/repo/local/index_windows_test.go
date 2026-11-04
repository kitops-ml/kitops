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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestWriteIndexFileWithOrdinaryWindowsReader(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "index.json")
	original := []byte(`{"original":true}`)
	require.NoError(t, os.WriteFile(path, original, 0600))
	reader, err := os.Open(path)
	require.NoError(t, err)
	defer reader.Close()
	updated := []byte(`{"new":true}`)
	replaceErr := writeIndexFile(path, updated)
	previous, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, original, previous)
	current, err := os.ReadFile(path)
	require.NoError(t, err)
	if replaceErr != nil {
		require.Equal(t, original, current)
	} else {
		require.Equal(t, updated, current)
	}
	entries, err := os.ReadDir(directory)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NoError(t, reader.Close())
	require.NoError(t, writeIndexFile(path, []byte(`{"new":true}`)))
}

func TestWriteIndexFilePreservesDACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"old":true}`), 0600))
	original, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	require.NoError(t, err)
	owner, _, err := original.Owner()
	require.NoError(t, err)
	restricted, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + owner.String() + ")")
	require.NoError(t, err)
	dacl, _, err := restricted.DACL()
	require.NoError(t, err)
	require.NoError(t, windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil))
	before, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	require.NoError(t, err)
	require.NoError(t, writeIndexFile(path, []byte(`{"new":true}`)))
	after, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	require.NoError(t, err)
	require.Equal(t, before.String(), after.String())
}
