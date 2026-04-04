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

package skill

import (
	"archive/tar"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	// MaxSkillEntrySize is the maximum size of a single tar entry in a skill layer.
	MaxSkillEntrySize = 10 * 1024 * 1024 // 10 MB
	// MaxSkillLayerSize is the maximum total size of all entries in a skill layer.
	MaxSkillLayerSize = 10 * 1024 * 1024 // 10 MB
)

// TarEntry holds a single file from a tar archive.
type TarEntry struct {
	Header  *tar.Header
	Content []byte
}

// ReadSkillLayer reads all entries from a tar reader, buffers them, and
// determines whether the layer contains a SKILL.md file. Returns the buffered
// entries, whether it qualifies as a skill, and the frontmatter "name" if
// present.
//
// Enforces that there are no symlinks/hardlinks and the size limits.
// The tar.Reader is fully consumed by this call.
func ReadSkillLayer(tr *tar.Reader) (entries []TarEntry, isSkill bool, frontmatterName string, err error) {
	var totalSize int64
	var skillContent []byte

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, "", fmt.Errorf("reading tar entry: %w", err)
		}

		// reject symlinks and hard links
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return nil, false, "", fmt.Errorf("unsupported entry type in prompt layer: %s (symlinks and hard links are not allowed)", header.Name)
		}

		// check individual entry size
		if header.Size > MaxSkillEntrySize {
			return nil, false, "", fmt.Errorf("prompt layer entry %q exceeds maximum size (%d bytes)", header.Name, MaxSkillEntrySize)
		}

		var content []byte
		if header.Typeflag == tar.TypeReg {
			// enforce size limit with LimitReader
			totalSize += header.Size
			if totalSize > MaxSkillLayerSize {
				return nil, false, "", fmt.Errorf("prompt layer exceeds maximum total size (%d bytes)", MaxSkillLayerSize)
			}
			lr := io.LimitReader(tr, header.Size+1) // +1 to detect lying headers
			content, err = io.ReadAll(lr)
			if err != nil {
				return nil, false, "", fmt.Errorf("reading tar entry %q: %w", header.Name, err)
			}
			if int64(len(content)) > header.Size {
				content = content[:header.Size]
			}
		}

		entries = append(entries, TarEntry{
			Header:  header,
			Content: content,
		})

		// Check if this entry is a SKILL.md
		if isSkillMd(header.Name) && header.Typeflag == tar.TypeReg {
			skillContent = content
		}
	}

	if skillContent != nil {
		isSkill = true
		if fm := ParseSkillFrontmatter(skillContent); fm != nil {
			frontmatterName = fm.Name
		}
	}

	return entries, isSkill, frontmatterName, nil
}

// isSkillMd checks if the tar entry name is a SKILL.md file at any depth.
// Each prompt layer tar represents a single Kitfile entry, so any SKILL.md
// within it is the skill definition for that layer.
func isSkillMd(name string) bool {
	return filepath.Base(filepath.Clean(name)) == "SKILL.md"
}

// SkillFrontmatter represents the YAML frontmatter in a SKILL.md file.
type SkillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	License     string `yaml:"license,omitempty"`
}

// ParseSkillFrontmatter extracts and parses YAML frontmatter from SKILL.md
// content. Returns nil if no valid frontmatter is found.
func ParseSkillFrontmatter(data []byte) *SkillFrontmatter {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return nil
	}

	fmYAML := strings.Join(lines[1:end], "\n")
	if strings.TrimSpace(fmYAML) == "" {
		return nil
	}

	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(fmYAML), &fm); err != nil {
		return nil
	}
	return &fm
}

