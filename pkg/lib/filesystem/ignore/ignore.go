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

package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kfutils "github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/moby/patternmatcher"
	"github.com/moby/patternmatcher/ignorefile"
)

type Paths interface {
	Matches(path, layerPath string) (bool, error)
	HasExclusions() bool
}

func NewFromContext(contextDir string, kitfile *artifact.KitFile, extraLayers ...string) (Paths, error) {
	kitIgnorePaths, err := readIgnoreFile(contextDir)
	if err != nil {
		return nil, err
	}
	return New(kitIgnorePaths, kitfile, extraLayers...)
}

func New(kitIgnorePaths []string, kitfile *artifact.KitFile, extraLayers ...string) (Paths, error) {
	kitIgnorePaths = append(kitIgnorePaths, constants.DefaultKitfileNames()...)
	kitIgnorePaths = append(kitIgnorePaths, constants.IgnoreFileName)
	kitIgnorePM, err := patternmatcher.New(kitIgnorePaths)
	if err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", constants.IgnoreFileName, err)
	}

	layerPaths := kfutils.LayerPathsFromKitfile(kitfile)
	layerPaths = append(layerPaths, extraLayers...)
	return &ignorePaths{
		ignoreFileMatcher: kitIgnorePM,
		layers:            layerPaths,
	}, nil
}

type ignorePaths struct {
	ignoreFileMatcher *patternmatcher.PatternMatcher
	layers            []string
}

func (pm *ignorePaths) Matches(path, layerPath string) (bool, error) {
	path = cleanPath(path)
	layerPath = cleanPath(layerPath)
	ignoreFileMatches, err := pm.ignoreFileMatcher.MatchesOrParentMatches(path)
	if err != nil {
		return false, err
	}
	if ignoreFileMatches {
		return true, nil
	}
	// ignore file doesn't exclude the current path, check if it should be excluded
	// since it's included in another layer
	for _, layer := range pm.layers {
		layer = cleanPath(layer)
		if strings.HasPrefix(layerPath, layer) {
			// ignore other layer paths if they are parents of the current layer's path,
			// e.g. ignore ./main-dir when current layer is ./main-dir/sub-dir
			continue
		}
		if strings.HasPrefix(path, layer) {
			// The current path is included in another layer that is a subdirectory of the current layer
			return true, nil
		}
	}

	return false, nil
}

func (pm *ignorePaths) HasExclusions() bool {
	return pm.ignoreFileMatcher.Exclusions()
}

func readIgnoreFile(contextDir string) ([]string, error) {
	ignorePath := filepath.Join(contextDir, constants.IgnoreFileName)
	ignoreFile, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open %s file: %w", constants.IgnoreFileName, err)
	}
	patterns, err := ignorefile.ReadAll(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", constants.IgnoreFileName, err)
	}
	return patterns, nil
}

func cleanPath(path string) string {
	return filepath.Clean(strings.TrimSpace(path))
}
