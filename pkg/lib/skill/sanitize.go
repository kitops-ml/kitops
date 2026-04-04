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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"oras.land/oras-go/v2/registry"
)

var (
	nonAlphanumRegex    = regexp.MustCompile(`[^a-z0-9._]+`)
	leadingTrailingTrim = regexp.MustCompile(`^[.\-]+|[.\-]+$`)
)

// SanitizeName normalizes a skill name to lowercase kebab-case suitable for
// use as a directory name. Non-alphanumeric characters (except dots and
// underscores) are replaced with hyphens. Leading/trailing dots and hyphens
// are stripped. Result is truncated to 255 characters. Returns
// "unnamed-skill" if the result would be empty.
func SanitizeName(name string) string {
	s := strings.ToLower(name)
	s = nonAlphanumRegex.ReplaceAllString(s, "-")
	s = leadingTrailingTrim.ReplaceAllString(s, "")
	if len(s) > 255 {
		s = s[:255]
	}
	if s == "" {
		return "unnamed-skill"
	}
	return s
}

// DeriveSkillName determines the skill directory name from a prompt entry
// and optional SKILL.md frontmatter name.
//
// Priority:
//  1. frontmatterName (from SKILL.md frontmatter "name" field)
//  2. prompt.Name (from Kitfile)
//  3. Parent directory name (for SKILL.md files or directory prompts)
//  4. Fallback: repository name from modelRef
func DeriveSkillName(frontmatterName string, prompt artifact.Prompt, modelRef *registry.Reference) string {
	if frontmatterName != "" {
		return SanitizeName(frontmatterName)
	}
	if prompt.Name != "" {
		return SanitizeName(prompt.Name)
	}

	path := prompt.Path
	// Directory prompt (ends with /)
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
		base := filepath.Base(path)
		if base != "." && base != "" {
			return SanitizeName(base)
		}
	} else {
		base := filepath.Base(path)
		// If the file is SKILL.md, use the parent directory name
		if strings.EqualFold(base, "SKILL.md") {
			dir := filepath.Dir(path)
			parent := filepath.Base(dir)
			if parent != "." && parent != "" {
				return SanitizeName(parent)
			}
			// SKILL.md at root — fall through to modelRef
		} else {
			// Use filename without extension
			ext := filepath.Ext(base)
			name := strings.TrimSuffix(base, ext)
			if name != "" {
				return SanitizeName(name)
			}
		}
	}

	// Fallback: repository name from model reference
	if modelRef != nil {
		repo := modelRef.Repository
		// Extract last path component (e.g., "myrepo/my-model" → "my-model")
		if idx := strings.LastIndex(repo, "/"); idx >= 0 {
			repo = repo[idx+1:]
		}
		if repo != "" {
			return SanitizeName(repo)
		}
	}

	return "unnamed-skill"
}
