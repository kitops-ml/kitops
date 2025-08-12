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

import "errors"

// Common errors returned by the unpack library.
// Following KitOps error handling patterns.
var (
	// ErrModelKitNotFound is returned when a ModelKit reference cannot be found
	// in local storage or remote registries.
	ErrModelKitNotFound = errors.New("ModelKit not found")

	// ErrInvalidReference is returned when a ModelKit reference is malformed
	// or cannot be parsed.
	ErrInvalidReference = errors.New("invalid ModelKit reference")

	// ErrInvalidFilter is returned when a filter string cannot be parsed
	// or contains invalid filter specifications.
	ErrInvalidFilter = errors.New("invalid filter specification")

	// ErrConfigRequired is returned when required configuration is missing
	// (e.g., ConfigHome not specified).
	ErrConfigRequired = errors.New("required configuration missing")

	// ErrExtractionFailed is returned when the extraction process fails
	// for reasons other than missing ModelKits or invalid references.
	ErrExtractionFailed = errors.New("ModelKit extraction failed")
)
