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
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"oras.land/oras-go/v2/registry"
)

// UnpackOptions represents the configuration for unpacking operations.
// This is the main options structure used by both command and library interfaces.
type UnpackOptions struct {
	options.NetworkOptions
	ConfigHome     string
	UnpackDir      string
	Filters        []string
	FilterConfs    []FilterConf
	ModelRef       *registry.Reference
	Overwrite      bool
	IgnoreExisting bool
}

// ExtractOption configures ModelKit extraction behavior.
// Following KitOps functional options pattern.
type ExtractOption func(*extractConfig)

// FilterConfig represents a parsed filter configuration.
// This is the legacy structure maintained for backward compatibility.
type FilterConfig struct {
	BaseTypes []string
	Filters   []string
}

// extractConfig holds the internal configuration for extraction.
type extractConfig struct {
	configHome     string
	filters        []string
	filterConfigs  []FilterConfig
	overwrite      bool
	ignoreExisting bool
	networkOptions *options.NetworkOptions
}

// WithConfigHome sets the KitOps configuration directory.
// This is required for all extraction operations.
func WithConfigHome(configHome string) ExtractOption {
	return func(c *extractConfig) {
		c.configHome = configHome
	}
}

// WithFilters sets the filters to apply during extraction.
// Filters control which components are extracted (e.g., "model", "datasets:my-data").
//
// Example:
//
//	WithFilters("model")                    // Extract only model components
//	WithFilters("model", "datasets:train")  // Extract model and specific dataset
func WithFilters(filters ...string) ExtractOption {
	return func(c *extractConfig) {
		c.filters = filters
	}
}

// WithFilterConfigs sets pre-parsed filter configurations.
// Use this with the result of ParseFilters() for validation before extraction.
func WithFilterConfigs(configs []FilterConfig) ExtractOption {
	return func(c *extractConfig) {
		c.filterConfigs = configs
	}
}

// WithOverwrite controls whether existing files should be overwritten.
// Default is true for library usage (safe for temporary directories).
func WithOverwrite(overwrite bool) ExtractOption {
	return func(c *extractConfig) {
		c.overwrite = overwrite
	}
}

// WithIgnoreExisting controls whether existing files should be skipped.
// Default is false (report conflicts rather than silently skipping).
func WithIgnoreExisting(ignore bool) ExtractOption {
	return func(c *extractConfig) {
		c.ignoreExisting = ignore
	}
}

// WithNetworkOptions sets network configuration for remote registry access.
// Required when extracting ModelKits from remote registries.
func WithNetworkOptions(opts *options.NetworkOptions) ExtractOption {
	return func(c *extractConfig) {
		c.networkOptions = opts
	}
}
