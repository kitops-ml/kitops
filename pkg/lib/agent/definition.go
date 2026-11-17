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

package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

const (
	// DefinitionFileName is the conventional name of a standalone Agent
	// Definition, packed in the code layer.
	DefinitionFileName = artifact.AgentDefinitionFileName
	// HiddenDefinitionFileName is the dot-prefixed spelling of the same file,
	// for projects that keep tool configuration out of the directory listing.
	HiddenDefinitionFileName = artifact.HiddenAgentDefinitionFileName
)

// definitionFileNames are the paths, relative to the context directory, at
// which kit init looks for a standalone Agent Definition.
var definitionFileNames = []string{DefinitionFileName, HiddenDefinitionFileName}

// existingDefinition looks for a standalone Agent Definition at the paths in
// definitionFileNames. It returns the path relative to contextDir and the
// definition's framework, or an empty path when there is none. A file that
// cannot be parsed or is not a regular file is an error, as is a definition at
// both paths, since a ModelKit carries only one.
func existingDefinition(contextDir string) (path, framework string, err error) {
	var found []string
	for _, name := range definitionFileNames {
		isDefinition, nameFramework, err := readDefinition(contextDir, name)
		if err != nil {
			return "", "", err
		}
		if isDefinition {
			found = append(found, name)
			path, framework = name, nameFramework
		}
	}
	if len(found) > 1 {
		return "", "", fmt.Errorf("both %s hold an agent definition; a ModelKit may carry only one, so remove one of them",
			strings.Join(found, " and "))
	}
	return path, framework, nil
}

// readDefinition reports whether the file name in contextDir holds an
// AgentDefinition document, and that document's framework default.
func readDefinition(contextDir, name string) (isDefinition bool, framework string, err error) {
	full := filepath.Join(contextDir, name)
	info, err := os.Lstat(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to read %s: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		return false, "", fmt.Errorf("%s is not a regular file; kit packs regular files only, so the definition would not travel in the ModelKit", name)
	}

	file, err := os.Open(full)
	if err != nil {
		return false, "", fmt.Errorf("failed to read %s: %w", name, err)
	}
	defer file.Close()

	isDefinition, framework, err = artifact.DetectAgentDefinition(file)
	if err != nil {
		return false, "", fmt.Errorf("failed to parse %s: %w", name, err)
	}
	return isDefinition, framework, nil
}
