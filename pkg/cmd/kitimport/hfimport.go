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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/external/hf"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

func importUsingHF(ctx context.Context, opts *importOptions) (*ProvenanceData, error) {
	var prov *ProvenanceData
	if opts.attestationOutput != "" {
		prov = newProvenanceData()
	}

	// Handle full HF URLs by extracting repository name from URL
	repo, repoType, err := hf.ParseHuggingFaceRepo(opts.repo)
	if err != nil {
		return nil, fmt.Errorf("could not process repository %s: %w", opts.repo, err)
	}

	// Resolve the user-supplied ref to an immutable commit SHA up front.
	// HF accepts commit SHAs anywhere a ref is expected, so threading this
	// SHA through ListFiles and DownloadFiles binds the entire import to one
	// snapshot — even if the branch moves between calls.
	commitSHA, err := hf.PinCommit(ctx, repo, opts.repoRef, opts.token, repoType)
	if err != nil {
		return nil, fmt.Errorf("failed to pin commit for %s@%s: %w", repo, opts.repoRef, err)
	}

	// Reuse one cache directory per immutable HF snapshot so interrupted
	// downloads can be resumed. Successful imports clear the import cache below.
	tmpDir, _, err := cache.MkCacheDir(cache.CacheImportSubdir, hfImportCacheKey(repo, repoType, commitSHA))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	dirListing, err := hf.ListFiles(ctx, repo, commitSHA, opts.token, repoType)
	if err != nil {
		return nil, fmt.Errorf("failed to list files from HuggingFace API: %w", err)
	}
	if prov != nil {
		prov.SourceURI = hfSourceURI(repo, repoType)
		prov.SourceCommitSHA = commitSHA
	}

	var kitfile *artifact.KitFile
	if opts.kitfilePath == "-" {
		kitfile = &artifact.KitFile{}
		if err := kitfile.LoadModel(os.Stdin); err != nil {
			return nil, fmt.Errorf("failed to read Kitfile from input: %w", err)
		}
	} else if opts.kitfilePath != "" {
		kf, err := readExistingKitfile(opts.kitfilePath)
		if err != nil {
			return nil, err
		}
		kitfile = kf
	} else {
		kf, err := generateKitfile(dirListing, repo, tmpDir, opts.depth)
		if err != nil {
			return nil, err
		}
		kitfile = kf

		if util.IsInteractiveSession() {
			// If we hit an error here, we don't want to clean up files so that user
			// can manually edit them.
			newKitfile, err := promptToEditKitfile(tmpDir, kf)
			if err != nil {
				if errors.Is(err, ErrNoEditorFound) {
					kfPath := filepath.Join(tmpDir, constants.DefaultKitfileName)
					output.Logf(output.LogLevelWarn, "Could not determine default editor from $EDITOR environment variable")
					output.Logf(output.LogLevelWarn, "Please manually edit Kitfile at path")
					output.Logf(output.LogLevelWarn, "    %s", kfPath)
					output.Logf(output.LogLevelWarn, "and run command")
					output.Logf(output.LogLevelWarn, "    kit import %s -t %s -f %s", opts.repo, opts.tag, kfPath)
					output.Logf(output.LogLevelWarn, "to complete process")
					return nil, err
				}
				return nil, err
			}
			kitfile = newKitfile
		}
	}

	toDownload, err := filterListingForKitfile(dirListing, kitfile)
	if err != nil {
		return nil, err
	}
	if err := hf.DownloadFiles(ctx, repo, commitSHA, tmpDir, toDownload, opts.token, opts.concurrency, repoType); err != nil {
		return nil, fmt.Errorf("error downloading repository: %w", err)
	}

	output.Infof("Packing model to %s", opts.tag)
	manifestDesc, err := packDirectory(ctx, opts.configHome, tmpDir, kitfile, opts.modelKitRef)
	if err != nil {
		return nil, fmt.Errorf("failed to pack ModelKit: %w", err)
	}
	output.Infof("Model is packed as %s", opts.tag)
	if prov != nil {
		if err := prov.finalize(manifestDesc, kitfile); err != nil {
			return nil, err
		}
	}

	if err := cache.CleanCacheDir(cache.CacheImportSubdir); err != nil {
		output.Logf(output.LogLevelWarn, "Failed to clean cache directory: %s", err)
	}

	return prov, nil
}

// filterListingForKitfile walks the directory listing and returns the subset
// of files that the Kitfile would pack, so HF only downloads what will end up
// in the ModelKit. Layer assignment is decided by kitfilePathFilter, which
// mirrors pack's own ignore-aware layer routing.
func filterListingForKitfile(contents *kfgen.DirectoryListing, kitfile *artifact.KitFile) ([]kfgen.FileListing, error) {
	filter, err := newKitfilePathFilter(kitfile)
	if err != nil {
		return nil, err
	}

	var pathsToDownload []kfgen.FileListing
	var processDir func(dir *kfgen.DirectoryListing) error
	processDir = func(dir *kfgen.DirectoryListing) error {
		for _, file := range dir.Files {
			matches, err := filter.Matches(file.Path)
			if err != nil {
				return fmt.Errorf("failed to process path %s: %w", file.Path, err)
			}
			if matches {
				pathsToDownload = append(pathsToDownload, file)
			}
		}
		for _, subDir := range dir.Subdirs {
			if err := processDir(&subDir); err != nil {
				return err
			}
		}
		return nil
	}
	if err := processDir(contents); err != nil {
		return nil, err
	}

	return pathsToDownload, nil
}

// hfSourceURI builds the SourceURI for a HF import. Datasets get a "datasets/"
// segment so a verifier can distinguish them from models — both kinds of repo
// share the org/name shape but live at different URLs on huggingface.co.
func hfSourceURI(repo string, repoType hf.RepositoryType) string {
	if repoType == hf.RepoTypeDataset {
		return "hf://datasets/" + repo
	}
	return "hf://" + repo
}

func hfImportCacheKey(repo string, repoType hf.RepositoryType, commitSHA string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d\x00%s\x00%s", repoType, repo, commitSHA)))
	return "hf_" + hex.EncodeToString(sum[:])
}
