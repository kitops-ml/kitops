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

package index

import (
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"
)

// ParseLabelSelector reads key=value arguments into a label selector.
func ParseLabelSelector(args []string) (map[string]string, error) {
	selector := map[string]string{}
	for _, arg := range args {
		key, value, isAssignment := strings.Cut(arg, "=")
		if !isAssignment || key == "" {
			return nil, fmt.Errorf("invalid label %q: expected key=value", arg)
		}
		if existing, ok := selector[key]; ok {
			return nil, fmt.Errorf("invalid label %q: %s is already set to %q; each label can be specified only once", arg, key, existing)
		}
		selector[key] = value
	}
	return selector, nil
}

// SelectEntry returns the single entry carrying every label in the selector, and reports what
// the index at indexRef holds when the selector matches none or several of them.
func SelectEntry(idx *ModelKitIndex, selector map[string]string, indexRef *registry.Reference) (ModelKitIndexDescriptor, error) {
	var matches []ModelKitIndexDescriptor
	for _, entry := range idx.Manifests {
		if entryHasLabels(entry, selector) {
			matches = append(matches, entry)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		if len(idx.Manifests) == 0 {
			return ModelKitIndexDescriptor{}, fmt.Errorf("index contains no ModelKits")
		}
		return ModelKitIndexDescriptor{}, fmt.Errorf("no ModelKit in the index has %s. The index contains:\n%s",
			describeSelector(selector), DescribeEntries(idx.Manifests, indexRef))
	default:
		return ModelKitIndexDescriptor{}, fmt.Errorf("%d ModelKits in the index match; add labels to select one of:\n%s",
			len(matches), DescribeEntries(matches, indexRef))
	}
}

func entryHasLabels(entry ModelKitIndexDescriptor, selector map[string]string) bool {
	for key, want := range selector {
		value, found := entry.ModelMeta[key]
		if !found || FormatLabelValue(value) != want {
			return false
		}
	}
	return true
}

// DescribeEntries lists entries by the reference they were added under, falling back to their
// digest, along with their labels.
func DescribeEntries(entries []ModelKitIndexDescriptor, indexRef *registry.Reference) string {
	var lines []string
	for _, entry := range entries {
		labels := "no labels"
		if len(entry.ModelMeta) > 0 {
			labels = strings.Join(LabelPairs(entry.ModelMeta), ", ")
		}
		identity := EntryReference(entry, indexRef)
		if identity == "" {
			identity = entry.Digest.String()
		}
		lines = append(lines, fmt.Sprintf("  %s  (%s)", identity, labels))
	}
	return strings.Join(lines, "\n")
}

func describeSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return "no labels"
	}
	return strings.Join(KeyValuePairs(selector), ", ")
}
