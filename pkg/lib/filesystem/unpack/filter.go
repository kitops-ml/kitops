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
	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/kitfile"
)

type FilterConf = kitfile.FilterConf

func ParseFilter(filter string) (*FilterConf, error) {
	return kitfile.ParseFilter(filter)
}

func KitfileContainsMatchingLayer(kf *artifact.KitFile, filters []FilterConf) bool {
	return kitfile.KitfileContainsMatchingLayer(kf, filters)
}

func FiltersFromUnpackConf(unpackKitfile, unpackModels, unpackCode, unpackDatasets, unpackDocs bool) []FilterConf {
	return kitfile.FiltersFromUnpackConf(unpackKitfile, unpackModels, unpackCode, unpackDatasets, unpackDocs)
}

// AGENT_MODIFIED: Human review required before merge
