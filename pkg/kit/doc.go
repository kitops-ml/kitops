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

// Package kit provides a library API for interacting with ModelKits and OCI registries.
//
// The kit package exposes the core functionality of the Kit CLI as reusable Go functions,
// allowing other programs to integrate ModelKit operations without shelling out to the CLI.
//
// # Basic Usage
//
// The library follows standard Go patterns with context-aware functions that accept
// options structs and return errors:
//
//	import (
//		"context"
//		"github.com/kitops-ml/kitops/pkg/kit"
//	)
//
//	// List local ModelKits
//	infos, err := kit.List(ctx, &kit.ListOptions{
//		ConfigHome: "/path/to/config",
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	// Push to registry
//	pushResult, err := kit.Push(ctx, &kit.PushOptions{
//		ConfigHome:    "/path/to/config",
//		SrcModelRef:   modelRef,
//		DestModelRef:  registryRef,
//	})
//
// # Output and Logging
//
// The library uses the global output functions from pkg/output for logging and progress.
// Configure output using output.SetOut(), output.SetErr(), and output.SetLogLevel():
//
//	import "github.com/kitops-ml/kitops/pkg/output"
//
//	output.SetOut(myWriter)
//	output.SetLogLevel(output.LogLevelDebug)
//
// # Error Handling
//
// All library functions return standard Go errors. Wrap them for context or use
// output functions from pkg/output for user-facing messages.
//
//	import "github.com/kitops-ml/kitops/pkg/output"
//
//	result, err := kit.Push(ctx, opts)
//	if err != nil {
//		output.Fatalf("Failed to push: %s", err)
//	}
//
// # Network Configuration
//
// NetworkOptions can be embedded in option structs to control remote registry access:
//
//	import "github.com/kitops-ml/kitops/pkg/cmd/options"
//
//	opts := &kit.PushOptions{
//		NetworkOptions: options.NetworkOptions{
//			PlainHTTP: false,
//			TLSVerify: true,
//			Concurrency: 5,
//		},
//	}
//
// For complete examples, see the Example functions in this package.
package kit
