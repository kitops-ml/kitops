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

package mediatype

import "fmt"

type modelpackMediatype struct {
	baseType        BaseType
	compressionType CompressionType
	format          Format
}

func (mt *modelpackMediatype) Base() BaseType {
	return mt.baseType
}

func (mt *modelpackMediatype) Compression() CompressionType {
	return mt.compressionType
}

func (mt *modelpackMediatype) Format() Format {
	return Format(mt.format)
}

func (mt *modelpackMediatype) String() string {
	if mt.baseType == ConfigBaseType {
		return "application/vnd.cncf.model.config.v1+json"
	}
	return fmt.Sprintf("application/vnd.cncf.model.%s.v1.%s", mt.baseTypeString(), mt.formatAndCompression())
}

func (mt *modelpackMediatype) UserString() string {
	return mt.baseTypeString()
}

func (mt *modelpackMediatype) baseTypeString() string {
	switch mt.baseType {
	case ConfigBaseType:
		return "config"
	case ModelBaseType:
		return "weight"
	case ModelPartBaseType:
		return "weight.config"
	case DatasetBaseType:
		return "dataset"
	case CodeBaseType:
		return "code"
	case DocsBaseType:
		return "doc"
	}
	return "invalid mediatype"
}

func (mt *modelpackMediatype) formatAndCompression() string {
	// ModelPack does not support compression for raw layers
	switch mt.format {
	case RawFormat:
		return "raw"
	case TarFormat:
		switch mt.compressionType {
		case NoneCompression:
			return "tar"
		case GzipCompression, GzipFastestCompression:
			return "tar+gzip"
		case ZstdCompression:
			return "tar+zstd"
		}
	}
	return "invalid mediatype"
}

var _ MediaType = (*modelpackMediatype)(nil)

func ParseModelPackBaseType(s string) (BaseType, error) {
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
