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

//go:build linux || darwin

package local

import (
	"os"

	"golang.org/x/sys/unix"
)

func openIndex(path string) (*os.File, error) { return os.Open(path) }

func copyIndexSecurity(source, target *os.File) error {
	var original, temporary unix.Stat_t
	if err := unix.Fstat(int(source.Fd()), &original); err != nil {
		return err
	}
	if err := unix.Fstat(int(target.Fd()), &temporary); err != nil {
		return err
	}
	if original.Uid != temporary.Uid || original.Gid != temporary.Gid {
		if err := target.Chown(int(original.Uid), int(original.Gid)); err != nil {
			return err
		}
	}
	if err := target.Chmod(os.FileMode(original.Mode & 0777)); err != nil {
		return err
	}
	return copyIndexACL(source, target)
}

func replaceIndexFile(temporaryPath, path string) error {
	return os.Rename(temporaryPath, path)
}
