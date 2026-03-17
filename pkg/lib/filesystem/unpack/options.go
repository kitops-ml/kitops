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

package unpack

import (
	"oras.land/oras-go/v2/registry"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/kitfile"
)

// UnpackOptions represents the configuration for unpacking operations.
// This is the main options structure used by both command and library interfaces.
type UnpackOptions struct {
	options.NetworkOptions
	ConfigHome     string
	UnpackDir      string
	Filters        []string
	FilterConfs    []kitfile.FilterConf
	ModelRef       *registry.Reference
	Overwrite      bool
	IgnoreExisting bool
	IncludeRemote  bool
}
