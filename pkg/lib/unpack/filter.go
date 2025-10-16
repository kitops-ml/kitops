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
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
)

// FilterConf represents filter configuration for unpacking operations.
type FilterConf struct {
	BaseTypes []string
	Filters   []string
}

func (fc *FilterConf) matches(baseType, field string) bool {
	return fc.matchesBaseType(baseType) && fc.matchesField(field)
}

func (fc *FilterConf) matchesBaseType(baseType string) bool {
	// Treat modelparts as covered by the 'model' filter
	if baseType == mediatype.ModelPartType {
		baseType = mediatype.ModelType
	}
	for _, t := range fc.BaseTypes {
		if t == baseType {
			return true
		}
	}
	return false
}

func (fc *FilterConf) matchesField(field string) bool {
	if len(fc.Filters) == 0 {
		// By default everything matches
		return true
	}
	return slices.Contains(fc.Filters, field)
}

// ParseFilter parses a filter string and returns a FilterConf.
func ParseFilter(filter string) (*FilterConf, error) {
	typesAndIds := strings.Split(filter, ":")

	if len(typesAndIds) > 2 {
		return nil, fmt.Errorf("invalid filter: should be in format <type1>,<type2>[:<filter1>,<filter2>]")
	}

	conf := &FilterConf{}

	for _, filterType := range strings.Split(typesAndIds[0], ",") {
		baseType, err := filterToMediaBaseType(filterType)
		if err != nil {
			return nil, err
		}
		conf.BaseTypes = append(conf.BaseTypes, baseType)
	}

	// Check for additional filtering based on name/path
	if len(typesAndIds) == 1 {
		return conf, nil
	}

	filters := strings.Split(typesAndIds[1], ",")
	conf.Filters = filters
	return conf, nil
}

// shouldUnpackLayer determines if we should unpack a layer in a Kitfile by matching
// fields against the filters. Matching is done against path and name (if present).
// If filters is empty, we assume everything should be unpacked
func shouldUnpackLayer(layer any, filters []FilterConf) bool {
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
			if slices.Contains(filter.BaseTypes, mediatype.ConfigType) {
				return true
			}
		}
		return false
	case artifact.Model:
		return matchesFilters(l.Name, mediatype.ModelType, filters) || matchesFilters(l.Path, mediatype.ModelType, filters)
	case artifact.ModelPart:
		return matchesFilters(l.Name, mediatype.ModelPartType, filters) || matchesFilters(l.Path, mediatype.ModelPartType, filters)
	case artifact.Docs:
		// Docs does not have an ID/name field so we can only match on path
		return matchesFilters(l.Path, mediatype.DocsType, filters)
	case artifact.DataSet:
		return matchesFilters(l.Name, mediatype.DatasetType, filters) || matchesFilters(l.Path, mediatype.DatasetType, filters)
	case artifact.Code:
		// Code does not have a ID/name field so we can only match on path
		return matchesFilters(l.Path, mediatype.CodeType, filters)
	default:
		return false
	}
}

func matchesFilters(field string, baseType string, filterConfs []FilterConf) bool {
	for _, filterConf := range filterConfs {
		if filterConf.matches(baseType, field) {
			return true
		}
	}
	return false
}

// FiltersFromUnpackConf converts a (deprecated) unpackConf to a set of filters to enable supporting the old flags
func FiltersFromUnpackConf(unpackKitfile, unpackModels, unpackCode, unpackDatasets, unpackDocs bool) []FilterConf {
	filter := FilterConf{}

	if unpackKitfile {
		filter.BaseTypes = append(filter.BaseTypes, mediatype.ConfigType)
	}
	if unpackModels {
		filter.BaseTypes = append(filter.BaseTypes, mediatype.ModelType)
	}
	if unpackDocs {
		filter.BaseTypes = append(filter.BaseTypes, mediatype.DocsType)
	}
	if unpackDatasets {
		filter.BaseTypes = append(filter.BaseTypes, mediatype.DatasetType)
	}
	if unpackCode {
		filter.BaseTypes = append(filter.BaseTypes, mediatype.CodeType)
	}
	return []FilterConf{filter}
}

func filterToMediaBaseType(filterType string) (string, error) {
	switch filterType {
	case "kitfile":
		return mediatype.ConfigType, nil
	case "datasets":
		// annoyingly, the mediatype is dataset, but for the filter we want the plural
		return mediatype.DatasetType, nil
	case mediatype.ModelType, mediatype.CodeType, mediatype.DocsType:
		return filterType, nil
	default:
		return "", fmt.Errorf("invalid filter type %s (must be one of 'kitfile', 'model', 'datasets', 'code', or 'docs')", filterType)
	}
}
