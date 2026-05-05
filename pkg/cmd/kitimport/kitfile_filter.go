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

package kitimport

import (
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
	repoutil "github.com/kitops-ml/kitops/pkg/lib/repo/util"
)

// kitfilePathFilter answers whether a given relative path would be packed
// into some layer of the ModelKit defined by a Kitfile. It mirrors the layer
// assignment that pack uses, so the HF download list agrees with what pack
// will eventually include.
type kitfilePathFilter struct {
	ig     ignore.Paths
	layers []string
}

func newKitfilePathFilter(kitfile *artifact.KitFile) (*kitfilePathFilter, error) {
	ig, err := ignore.New(nil, kitfile)
	if err != nil {
		return nil, fmt.Errorf("failed to process Kitfile to get file list: %w", err)
	}
	return &kitfilePathFilter{
		ig:     ig,
		layers: repoutil.LayerPathsFromKitfile(kitfile),
	}, nil
}

// Matches returns true iff path would be packed into at least one layer.
//
// A file is packed when (a) it falls under some layer's path prefix and
// (b) it is not excluded from that layer. ignore.Paths.Matches(path, layer)
// returns true when path is *excluded* from the given layer — either because
// a .kitignore-style pattern (which always covers Kitfile names and
// .kitignore itself) matches it, or because path belongs to a different
// layer's prefix. The prefix gate is required because ignore.Matches alone
// has no notion of "outside any layer": a path with no layer membership
// returns false for every layer and would otherwise be misreported as
// packed.
func (f *kitfilePathFilter) Matches(path string) (bool, error) {
	for _, layer := range f.layers {
		if !pathUnderLayer(path, layer) {
			continue
		}
		excluded, err := f.ig.Matches(path, layer)
		if err != nil {
			return false, err
		}
		if !excluded {
			return true, nil
		}
	}
	return false, nil
}

// pathUnderLayer reports whether path is contained in the layer's directory.
// "." is a catchall and contains everything. Otherwise the match is on
// directory boundaries so "model_dir" does not contain "model_directory".
//
// Both inputs are normalized to forward-slash form before comparison. Input
// paths reach this filter as forward-slash (HF tree API; kfgen runs
// filepath.ToSlash on filesystem walks), but layer paths go through
// filepath.Clean in repoutil.LayerPathsFromKitfile, which on Windows rewrites
// separators to backslashes for any layer with subdirectories. We do an
// unconditional ReplaceAll rather than filepath.ToSlash so the comparison is
// independent of GOOS — filepath.ToSlash is a no-op on non-Windows hosts and
// would leave a Windows-cleaned layer string mismatched with a forward-slash
// path during cross-platform handling.
func pathUnderLayer(path, layer string) bool {
	if layer == "." {
		return true
	}
	layer = strings.ReplaceAll(layer, `\`, "/")
	path = strings.ReplaceAll(path, `\`, "/")
	if path == layer {
		return true
	}
	return strings.HasPrefix(path, layer+"/")
}
