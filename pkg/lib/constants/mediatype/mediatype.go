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

import (
	"fmt"
	"regexp"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var kitopsMediaTypeRegexp = regexp.MustCompile(`^application/vnd\.kitops\.modelkit\.(\w+)\.v1\.tar(?:\+(\w+))?$`)
var modelPackMediaTypeRegexp = regexp.MustCompile(`^application/vnd\.cncf\.model\.(\w+(?:\.\w+)?)\.v1\.(\w+)(?:\+?(\w+))?$`)

type MediaType interface {
	Base() BaseType
	Compression() CompressionType
	Format() Format
	String() string
	UserString() string
}

type BaseType int

const (
	UnknownBaseType BaseType = iota
	// ConfigBaseType is the base type for model configs
	ConfigBaseType
	// ModelBaseType is the base type for primary model files
	ModelBaseType
	// ModelPartBaseType is the base type for model-related files. In ModelPack formats, it is
	// reused for the `model.config` type
	ModelPartBaseType
	// DatasetBaseType is the base type for dataset layers
	DatasetBaseType
	// CodeBaseType is the base type for code layers
	CodeBaseType
	// DocsBaseType is the base type for documentation layers
	DocsBaseType
)

type CompressionType int

const (
	UnknownCompression CompressionType = iota
	NoneCompression
	GzipCompression
	GzipFastestCompression
	ZstdCompression
)

type Format int

const (
	UnknownFormat Format = iota
	TarFormat
	RawFormat
)

func ParseMediaType(s string) (MediaType, error) {
	if s == "application/vnd.kitops.modelkit.config.v1+json" {
		return &kitopsMediaType{
			baseType: ConfigBaseType,
		}, nil
	}
	if s == "application/vnd.cncf.model.config.v1+json" {
		return &modelpackMediatype{
			baseType: ConfigBaseType,
		}, nil
	}

	if kitopsMediaTypeRegexp.MatchString(s) {
		match := kitopsMediaTypeRegexp.FindStringSubmatch(s)
		base, compression := match[1], match[2]
		baseType, err := ParseKitBaseType(base)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media type: %w", err)
		}
		compressionType, err := ParseCompression(compression)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media type: %w", err)
		}
		return &kitopsMediaType{
			baseType:        baseType,
			compressionType: compressionType,
		}, nil
	} else if modelPackMediaTypeRegexp.MatchString(s) {
		match := modelPackMediaTypeRegexp.FindStringSubmatch(s)
		base, format, compression := match[1], match[2], match[3]
		baseType, err := ParseModelPackBaseType(base)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media type: %w", err)
		}
		compressionType, err := ParseCompression(compression)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media type: %w", err)
		}
		formatType, err := ParseFormat(format)
		if err != nil {
			return nil, fmt.Errorf("failed to parse media type: %w", err)
		}
		return &modelpackMediatype{
			baseType:        baseType,
			compressionType: compressionType,
			format:          formatType,
		}, nil
	}
	return nil, fmt.Errorf("unrecognized media type %s", s)
}

func NewKit(base BaseType, comp CompressionType) MediaType {
	return &kitopsMediaType{
		baseType:        base,
		compressionType: comp,
	}
}

func ParseCompression(c string) (CompressionType, error) {
	switch c {
	case "", "none":
		return NoneCompression, nil
	case "gzip":
		return GzipCompression, nil
	case "gzip-fastest":
		return GzipFastestCompression, nil
	case "zstd":
		return ZstdCompression, nil
	default:
		return UnknownCompression, fmt.Errorf("invalid compression %s", c)
	}
}

func ParseFormat(f string) (Format, error) {
	switch f {
	case "raw":
		return RawFormat, nil
	case "tar":
		return TarFormat, nil
	}
	return UnknownFormat, fmt.Errorf("invalid format %s", f)
}

func IsValidCompression(c string) error {
	// Not supporting zstd for now; no stable implementation available
	switch c {
	case "none", "gzip", "gzip-fastest":
		return nil
	default:
		return fmt.Errorf("invalid compression type: must be one of 'none', 'gzip', or 'gzip-fastest'")
	}
}

func FormatMediaTypeForUser(mediatype string) string {
	if mediatype == ocispec.MediaTypeImageManifest {
		return "manifest"
	}
	parsed, err := ParseMediaType(mediatype)
	if err != nil {
		// Should never happen
		return "(invalid media type)"
	}
	return parsed.UserString()
}
