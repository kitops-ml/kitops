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

package kit

import (
	"context"

	"github.com/kitops-ml/kitops/pkg/lib/filesystem/unpack"
)

// UnpackOptions re-exports unpack.UnpackOptions for library API.
type UnpackOptions = unpack.UnpackOptions

// FilterConf re-exports unpack.FilterConf for library API.
type FilterConf = unpack.FilterConf

// Unpack unpacks a ModelKit to the filesystem.
func Unpack(ctx context.Context, opts *UnpackOptions) error {
	return unpack.UnpackModelKit(ctx, opts)
}

// ParseFilter re-exports unpack.ParseFilter for convenience.
func ParseFilter(filter string) (*FilterConf, error) {
	return unpack.ParseFilter(filter)
}

//
