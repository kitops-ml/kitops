// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/unpack"
	"github.com/kitops-ml/kitops/pkg/lib/harness"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

func runDev(ctx context.Context, options *DevStartOptions) error {
	// Create a context for long-running operations like unpack
	signalCtx, stopSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	if options.modelRef != nil {
		if err := extractModelKitToCache(signalCtx, options); err != nil {
			return fmt.Errorf("failed to extract ModelKit: %w", err)
		}
		// If a signal was received right after extraction, clean up and stop here
		if err := signalCtx.Err(); err != nil {
			output.Infof("Interrupted, cleaning up cache...")
			if cleanupErr := options.cleanup(); cleanupErr != nil {
				output.Debugf("Failed to cleanup cache directory: %v", cleanupErr)
			}
			return signalCtx.Err()
		}
	}

	kitfile := &artifact.KitFile{}

	modelfile, err := os.Open(options.modelFile)
	if err != nil {
		return err
	}
	defer modelfile.Close()
	if err := kitfile.LoadModel(modelfile); err != nil {
		return err
	}
	output.Infof("Loaded Kitfile: %s", options.modelFile)
	if util.IsModelKitReference(kitfile.Model.Path) {
		resolvedKitfile, err := kfutils.ResolveKitfile(ctx, options.configHome, kitfile.Model.Path, kitfile.Model.Path)
		if err != nil {
			return fmt.Errorf("failed to resolve referenced modelkit %s: %w", kitfile.Model.Path, err)
		}
		kitfile.Model.Path = resolvedKitfile.Model.Path
		kitfile.Model.Parts = append(kitfile.Model.Parts, resolvedKitfile.Model.Parts...)
	}

	modelAbsPath, _, err := filesystem.VerifySubpath(options.contextDir, kitfile.Model.Path)
	if err != nil {
		return err
	}

	modelPath, err := findModelFile(modelAbsPath)
	if err != nil {
		return err
	}

	llmHarness := &harness.LLMHarness{}
	llmHarness.Host = options.host
	llmHarness.Port = options.port
	llmHarness.ConfigHome = options.configHome
	if err := llmHarness.Init(); err != nil {
		return err
	}

	if err := llmHarness.Start(modelPath); err != nil {
		return err
	}

	return nil
}

func stopDev(_ context.Context, options *DevBaseOptions) error {

	// Don't fail stopDev if harness is not running.
	// We still want to clean up any cached contents.
	var stopErr error
	llmHarness := &harness.LLMHarness{ConfigHome: options.configHome}
	if err := llmHarness.Init(); err != nil {
		output.Debugf("Harness init failed during stop: %v", err)
		stopErr = err
	} else if err := llmHarness.Stop(); err != nil {
		output.Debugf("Failed to stop dev server: %v", err)
		stopErr = err
	}

	// Always attempt to clean up cache directory
	cacheDir := constants.ExtractedDevModelPath(options.configHome)
	cleanupErr := os.RemoveAll(cacheDir)
	if cleanupErr == nil {
		output.Infof("Cleaned up cache directory")
	} else {
		output.Debugf("Failed to clean up cache directory: %v", cleanupErr)
	}

	// Return errors according to precedence: both -> join; cleanup -> cleanup; stop -> stop; else nil
	if cleanupErr != nil && stopErr != nil {
		return errors.Join(stopErr, fmt.Errorf("failed to clean up cache directory: %w", cleanupErr))
	}
	if cleanupErr != nil {
		return fmt.Errorf("failed to clean up cache directory: %w", cleanupErr)
	}
	if stopErr != nil {
		return stopErr
	}
	return nil
}

func findModelFile(absPath string) (string, error) {
	stat, err := os.Lstat(absPath)
	if err != nil {
		return "", err
	}
	if stat.Mode().IsRegular() {
		// model path refers to a regular file; assume it's fine to use
		return absPath, nil
	} else if !stat.IsDir() {
		return "", fmt.Errorf("could not find model file in %s: path is not regular file or directory", absPath)
	}

	modelPath := ""
	if err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".gguf") && d.Type().IsRegular() {
			if modelPath == "" {
				modelPath = path
			} else {
				return fmt.Errorf("multiple model files found: %s and %s", modelPath, path)
			}
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("error searching for model file in %s: %w", absPath, err)
	} else if modelPath == "" {
		return "", fmt.Errorf("could not find model file in %s", absPath)
	}
	output.Debugf("Found model path in directory %s at %s", absPath, modelPath)
	return modelPath, nil
}

// extractModelKitToCache extracts a ModelKit reference to a cache directory
// using the unpack library with model filter
func extractModelKitToCache(ctx context.Context, options *DevStartOptions) error {
	output.Infof("Extracting ModelKit %s to cache directory...", options.modelRef.String())

	// Use consistent cache directory for extraction
	extractDir := constants.ExtractedDevModelPath(options.configHome)
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	options.contextDir = extractDir

	// Extract the ModelKit using the library directly
	libOpts := &unpack.UnpackOptions{
		ModelRef:       options.modelRef,
		UnpackDir:      extractDir,
		ConfigHome:     options.configHome,
		Overwrite:      true, // Safe for extraction directory
		NetworkOptions: options.NetworkOptions,
	}

	// Add model filter
	modelFilter, err := unpack.ParseFilter("model,kitfile")
	if err != nil {
		return fmt.Errorf("failed to create model filter: %w", err)
	}
	libOpts.FilterConfs = []unpack.FilterConf{*modelFilter}

	err = unpack.UnpackModelKit(ctx, libOpts)
	if err != nil {
		cleanUpErr := os.RemoveAll(extractDir)
		if cleanUpErr != nil {
			return errors.Join(
				fmt.Errorf("failed to extract ModelKit: %w", err),
				fmt.Errorf("failed to cleanup cache directory: %w", cleanUpErr),
			)
		}
		return fmt.Errorf("failed to extract ModelKit: %w", err)
	}

	kitfilePath, err := filesystem.FindKitfileInPath(extractDir)
	if err != nil {
		return fmt.Errorf("kitfile not found in extracted ModelKit: %w", err)
	}
	options.modelFile = kitfilePath

	output.Infof("ModelKit extracted to %s", extractDir)
	return nil
}
