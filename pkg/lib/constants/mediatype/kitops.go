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

package mediatype

import "fmt"

var KitConfigMediaType MediaType = &kitopsMediaType{
	baseType: ConfigBaseType,
}

type kitopsMediaType struct {
	baseType        BaseType
	compressionType CompressionType
	format          Format
}

func (mt *kitopsMediaType) Base() BaseType {
	return mt.baseType
}

func (mt *kitopsMediaType) Compression() CompressionType {
	return mt.compressionType
}

func (mt *kitopsMediaType) Format() Format {
	return mt.format
}

func (mt *kitopsMediaType) String() string {
	if mt.baseType == ConfigBaseType {
		return "application/vnd.kitops.modelkit.config.v1+json"
	}

	switch mt.compressionType {
	case NoneCompression:
		return fmt.Sprintf("application/vnd.kitops.modelkit.%s.v1.%s", mt.baseTypeString(), mt.format.String())
	case GzipCompression, GzipFastestCompression:
		return fmt.Sprintf("application/vnd.kitops.modelkit.%s.v1.%s+gzip", mt.baseTypeString(), mt.format.String())
	case ZstdCompression:
		return fmt.Sprintf("application/vnd.kitops.modelkit.%s.v1.%s+zstd", mt.baseTypeString(), mt.format.String())
	}
	// Should never happen since parsing should only result in valid values
	return "invalid mediatype"
}

func (mt *kitopsMediaType) UserString() string {
	return mt.baseTypeString()
}

func (mt *kitopsMediaType) baseTypeString() string {
	switch mt.baseType {
	case ConfigBaseType:
		return "config"
	case ModelBaseType:
		return "model"
	case ModelPartBaseType:
		return "modelpart"
	case DatasetBaseType:
		return "dataset"
	case CodeBaseType:
		return "code"
	case DocsBaseType:
		return "docs"
	}
	return "invalid mediatype"
}

var _ MediaType = (*kitopsMediaType)(nil)

func ParseKitBaseType(s string) (BaseType, error) {
	switch s {
	case "config":
		return ConfigBaseType, nil
	case "model":
		return ModelBaseType, nil
	case "modelpart":
		return ModelPartBaseType, nil
	case "dataset":
		return DatasetBaseType, nil
	case "code":
		return CodeBaseType, nil
	case "docs":
		return DocsBaseType, nil
	default:
		return UnknownBaseType, fmt.Errorf("invalid base type %s", s)
	}
}
