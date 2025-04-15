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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/errdef"
)

type localIndex struct {
	indexPath string
	modelTags *tagsIndex
	ocispec.Index
}

func newLocalIndex(storagePath, repoName string) (*localIndex, error) {
	li := &localIndex{}
	indexPath := constants.IndexJsonPathForRepo(storagePath, repoName)
	index, err := parseIndex(indexPath)
	if err != nil {
		return nil, err
	}
	li.indexPath = indexPath
	li.Index = *index

	tagsIndexPath := constants.TagIndexPathForRepo(storagePath, repoName)
	tags, err := parseTagsIndex(tagsIndexPath)
	if err != nil {
		return nil, err
	}
	li.modelTags = tags

	return li, nil
}

func (li *localIndex) addManifest(manifestDesc ocispec.Descriptor) error {
	curTag := manifestDesc.Annotations[ocispec.AnnotationRefName]
	delete(manifestDesc.Annotations, ocispec.AnnotationRefName)
	if !li.exists(manifestDesc) {
		li.Manifests = append(li.Manifests, manifestDesc)
	}
	if err := li.save(); err != nil {
		return err
	}
	if curTag != "" {
		li.modelTags.tagToDigest[curTag] = manifestDesc
		return li.modelTags.save()
	}
	return nil
}

func (li *localIndex) save() error {
	if err := li.modelTags.save(); err != nil {
		return err
	}

	if len(li.Manifests) == 0 {
		// If there are no tags left in this repo, delete the file to clean up
		if err := os.Remove(li.indexPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}

	seenDigests := map[digest.Digest]bool{}
	var uniqueManifests []ocispec.Descriptor
	for _, desc := range li.Manifests {
		if !seenDigests[desc.Digest] {
			uniqueManifests = append(uniqueManifests, desc)
		}
		seenDigests[desc.Digest] = true
	}
	li.Manifests = uniqueManifests

	indexJson, err := json.Marshal(li.Index)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}
	if err := os.WriteFile(li.indexPath, indexJson, 0666); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}
	return nil
}

func (li *localIndex) exists(target ocispec.Descriptor) bool {
	for _, manifestDesc := range li.Manifests {
		if manifestDesc.Digest == target.Digest {
			return true
		}
	}
	return false
}

func (li *localIndex) delete(target ocispec.Descriptor) error {
	tags := li.listTags(target)
	for _, tag := range tags {
		if err := li.untag(tag); err != nil {
			return err
		}
	}

	var newManifests []ocispec.Descriptor
	for _, manifestDesc := range li.Manifests {
		if manifestDesc.Digest != target.Digest {
			newManifests = append(newManifests, manifestDesc)
		}
	}
	li.Manifests = newManifests
	if err := li.save(); err != nil {
		return err
	}
	return nil
}

func (li *localIndex) resolve(reference string) (ocispec.Descriptor, error) {
	if reference == "" {
		return ocispec.DescriptorEmptyJSON, errdef.ErrMissingReference
	}
	if util.ReferenceIsDigest(reference) {
		for _, desc := range li.Manifests {
			if desc.Digest == digest.Digest(reference) {
				return desc, nil
			}
		}
		return ocispec.DescriptorEmptyJSON, errdef.ErrNotFound
	} else {
		return li.modelTags.get(reference)
	}
}

func (li *localIndex) tag(desc ocispec.Descriptor, reference string) error {
	if !li.hasManifest(desc) {
		return fmt.Errorf("%s: %s: %w", desc.Digest, desc.MediaType, errdef.ErrNotFound)
	}
	li.modelTags.tagToDigest[reference] = desc
	return li.modelTags.save()
}

func (li *localIndex) untag(reference string) error {
	if _, err := li.modelTags.get(reference); err != nil {
		return err
	}
	delete(li.modelTags.tagToDigest, reference)
	return li.modelTags.save()
}

func (li *localIndex) listTags(desc ocispec.Descriptor) []string {
	var tags []string
	for tag, manifestDesc := range li.modelTags.tagToDigest {
		if manifestDesc.Digest == desc.Digest {
			tags = append(tags, tag)
		}
	}
	sort.Strings(tags)
	return tags
}

func (li *localIndex) hasManifest(desc ocispec.Descriptor) bool {
	for _, m := range li.Manifests {
		if m.Digest == desc.Digest {
			return true
		}
	}
	return false
}

type tagsIndex struct {
	tagsIndexPath string
	tagToDigest   map[string]ocispec.Descriptor
}

func emptyTagsIndex(tagsIndexPath string) *tagsIndex {
	return &tagsIndex{
		tagsIndexPath: tagsIndexPath,
		tagToDigest:   map[string]ocispec.Descriptor{},
	}
}

func (ti *tagsIndex) get(reference string) (ocispec.Descriptor, error) {
	if desc, exists := ti.tagToDigest[reference]; exists {
		return desc, nil
	}
	return ocispec.DescriptorEmptyJSON, errdef.ErrNotFound
}

func (ti *tagsIndex) save() error {
	if len(ti.tagToDigest) == 0 {
		// If there are no tags left in this repo, delete the file to clean up
		if err := os.Remove(ti.tagsIndexPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}

	jsonBytes, err := json.Marshal(ti.tagToDigest)
	if err != nil {
		return fmt.Errorf("failed to marshal tags index: %w", err)
	}
	if err := os.WriteFile(ti.tagsIndexPath, jsonBytes, 0666); err != nil {
		return fmt.Errorf("failed to save tags index: %w", err)
	}
	return nil
}
