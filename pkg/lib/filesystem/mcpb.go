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

package filesystem

import (
	"archive/zip"
	"fmt"
	"os"
)

// ValidateMCPB verifies that the file at path is a well-formed MCP Bundle: a
// regular file containing a ZIP archive with a manifest.json at its root.
func ValidateMCPB(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to read MCP bundle %s: %w", path, err)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("MCP bundle path %s must be a single .mcpb file", path)
	}
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("MCP bundle %s is not a valid ZIP archive: %w", path, err)
	}
	defer zipReader.Close()
	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			return nil
		}
	}
	return fmt.Errorf("MCP bundle %s does not contain a manifest.json at its root", path)
}
