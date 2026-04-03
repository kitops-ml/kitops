// Copyright 2025 The KitOps Authors.
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

package hf

import (
	"fmt"
	"net/url"
	"strings"
)

// RepositoryType represents the kind of Hugging Face repository.
type RepositoryType int

const (
	RepoTypeUnknown RepositoryType = iota
	RepoTypeModel
	RepoTypeDataset
)

// ParseHuggingFaceRepo parses a Hugging Face repository URL or path and returns
// the normalized repository path (org/repo) and the repository type.
func ParseHuggingFaceRepo(rawURL string) (string, RepositoryType, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", RepoTypeUnknown, fmt.Errorf("failed to parse url: %w", err)
	}

	// Strict host validation - must be exactly "huggingface.co" or a proper subdomain
	if u.Host != "" {
		if u.Host != "huggingface.co" && !strings.HasSuffix(u.Host, ".huggingface.co") {
			return "", RepoTypeUnknown, fmt.Errorf("not a Hugging Face repository")
		}
	}

	path := strings.Trim(u.Path, "/")
	segments := strings.Split(path, "/")

	// Handle scheme-less URLs like "huggingface.co/datasets/org/repo"
	// When there's no scheme, url.Parse puts everything in the path
	if len(segments) > 0 && (segments[0] == "huggingface.co" || strings.HasSuffix(segments[0], ".huggingface.co")) {
		segments = segments[1:] // Skip the host part from path
	}

	if len(segments) >= 3 && segments[0] == "datasets" {
		return strings.Join(segments[1:3], "/"), RepoTypeDataset, nil
	}
	if len(segments) == 2 && segments[0] == "datasets" {
		return "", RepoTypeUnknown, fmt.Errorf("could not extract repository from path '%s'", path)
	}

	if len(segments) >= 2 {
		return strings.Join(segments[len(segments)-2:], "/"), RepoTypeModel, nil
	}

	return "", RepoTypeUnknown, fmt.Errorf("could not extract repository from path '%s'", path)
}
