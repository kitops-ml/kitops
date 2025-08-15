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

package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"
	"oras.land/oras-go/v2/registry"
)

type DevBaseOptions struct {
	configHome string
}

type DevLogsOptions struct {
	DevBaseOptions
	follow bool
}

func (opts *DevBaseOptions) complete(ctx context.Context, _ []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	return nil
}

type DevStartOptions struct {
	DevBaseOptions
	options.NetworkOptions
	host       string
	port       int
	modelFile  string
	contextDir string
	modelRef   *registry.Reference // For ModelKit references
}

func (opts *DevStartOptions) complete(ctx context.Context, args []string) error {
	if err := opts.DevBaseOptions.complete(ctx, args); err != nil {
		return err
	}

	if len(args) == 1 {
		// Check if the argument is a ModelKit reference
		if ref, _, err := util.ParseReference(args[0]); err == nil && ref.Reference != "" {
			// This looks like a ModelKit reference
			opts.modelRef = ref
			output.Debugf("Detected ModelKit reference: %s", ref.String())
		} else {
			// This is a directory path
			opts.contextDir = args[0]
			output.Debugf("Using directory path: %s", args[0])
		}
	}

	// If we have a ModelKit reference but no explicit modelFile flag, we'll extract to cache
	if opts.modelRef != nil && opts.modelFile == "" {
		// We'll set contextDir to cache directory after extraction
		output.Debugf("Will extract ModelKit reference to cache directory")
	} else {
		// Original directory-based logic
		if opts.modelFile == "" {
			foundKitfile, err := filesystem.FindKitfileInPath(opts.contextDir)
			if err != nil {
				if opts.modelRef == nil && opts.contextDir == "" {
					return fmt.Errorf("no directory or ModelKit reference provided - specify a directory path or ModelKit reference (e.g., myrepo/my-model:latest)")
				}
				return fmt.Errorf("failed to find Kitfile in directory %s: %w", opts.contextDir, err)
			}
			opts.modelFile = foundKitfile
		}
	}

	if opts.host == "" {
		opts.host = "127.0.0.1"
	}

	if opts.port == 0 {
		availPort, err := findAvailablePort()
		if err != nil {
			return fmt.Errorf("Invalid arguments: %s", err)
		}
		opts.port = availPort
	}

	// Complete network options for remote access
	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func (opts *DevStartOptions) cleanup() error {
	// Only clean up if we extracted a ModelKit reference to contextDir
	if opts.modelRef != nil && opts.contextDir != "" {
		return os.RemoveAll(opts.contextDir)
	}
	return nil
}
