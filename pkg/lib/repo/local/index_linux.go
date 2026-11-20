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
	"bytes"
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func copyIndexACL(source, target *os.File) error {
	for _, name := range []string{"security.selinux", "system.posix_acl_access"} {
		size, err := unix.Fgetxattr(int(source.Fd()), name, nil)
		if errors.Is(err, unix.ENOTSUP) {
			continue
		}
		if errors.Is(err, unix.ENODATA) {
			_, err = unix.Fgetxattr(int(target.Fd()), name, nil)
			if errors.Is(err, unix.ENODATA) || errors.Is(err, unix.ENOTSUP) {
				continue
			}
			if err != nil {
				return err
			}
			err = unix.Fremovexattr(int(target.Fd()), name)
			if err != nil && !errors.Is(err, unix.ENODATA) && !errors.Is(err, unix.ENOTSUP) {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		value := make([]byte, size)
		size, err = unix.Fgetxattr(int(source.Fd()), name, value)
		if err != nil {
			return err
		}
		current := make([]byte, size)
		if n, err := unix.Fgetxattr(int(target.Fd()), name, current); err == nil && bytes.Equal(value[:size], current[:n]) {
			continue
		}
		if err := unix.Fsetxattr(int(target.Fd()), name, value[:size], 0); err != nil {
			return err
		}
	}
	return nil
}

func createPrivateIndexFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
}
