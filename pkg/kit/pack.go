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

package kit

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

type PackOptions struct {
	ConfigHome   string
	ContextDir   string
	KitfilePath  string // "" = auto-discover in ContextDir; "-" = read from stdin
	TagRef       string // comma-separated tags; "" = untagged
	Compression  string // "none" | "gzip" | "gzip-fastest"; "" defaults to "none"
	UseModelPack bool
}

type PackResult struct {
	Descriptor ocispec.Descriptor
}

// Pack packs a ModelKit from a Kitfile and context directory into local storage.
// Note: Pack temporarily changes the process working directory to ContextDir
// (required for correct tar relative paths) and restores it on return.
// It is not safe to call concurrently from multiple goroutines.
func Pack(ctx context.Context, opts *PackOptions) (*PackResult, error) {
	contextDir, err := filepath.Abs(opts.ContextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve context dir %s: %w", opts.ContextDir, err)
	}

	kitfilePath := opts.KitfilePath
	if kitfilePath == "" {
		foundModel, err := filesystem.FindKitfileInPath(contextDir)
		if err != nil {
			return nil, err
		}
		kitfilePath = foundModel
	}

	storageHome := constants.StoragePath(opts.ConfigHome)

	var modelRef *registry.Reference
	var extraRefs []string
	if opts.TagRef != "" {
		ref, extras, err := artifact.ParseReference(opts.TagRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reference: %w", err)
		}
		if ref.Reference == "" {
			ref.Reference = "latest"
		}
		modelRef = ref
		extraRefs = extras
	} else {
		modelRef = artifact.DefaultReference()
	}

	compression := opts.Compression
	if compression == "" {
		compression = "none"
	}
	if err := mediatype.IsValidCompression(compression); err != nil {
		return nil, err
	}

	kitfile, err := readKitfile(kitfilePath)
	if err != nil {
		return nil, err
	}

	localRepo, err := local.NewLocalRepo(storageHome, modelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to open local storage: %w", err)
	}

	// Change working directory to context path so relative paths within tarballs
	// are correct. This is the equivalent of using the -C parameter for tar.
	// Save and restore CWD for library safety.
	origDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	if err := os.Chdir(contextDir); err != nil {
		return nil, fmt.Errorf("failed to use context path %s: %w", contextDir, err)
	}
	defer os.Chdir(origDir)

	var extraLayerPaths []string
	if kitfile.Model != nil && artifact.IsModelKitReference(kitfile.Model.Path) {
		baseRef := artifact.FormatRepositoryForDisplay(modelRef.String())
		parentKitfile, err := kfutils.ResolveKitfile(ctx, opts.ConfigHome, kitfile.Model.Path, baseRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve referenced modelkit %s: %w", kitfile.Model.Path, err)
		}
		extraLayerPaths = util.LayerPathsFromKitfile(parentKitfile)
	}

	ign, err := ignore.NewFromContext(contextDir, kitfile, extraLayerPaths...)
	if err != nil {
		return nil, err
	}

	modelFormat := mediatype.KitFormat
	if opts.UseModelPack {
		modelFormat = mediatype.ModelPackFormat
	}
	comp, err := mediatype.ParseCompression(compression)
	if err != nil {
		return nil, err
	}
	manifestDesc, err := filesystem.SaveModel(ctx, localRepo, kitfile, ign, &filesystem.SaveModelOptions{
		ModelFormat: modelFormat,
		Compression: comp,
		LayerFormat: mediatype.TarFormat,
	})
	if err != nil {
		return nil, err
	}

	if modelRef != nil && modelRef.Reference != "" {
		if err := localRepo.Tag(ctx, *manifestDesc, modelRef.Reference); err != nil {
			return nil, fmt.Errorf("failed to tag manifest: %w", err)
		}
	}
	for _, tag := range extraRefs {
		if err := localRepo.Tag(ctx, *manifestDesc, tag); err != nil {
			return nil, err
		}
	}

	return &PackResult{Descriptor: *manifestDesc}, nil
}

func readKitfile(kitfilePath string) (*artifact.KitFile, error) {
	kitfile := &artifact.KitFile{}
	reader, err := readerForKitfile(kitfilePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if err := kitfile.LoadModel(reader); err != nil {
		return nil, err
	}
	return kitfile, nil
}

func readerForKitfile(kitfilePath string) (io.ReadCloser, error) {
	if kitfilePath == "-" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return os.Stdin, nil
		}
		return nil, fmt.Errorf("no input file specified and no data piped")
	}
	f, err := os.Open(kitfilePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}
