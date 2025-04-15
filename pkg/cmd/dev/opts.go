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

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/output"
)

type DevBaseOptions struct {
	configHome string
}

type DevLogsOptions struct {
	DevBaseOptions
	follow bool
}

func (opts *DevBaseOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	return nil
}

type DevStartOptions struct {
	DevBaseOptions
	host       string
	port       int
	modelFile  string
	contextDir string
}

func (opts *DevStartOptions) complete(ctx context.Context, args []string) error {
	if err := opts.DevBaseOptions.complete(ctx, args); err != nil {
		return err
	}

	opts.contextDir = ""
	if len(args) == 1 {
		opts.contextDir = args[0]
	}
	if opts.modelFile == "" {
		foundKitfile, err := filesystem.FindKitfileInPath(opts.contextDir)
		if err != nil {
			return err
		}
		opts.modelFile = foundKitfile
	}
	if opts.host == "" {
		opts.host = "127.0.0.1"
	}

	if opts.port == 0 {
		availPort, err := findAvailablePort()
		if err != nil {
			output.Fatalf("Invalid arguments: %s", err)
			return err
		}
		opts.port = availPort
	}
	return nil
}
