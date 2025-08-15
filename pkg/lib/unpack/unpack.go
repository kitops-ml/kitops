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

// Package unpack provides programmatic access to ModelKit unpacking functionality.
// This package exposes the core unpack operations as a library for use by other
// packages, separate from the CLI command implementation.
//
// The library provides a facade pattern around the core unpack functionality,
// making it easy for other packages to extract ModelKits without subprocess calls.
//
// Example usage:
//
//	// Extract only model components from a ModelKit reference
//	err := unpack.ExtractModelKit(ctx, modelRef, "/tmp/extracted",
//	    unpack.WithConfigHome("/home/user/.kit"),
//	    unpack.WithFilters("model"),
//	    unpack.WithOverwrite(true),
//	    unpack.WithNetworkOptions(&networkOpts),
//	)
//
//	// Extract multiple component types with specific filters
//	err := unpack.ExtractModelKit(ctx, modelRef, "/tmp/extracted",
//	    unpack.WithFilters("model", "datasets:training-data"),
//	    unpack.WithConfigHome(configHome),
//	)
package unpack

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"oras.land/oras-go/v2/registry"
)

// ExtractModelKit extracts a ModelKit to the specified directory.
// This is the main entry point for programmatic ModelKit extraction.
//
// Example:
//
//	ref, _ := util.ParseReference("myrepo/model:latest")
//	err := unpack.ExtractModelKit(ctx, ref, "/tmp/extracted",
//	                              unpack.WithConfigHome(configHome),
//	                              unpack.WithFilters("model"))
func ExtractModelKit(ctx context.Context, modelRef *registry.Reference, unpackDir string, opts ...ExtractOption) error {
	if ctx == nil {
		return fmt.Errorf("cannot extract ModelKit: context is required")
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("cannot extract ModelKit: %w", ctx.Err())
	default:
	}

	if modelRef == nil {
		return fmt.Errorf("cannot extract ModelKit: reference is required")
	}
	if unpackDir == "" {
		return fmt.Errorf("cannot extract ModelKit: unpack directory is required")
	}

	// Build configuration from options
	config := &extractConfig{
		overwrite:      true,  // Default for library usage
		ignoreExisting: false, // Default for library usage
	}

	for _, opt := range opts {
		opt(config)
	}

	// Validate required configuration
	if config.configHome == "" {
		return fmt.Errorf("failed to initialize unpack operation: config home is required")
	}

	absUnpackDir, err := filepath.Abs(unpackDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path %s: %w", unpackDir, err)
	}

	unpackOpts := &UnpackOptions{
		ModelRef:       modelRef,
		UnpackDir:      absUnpackDir,
		ConfigHome:     config.configHome,
		Filters:        config.filters,
		Overwrite:      config.overwrite,
		IgnoreExisting: config.ignoreExisting,
	}

	if config.networkOptions != nil {
		unpackOpts.NetworkOptions = *config.networkOptions
	}

	// Parse filters if provided
	if len(config.filters) > 0 {
		for _, filter := range config.filters {
			filterConf, err := ParseFilter(filter)
			if err != nil {
				return fmt.Errorf("invalid filter %q: %w", filter, err)
			}
			unpackOpts.FilterConfs = append(unpackOpts.FilterConfs, *filterConf)
		}
	} else if len(config.filterConfigs) > 0 {
		// Convert legacy FilterConfig to FilterConf
		for _, fc := range config.filterConfigs {
			unpackOpts.FilterConfs = append(unpackOpts.FilterConfs, FilterConf(fc))
		}
	}

	return UnpackModelKit(ctx, unpackOpts)
}

// ParseFilters parses filter strings and validates them.
// Returns filter configurations that can be used with WithFilterConfigs.
//
// Example:
//
//	filters, err := unpack.ParseFilters([]string{"model", "datasets:my-data"})
//	if err != nil { ... }
//	err = unpack.ExtractModelKit(ctx, ref, dir, unpack.WithFilterConfigs(filters))
func ParseFilters(filters []string) ([]FilterConfig, error) {
	configs := make([]FilterConfig, len(filters))

	for i, filter := range filters {
		filterConf, err := ParseFilter(filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter %q: %w", filter, err)
		}

		configs[i] = FilterConfig{
			BaseTypes: filterConf.BaseTypes,
			Filters:   filterConf.Filters,
		}
	}

	return configs, nil
}

// GetModelKitStore returns the appropriate store (local or remote) for a ModelKit reference.
// This is useful for checking availability before extraction.
//
// Example:
//
//	store, err := unpack.GetModelKitStore(ctx, ref, configHome, networkOpts)
//	if err != nil { ... }
//	// Use store to check ModelKit details before extraction
func GetModelKitStore(ctx context.Context, modelRef *registry.Reference, configHome string, networkOpts *options.NetworkOptions) (interface{}, error) {
	if modelRef == nil {
		return nil, fmt.Errorf("cannot get ModelKit store: reference is required")
	}
	if configHome == "" {
		return nil, fmt.Errorf("cannot get ModelKit store: config home is required")
	}

	opts := &UnpackOptions{
		ModelRef:   modelRef,
		ConfigHome: configHome,
	}
	if networkOpts != nil {
		opts.NetworkOptions = *networkOpts
	}

	return GetStoreForRef(ctx, opts)
}
