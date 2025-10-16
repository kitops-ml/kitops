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

package kitimport

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"oras.land/oras-go/v2/registry"
)

var ErrNoEditorFound = errors.New("no editor found")

func generateKitfile(dirContents *kfgen.DirectoryListing, repo string, outDir string) (*artifact.KitFile, error) {
	// Fill fields in package so that they're not empty in `kit list` later.
	sections := strings.Split(repo, "/")
	var modelPackage *artifact.Package
	if len(sections) >= 2 {
		modelPackage = &artifact.Package{
			Name:    sections[len(sections)-1],
			Authors: []string{sections[len(sections)-2]},
		}
	}
	kitfile, err := kfgen.GenerateKitfile(dirContents, modelPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Kitfile: %w", err)
	}
	kitfileBytes, err := kitfile.MarshalToYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to write Kitfile: %w", err)
	}
	kitfilePath := filepath.Join(outDir, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, kitfileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write Kitfile: %s", err)
	}
	output.Infof("Generated Kitfile:\n\n%s\n", string(kitfileBytes))
	return kitfile, nil
}

func readExistingKitfile(kfPath string) (*artifact.KitFile, error) {
	kfFile, err := os.Open(kfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kitfile: %w", err)
	}
	defer kfFile.Close()
	kitfile := &artifact.KitFile{}
	if err := kitfile.LoadModel(kfFile); err != nil {
		return nil, fmt.Errorf("failed to load Kitfile from %s: %w", kfPath, err)
	}
	if err := kfutils.ValidateKitfile(kitfile); err != nil {
		return nil, fmt.Errorf("kitfile (%s) is invalid: %w", kfPath, err)
	}
	return kitfile, nil
}

func packDirectory(ctx context.Context, configHome, contextDir string, kitfile *artifact.KitFile, ref *registry.Reference) error {
	// Packing requires the working dir to be the context dir so that relative paths are correct in the tarball
	// On Windows, we need to switch back to the current directory or removing the temporary directory will fail
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(contextDir); err != nil {
		return fmt.Errorf("failed to use context path %s: %w", contextDir, err)
	}
	defer func() {
		if err := os.Chdir(curDir); err != nil {
			output.Logf(output.LogLevelWarn, "Failed to change directory to %s: %s", curDir, err)
		}
	}()

	localRepo, err := local.NewLocalRepo(constants.StoragePath(configHome), ref)
	if err != nil {
		return err
	}
	ignore, err := ignore.NewFromContext(contextDir, kitfile)
	if err != nil {
		return err
	}
	manifestDesc, err := filesystem.SaveModel(ctx, localRepo, kitfile, ignore, constants.NoneCompression)
	if err != nil {
		return err
	}
	if err := localRepo.Tag(ctx, *manifestDesc, ref.Reference); err != nil {
		return fmt.Errorf("failed to tag manifest: %w", err)
	}
	return nil
}

func promptToEditKitfile(contextDir string, currentKitfile *artifact.KitFile) (*artifact.KitFile, error) {
	kitfilePath := filepath.Join(contextDir, constants.DefaultKitfileName)
	ans, err := util.PromptForInput("Would you like to edit Kitfile before packing? (y/N): ", false)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		// Current one is fine!
		return currentKitfile, nil
	}
	editor, err := getEditorName()
	if err != nil {
		return nil, err
	}
	editCmd := exec.Command(editor, kitfilePath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	if err := editCmd.Run(); err != nil {
		return nil, fmt.Errorf("error running external editor: %w", err)
	}
	return readExistingKitfile(kitfilePath)
}

func getEditorName() (string, error) {
	// Default: use the editor in the standard $EDITOR environment variable
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor, nil
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Check for nano, which should be the default editor on Linux and MacOS
		if path, err := exec.LookPath("nano"); err == nil {
			return path, nil
		}
	}
	if runtime.GOOS == "windows" {
		// On current Windows, using notepad will block until the editor (or editor tab)
		// is closed.
		return "notepad", nil
	}

	return "", ErrNoEditorFound
}

// extractRepoFromURL attempts to normalize a string or URL into a repository name as is used on GitHub and Huggingface.
// Returns an error we cannot automatically handle the input URL/string.
//   - https://example.com/segment1/segment2 --> segment1/segment2
//   - 'organization/repository'             --> 'organization/repository'
func extractRepoFromURL(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	path := strings.Trim(u.Path, "/")
	segments := strings.Split(path, "/")
	if len(segments) != 2 {
		return "", fmt.Errorf("could not extract organization and repository from '%s'", path)
	}

	return path, nil
}
