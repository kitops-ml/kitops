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
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	"github.com/kitops-ml/kitops/pkg/lib/git"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

func importUsingGit(ctx context.Context, opts *importOptions) error {
	tmpDir, cleanupTmp, err := cache.MkCacheDir(cache.CacheImportSubdir, "")
	if err != nil {
		return err
	}
	doCleanup := true
	defer func() {
		if doCleanup {
			cleanupTmp()
		}
	}()

	if err := cloneRepository(opts.repo, tmpDir, opts.token); err != nil {
		return err
	}

	var kitfile *artifact.KitFile
	if opts.kitfilePath == "-" {
		kitfile = &artifact.KitFile{}
		if err := kitfile.LoadModel(os.Stdin); err != nil {
			return fmt.Errorf("failed to read Kitfile from input: %w", err)
		}
		if err := kfutils.ValidateKitfile(kitfile); err != nil {
			return err
		}
	} else if opts.kitfilePath != "" {
		kf, err := readExistingKitfile(opts.kitfilePath)
		if err != nil {
			return err
		}
		kitfile = kf
	} else if kfpath, err := filesystem.FindKitfileInPath(tmpDir); err == nil {
		kf, err := readExistingKitfile(kfpath)
		if err != nil {
			return err
		}
		kitfile = kf
	} else {
		dirContents, err := kfgen.DirectoryListingFromFS(tmpDir)
		if err != nil {
			return fmt.Errorf("error processing directory: %w", err)
		}
		kf, err := generateKitfile(dirContents, opts.repo, tmpDir)
		if err != nil {
			return err
		}
		kitfile = kf

		if util.IsInteractiveSession() {
			// If we hit an error here, we don't want to clean up files so that user
			// can manually edit them.
			newKitfile, err := promptToEditKitfile(tmpDir, kf)
			if err != nil {
				if errors.Is(err, ErrNoEditorFound) {
					doCleanup = false
					output.Logf(output.LogLevelWarn, "Could not determine default editor from $EDITOR environment variable")
					output.Logf(output.LogLevelWarn, "Please manually edit Kitfile at path")
					output.Logf(output.LogLevelWarn, "    %s", filepath.Join(tmpDir, constants.DefaultKitfileName))
					output.Logf(output.LogLevelWarn, "and run command")
					output.Logf(output.LogLevelWarn, "    kit pack -t %s %s", opts.tag, tmpDir)
					output.Logf(output.LogLevelWarn, "to complete process")
					return err
				}
				return err
			}
			kitfile = newKitfile
		}
	}

	output.Infof("Packing model to %s", opts.tag)
	if err := packDirectory(ctx, opts.configHome, tmpDir, kitfile, opts.modelKitRef); err != nil {
		return fmt.Errorf("failed to pack ModelKit: %w", err)
	}
	output.Infof("Model is packed as %s", opts.tag)

	if err := cache.CleanCacheDir(cache.CacheImportSubdir); err != nil {
		output.Logf(output.LogLevelWarn, "Failed to clean cache directory: %s", err)
	}

	return nil
}

func cloneRepository(repo, destDir, token string) error {
	fullRepo := repo
	if !strings.HasPrefix(fullRepo, "http") {
		fullRepo = fmt.Sprintf("https://huggingface.co/%s", repo)
	}
	if err := git.CloneRepository(fullRepo, destDir, token); err != nil {
		return err
	}
	// Clean up git-related files, since we probably don't want those
	if err := git.CleanGitMetadata(destDir); err != nil {
		return err
	}
	return nil
}
