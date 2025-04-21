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
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
)

type filterConf struct {
	baseTypes []string
	filters   []string
}

func (fc *filterConf) matches(baseType, field string) bool {
	return fc.matchesBaseType(baseType) && fc.matchesField(field)
}

func (fc *filterConf) matchesBaseType(baseType string) bool {
	// Treat modelparts as covered by the 'model' filter
	if baseType == constants.ModelPartType {
		baseType = constants.ModelType
	}
	for _, t := range fc.baseTypes {
		if t == baseType {
			return true
		}
	}
	return false
}

func (fc *filterConf) matchesField(field string) bool {
	if len(fc.filters) == 0 {
		// By default everything matches
		return true
	}
	return slices.Contains(fc.filters, field)
}

func parseFilter(filter string) (*filterConf, error) {
	typesAndIds := strings.Split(filter, ":")

	if len(typesAndIds) > 2 {
		return nil, fmt.Errorf("invalid filter: should be in format <type1>,<type2>[:<filter1>,<filter2>]")
	}

	conf := &filterConf{}

	for _, filterType := range strings.Split(typesAndIds[0], ",") {
		baseType, err := filterToMediaBaseType(filterType)
		if err != nil {
			return nil, err
		}
		conf.baseTypes = append(conf.baseTypes, baseType)
	}

	// Check for additional filtering based on name/path
	if len(typesAndIds) == 1 {
		return conf, nil
	}

	filters := strings.Split(typesAndIds[1], ",")
	conf.filters = filters
	return conf, nil
}

// shouldUnpackLayer determines if we should unpack a layer in a Kitfile by matching
// fields against the filters. Matching is done against path and name (if present).
// If filters is empty, we assume everything should be unpacked
func shouldUnpackLayer(layer any, filters []filterConf) bool {
	if len(filters) == 0 {
		return true
	}
	// The type switch below checks for concrete (non-pointer) types. We need to use
	// reflect to dereference the pointer and get a new interface{} (any) type.
	if val := reflect.ValueOf(layer); val.Kind() == reflect.Ptr {
		layer = val.Elem().Interface()
	}

	switch l := layer.(type) {
	case artifact.KitFile:
		for _, filter := range filters {
			if slices.Contains(filter.baseTypes, constants.ConfigType) {
				return true
			}
		}
		return false
	case artifact.Model:
		return matchesFilters(l.Name, constants.ModelType, filters) || matchesFilters(l.Path, constants.ModelType, filters)
	case artifact.ModelPart:
		return matchesFilters(l.Name, constants.ModelPartType, filters) || matchesFilters(l.Path, constants.ModelPartType, filters)
	case artifact.Docs:
		// Docs does not have an ID/name field so we can only match on path
		return matchesFilters(l.Path, constants.DocsType, filters)
	case artifact.DataSet:
		return matchesFilters(l.Name, constants.DatasetType, filters) || matchesFilters(l.Path, constants.DatasetType, filters)
	case artifact.Code:
		// Code does not have a ID/name field so we can only match on path
		return matchesFilters(l.Path, constants.CodeType, filters)
	default:
		return false
	}
}

func matchesFilters(field string, baseType string, filterConfs []filterConf) bool {
	for _, filterConf := range filterConfs {
		if filterConf.matches(baseType, field) {
			return true
		}
	}
	return false
}

// filtersFromUnpackConf converts a (deprecated) unpackConf to a set of filters to enable supporting the old flags
func filtersFromUnpackConf(conf unpackConf) []filterConf {
	filter := filterConf{}

	if conf.unpackKitfile {
		filter.baseTypes = append(filter.baseTypes, constants.ConfigType)
	}
	if conf.unpackModels {
		filter.baseTypes = append(filter.baseTypes, constants.ModelType)
	}
	if conf.unpackDocs {
		filter.baseTypes = append(filter.baseTypes, constants.DocsType)
	}
	if conf.unpackDatasets {
		filter.baseTypes = append(filter.baseTypes, constants.DatasetType)
	}
	if conf.unpackCode {
		filter.baseTypes = append(filter.baseTypes, constants.CodeType)
	}
	return []filterConf{filter}
}

func filterToMediaBaseType(filterType string) (string, error) {
	switch filterType {
	case "kitfile":
		return constants.ConfigType, nil
	case "datasets":
		// annoyingly, the mediatype is dataset, but for the filter we want the plural
		return constants.DatasetType, nil
	case constants.ModelType, constants.CodeType, constants.DocsType:
		return filterType, nil
	default:
		return "", fmt.Errorf("invalid filter type %s (must be one of 'kitfile', 'model', 'datasets', 'code', or 'docs')", filterType)
	}
}
