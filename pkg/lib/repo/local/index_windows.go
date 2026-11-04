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
	"os"

	"golang.org/x/sys/windows"
)

func openIndex(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

func copyIndexSecurity(source, target *os.File) error {
	information := windows.SECURITY_INFORMATION(windows.OWNER_SECURITY_INFORMATION | windows.GROUP_SECURITY_INFORMATION | windows.DACL_SECURITY_INFORMATION)
	security, err := windows.GetSecurityInfo(windows.Handle(source.Fd()), windows.SE_FILE_OBJECT, information)
	if err != nil {
		return err
	}
	owner, _, err := security.Owner()
	if err != nil {
		return err
	}
	group, _, err := security.Group()
	if err != nil {
		return err
	}
	dacl, _, err := security.DACL()
	if err != nil {
		return err
	}
	control, _, err := security.Control()
	if err != nil {
		return err
	}
	if control&windows.SE_DACL_PROTECTED != 0 {
		information |= windows.PROTECTED_DACL_SECURITY_INFORMATION
	} else {
		information |= windows.UNPROTECTED_DACL_SECURITY_INFORMATION
	}
	return windows.SetNamedSecurityInfo(target.Name(), windows.SE_FILE_OBJECT, information, owner, group, dacl, nil)
}
