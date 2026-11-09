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
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestWriteIndexFilePreservesPOSIXACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"old":true}`), 0600))
	acl := binary.LittleEndian.AppendUint32(nil, 2)
	for _, entry := range []struct {
		tag, permissions uint16
		id               uint32
	}{
		{1, 6, ^uint32(0)}, {2, 4, 65534}, {4, 0, ^uint32(0)}, {16, 4, ^uint32(0)}, {32, 0, ^uint32(0)},
	} {
		acl = binary.LittleEndian.AppendUint16(acl, entry.tag)
		acl = binary.LittleEndian.AppendUint16(acl, entry.permissions)
		acl = binary.LittleEndian.AppendUint32(acl, entry.id)
	}
	err := unix.Setxattr(path, "system.posix_acl_access", acl, 0)
	if errors.Is(err, unix.ENOTSUP) {
		t.Skip("filesystem does not support POSIX ACLs")
	}
	require.NoError(t, err)
	require.NoError(t, writeIndexFile(path, []byte(`{"new":true}`)))
	got := make([]byte, len(acl))
	n, err := unix.Getxattr(path, "system.posix_acl_access", got)
	require.NoError(t, err)
	require.Equal(t, acl, got[:n])
}
