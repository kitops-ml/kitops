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

package util

import (
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

func LayerPathsFromKitfile(kitfile *artifact.KitFile) []string {
	cleanPath := func(path string) string {
		return filepath.Clean(strings.TrimSpace(path))
	}
	var layerPaths []string
	for _, code := range kitfile.Code {
		layerPaths = append(layerPaths, cleanPath(code.Path))
	}
	for _, dataset := range kitfile.DataSets {
		layerPaths = append(layerPaths, cleanPath(dataset.Path))
	}
	for _, docs := range kitfile.Docs {
		layerPaths = append(layerPaths, cleanPath(docs.Path))
	}
	for _, prompt := range kitfile.Prompts {
		layerPaths = append(layerPaths, cleanPath(prompt.Path))
	}

	if kitfile.Model != nil {
		if kitfile.Model.Path != "" {
			layerPaths = append(layerPaths, cleanPath(kitfile.Model.Path))
		}
		for _, part := range kitfile.Model.Parts {
			layerPaths = append(layerPaths, cleanPath(part.Path))
		}
	}
	return layerPaths
}
