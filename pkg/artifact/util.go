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

package artifact

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2/registry"
)

const (
	DefaultRegistry   = "localhost"
	DefaultRepository = "_"
)

var (
	startEndAlphanumeric = regexp.MustCompile(`[a-z0-9](.*[a-z0-9])?`)
)

// IsModelKitReference returns true if the ref string "looks" like a modelkit reference
func IsModelKitReference(ref string) bool {
	// If it doesn't have ':' or '@' it's probably not a reference
	if !strings.Contains(ref, ":") && !strings.Contains(ref, "@") {
		return false
	}
	// Does it parse?
	if _, _, err := ParseReference(ref); err != nil {
		return false
	}
	return true
}

// ParseReference parses a reference string into a Reference struct. It attempts to make
// references conform to an expected structure, with a defined registry and repository by filling
// default values for registry and repository where appropriate. Where the first part of a reference
// doesn't look like a registry URL, the default registry is used, turning e.g. testorg/testrepo into
// localhost/testorg/testrepo. If refString does not contain a registry or a repository (i.e. is a
// base SHA256 hash), the returned reference uses placeholder values for registry and repository.
//
// See FormatRepositoryForDisplay for removing default values from a registry for displaying to the
// user.
func ParseReference(refString string) (reference *registry.Reference, extraTags []string, err error) {
	// Check if provided input is a plain digest
	if _, err := digest.Parse(refString); err == nil {
		ref := &registry.Reference{
			Registry:   DefaultRegistry,
			Repository: DefaultRepository,
			Reference:  refString,
		}
		return ref, []string{}, nil
	}

	var reg, repo, ref, unprocessed string
	hasDigest := false
	hasTag := false

	// Trim extra tags, if present
	parts := strings.Split(refString, ",")
	unprocessed = parts[0]
	extraTags = parts[1:]

	// Split off registry
	parts = strings.SplitN(unprocessed, "/", 2)
	if len(parts) == 1 {
		// Just a repo, use default registry
		reg = DefaultRegistry
	} else {
		// Check if registry part "looks" like a URL; we're trying to distinguish between cases:
		// a) testorg/testrepo --> should be localhost/testorg/testrepo
		// b) registry.io/testrepo --> should be registry.io/testrepo
		// c) localhost:5000/testrepo --> should be localhost:5000/testrepo
		reg = parts[0]
		if !strings.Contains(reg, ":") && !strings.Contains(reg, ".") {
			reg = DefaultRegistry
		} else {
			unprocessed = parts[1]
		}
	}

	// Split tags/digest from repository
	if index := strings.Index(unprocessed, "@"); index != -1 {
		hasDigest = true
		repo = unprocessed[:index]
		ref = unprocessed[index+1:]
		if index := strings.Index(repo, ":"); index != -1 {
			repo = repo[:index]
		}
	} else if index := strings.Index(unprocessed, ":"); index != -1 {
		hasTag = true
		repo = unprocessed[:index]
		ref = unprocessed[index+1:]
	} else {
		// No tag or digest
		repo = unprocessed
	}

	// Check for common errors
	if strings.ToLower(repo) != repo {
		return nil, nil, fmt.Errorf("repository (%s) name must be lowercase", repo)
	}
	if !startEndAlphanumeric.MatchString(repo) {
		return nil, nil, fmt.Errorf("repository (%s) must start and end with a letter or number", repo)
	}

	reference = &registry.Reference{
		Registry:   reg,
		Repository: repo,
		Reference:  ref,
	}
	// Do full checks in case we missed something
	if err := reference.ValidateRegistry(); err != nil {
		return nil, nil, err
	}
	if err := reference.ValidateRepository(); err != nil {
		return nil, nil, err
	}
	if hasTag {
		if err := reference.ValidateReferenceAsTag(); err != nil {
			return nil, nil, err
		}
	} else if hasDigest {
		if err := reference.ValidateReferenceAsDigest(); err != nil {
			return nil, nil, err
		}
	}

	return reference, extraTags, nil
}

// ReferenceIsDigest returns if the reference is a digest. If false, reference should
// be treated as a tag
func ReferenceIsDigest(ref string) bool {
	err := digest.Digest(ref).Validate()
	return err == nil
}

// DefaultReference returns a reference that can be used when no reference is supplied. It uses
// the default registry and repository
func DefaultReference() *registry.Reference {
	return &registry.Reference{
		Registry:   DefaultRegistry,
		Repository: DefaultRepository,
	}
}

// FormatRepositoryForDisplay removes default values from a repository string to avoid surfacing defaulted fields
// when displaying references, which may be confusing.
func FormatRepositoryForDisplay(repo string) string {
	// Trim default registry, if present
	repo = strings.TrimPrefix(repo, DefaultRegistry+"/")
	// Trim default repository, if present
	repo = strings.TrimPrefix(repo, DefaultRepository)
	// Trim @ in case what's left is a bare digest
	repo = strings.TrimPrefix(repo, "@")
	return repo
}
