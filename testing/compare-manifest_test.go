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

package testing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/kitops-ml/kitops/pkg/cmd/diff"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type compareTestCase struct {
	name             string
	manifestAPath    string
	manifestBPath    string
	expectedDiffPath string
}

func loadManifest(t *testing.T, filename string) *ocispec.Manifest {
	t.Helper()

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read manifest file %q: %v", filename, err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("failed to unmarshal manifest file %q: %v", filename, err)
	}
	return &manifest
}

func loadDiffResult(t *testing.T, filename string) diff.DiffResult {
	t.Helper()

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read diff result file %q: %v", filename, err)
	}

	var dr diff.DiffResult
	if err := json.Unmarshal(data, &dr); err != nil {
		t.Fatalf("failed to unmarshal diff result file %q: %v", filename, err)
	}
	return dr
}

func TestCompareManifests(t *testing.T) {
	// Define a slice of test cases.
	testCases := []compareTestCase{
		{
			name:             "DifferentConfig",
			manifestAPath:    "manifestA.json",
			manifestBPath:    "manifestB.json",
			expectedDiffPath: "different-config.json",
		},
		{
			name:             "SameConfigOneDifferentLayer",
			manifestAPath:    "manifestA.json",
			manifestBPath:    "manifestC.json",
			expectedDiffPath: "different-layer.json",
		},
		{
			name:             "AnnotationsMismatch",
			manifestAPath:    "manifestA.json",
			manifestBPath:    "manifestD.json",
			expectedDiffPath: "annotation-mismatch.json",
		},
		{
			name:             "MixedDifference",
			manifestAPath:    "manifestA.json",
			manifestBPath:    "manifestE.json",
			expectedDiffPath: "mixed-diff.json",
		},
	}

	// Iterate over the test cases.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifestA := loadManifest(t, filepath.Join("testdata", "compare-manifest", tc.manifestAPath))
			manifestB := loadManifest(t, filepath.Join("testdata", "compare-manifest", tc.manifestBPath))
			expected := loadDiffResult(t, filepath.Join("testdata", "compare-manifest", tc.expectedDiffPath))

			result := diff.CompareManifests(manifestA, manifestB)

			if err := compareDiffResults(&expected, result); err != nil {
				t.Errorf("Test %s failed: %v", tc.name, err)
			}
		})
	}
}

func compareDescriptors(a, b []ocispec.Descriptor) bool {
	if len(a) != len(b) {
		return false
	}

	aSorted := make([]ocispec.Descriptor, len(a))
	bSorted := make([]ocispec.Descriptor, len(b))
	copy(aSorted, a)
	copy(bSorted, b)

	sort.Slice(aSorted, func(i, j int) bool {
		return aSorted[i].Digest < aSorted[j].Digest
	})
	sort.Slice(bSorted, func(i, j int) bool {
		return bSorted[i].Digest < bSorted[j].Digest
	})

	for i := range aSorted {
		if aSorted[i].Digest != bSorted[i].Digest {
			return false
		}
	}
	return true
}

func compareDiffResults(expected, received *diff.DiffResult) error {
	if expected.SameConfig != received.SameConfig {
		return fmt.Errorf("SameConfig mismatch: expected %v, got %v", expected.SameConfig, received.SameConfig)
	}
	if expected.AnnotationsMatch != received.AnnotationsMatch {
		return fmt.Errorf("AnnotationsMatch mismatch: expected %v, got %v", expected.AnnotationsMatch, received.AnnotationsMatch)
	}
	if !compareDescriptors(expected.SharedLayers, received.SharedLayers) {
		return fmt.Errorf("SharedLayers mismatch:\nexpected: %v\ngot: %v", extractDigests(expected.SharedLayers), extractDigests(received.SharedLayers))
	}
	if !compareDescriptors(expected.UniqueLayersA, received.UniqueLayersA) {
		return fmt.Errorf("UniqueLayersA mismatch:\nexpected: %v\ngot: %v", extractDigests(expected.UniqueLayersA), extractDigests(received.UniqueLayersA))
	}
	if !compareDescriptors(expected.UniqueLayersB, received.UniqueLayersB) {
		return fmt.Errorf("UniqueLayersB mismatch:\nexpected: %v\ngot: %v", extractDigests(expected.UniqueLayersB), extractDigests(received.UniqueLayersB))
	}
	return nil
}

// extractDigests is a helper that returns a slice of digests from a slice of descriptors.
func extractDigests(descs []ocispec.Descriptor) []string {
	digests := make([]string, len(descs))
	for i, d := range descs {
		digests[i] = string(d.Digest)
	}
	return digests
}
