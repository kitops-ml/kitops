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
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Darwin returns the ACL as an opaque attribute block, including its relative offsets.
func copyIndexACL(source, target *os.File) error {
	attributes := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	// Darwin limits ACLs to 128 entries; this also leaves room for the attribute header.
	buffer := make([]byte, 8192)
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, source.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(source)
	if errno != 0 {
		return errno
	}
	length := binary.NativeEndian.Uint32(buffer)
	if length < 12 || length > uint32(len(buffer)) {
		return fmt.Errorf("invalid index security attributes length: %d", length)
	}
	// A missing ACL needs an explicit filesec header to clear inherited entries.
	// Layout and sentinel values are defined in sys/attr.h and sys/kauth.h.
	emptyACL := struct {
		Offset            int32
		Length            uint32
		Magic             uint32
		Owner, Group      [16]byte
		EntryCount, Flags uint32
	}{Offset: 8, Length: 44, Magic: 0x012cc16d, EntryCount: ^uint32(0)}
	security := buffer[4:length]
	if binary.NativeEndian.Uint32(buffer[8:12]) == 0 {
		security = unsafe.Slice((*byte)(unsafe.Pointer(&emptyACL)), int(unsafe.Sizeof(emptyACL)))
	}
	// The setter takes the attribute block without the getter's leading length word.
	_, _, errno = unix.Syscall6(unix.SYS_FSETATTRLIST, target.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&security[0])), uintptr(len(security)), 0, 0)
	runtime.KeepAlive(target)
	if errno != 0 {
		return errno
	}
	return nil
}

func createPrivateIndexFile(path string) (*os.File, error) {
	name, err := unix.BytePtrFromString(path)
	if err != nil {
		return nil, err
	}
	// Supply an empty, protected ACL at creation to prevent inherited read access.
	// Layout and flag values are defined in sys/kauth.h.
	security := struct {
		Magic             uint32
		Owner, Group      [16]byte
		EntryCount, Flags uint32
	}{Magic: 0x012cc16d, Flags: 1 << 17}
	const noOwnerOrGroup = ^uint32(0) - 100 // KAUTH_UID_NONE and KAUTH_GID_NONE.
	flags := unix.O_WRONLY | unix.O_CREAT | unix.O_EXCL | unix.O_CLOEXEC
	fd, _, errno := unix.Syscall6(unix.SYS_OPEN_EXTENDED, uintptr(unsafe.Pointer(name)), uintptr(flags), uintptr(noOwnerOrGroup), uintptr(noOwnerOrGroup), 0600, uintptr(unsafe.Pointer(&security)))
	if errno != 0 {
		return nil, &os.PathError{Op: "open", Path: path, Err: errno}
	}
	return os.NewFile(fd, path), nil
}
