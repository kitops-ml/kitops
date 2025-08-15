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
	"github.com/kitops-ml/kitops/pkg/lib/harness"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/lib/unpack"
	"github.com/kitops-ml/kitops/pkg/output"
)

func runDev(ctx context.Context, options *DevStartOptions) error {
	signalCtx, cancelSignalHandling := context.WithCancel(ctx)
	defer cancelSignalHandling()
	cleanupDone := make(chan bool, 1)
	signalChan := make(chan os.Signal, 1)

	// If we have a ModelKit reference, extract it to cache directory first
	if options.modelRef != nil {
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signalChan)

		if err := extractModelKitToCache(ctx, options); err != nil {
			return fmt.Errorf("failed to extract ModelKit: %w", err)
		}

		// Set up cleanup goroutine for signals
		go func() {
			select {
			case <-signalChan:
				output.Infof("Received interrupt signal, cleaning up...")
				if cleanupErr := options.cleanup(); cleanupErr != nil {
					output.Debugf("Failed to cleanup cache directory: %v", cleanupErr)
				}
				cleanupDone <- true
			case <-cleanupDone:
			case <-signalCtx.Done():
				return
			}
		}()
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

	llmHarness := &harness.LLMHarness{}
	llmHarness.ConfigHome = options.configHome

	if err := llmHarness.Init(); err != nil {
		return err
	}

	if err := llmHarness.Stop(); err != nil {
		return err
	}

	// Clean up cache directory
	cacheDir := filepath.Join(options.configHome, "dev-models", "current")
	if err := os.RemoveAll(cacheDir); err != nil {
		output.Debugf("Failed to clean up cache directory: %v", err)
	} else {
		output.Infof("Cleaned up cache directory")
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
	extractDir := filepath.Join(options.configHome, "dev-models", "current")
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

	// Change working directory to cache directory unpack logic is relative to CWD
	originalWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	if err := os.Chdir(extractDir); err != nil {
		return fmt.Errorf("failed to change to cache directory %s: %w", extractDir, err)
	}
	// Restore original working directory when done
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			output.Debugf("Failed to restore working directory: %v", err)
		}
	}()

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

	// Find the Kitfile in the extracted directory
	kitfilePath := filepath.Join(extractDir, constants.DefaultKitfileName)
	if _, err := os.Stat(kitfilePath); err != nil {
		return fmt.Errorf("kitfile not found in extracted ModelKit: %w", err)
	}
	options.modelFile = kitfilePath

	output.Infof("ModelKit extracted to %s", extractDir)
	return nil
}
