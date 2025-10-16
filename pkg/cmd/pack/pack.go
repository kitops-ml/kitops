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

package pack

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/ignore"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// runPack compresses and stores a modelkit based on a Kitfile. Returns an error if packing
// fails for any reason, or if any path in the Kitfile is not a subdirectory of the current
// context directory.
//
// Packed modelkits are saved to the local on-disk cache. As OCI-spec indexes only support one
// registry/repository reference at a time, individual blobs may be duplicated on disk if stored
// under different references.
func runPack(ctx context.Context, options *packOptions) error {
	kitfile, err := readKitfile(options.modelFile)
	if err != nil {
		return err
	}

	storageHome := constants.StoragePath(options.configHome)
	localRepo, err := local.NewLocalRepo(storageHome, options.modelRef)
	if err != nil {
		return fmt.Errorf("failed to open local storage: %w", err)
	}

	manifestDesc, err := pack(ctx, options, kitfile, localRepo)
	if err != nil {
		return err
	}

	if options.modelRef != nil && options.modelRef.Reference != "" {
		if err := localRepo.Tag(ctx, *manifestDesc, options.modelRef.Reference); err != nil {
			return fmt.Errorf("failed to tag manifest: %w", err)
		}
		output.Debugf("Added tag to manifest: %s", options.modelRef.Reference)
	}

	for _, tag := range options.extraRefs {
		if err := localRepo.Tag(ctx, *manifestDesc, tag); err != nil {
			return err
		}
	}

	output.Infof("Model saved: %s", manifestDesc.Digest)

	return nil
}

func pack(ctx context.Context, opts *packOptions, kitfile *artifact.KitFile, localRepo local.LocalRepo) (*ocispec.Descriptor, error) {
	var extraLayerPaths []string
	if kitfile.Model != nil && util.IsModelKitReference(kitfile.Model.Path) {
		baseRef := util.FormatRepositoryForDisplay(opts.modelRef.String())
		parentKitfile, err := kfutils.ResolveKitfile(ctx, opts.configHome, kitfile.Model.Path, baseRef)
		if err != nil {
			return nil, fmt.Errorf("Failed to resolve referenced modelkit %s: %w", kitfile.Model.Path, err)
		}
		extraLayerPaths = util.LayerPathsFromKitfile(parentKitfile)
	}

	ignore, err := ignore.NewFromContext(opts.contextDir, kitfile, extraLayerPaths...)
	if err != nil {
		return nil, err
	}

	compression, err := mediatype.ParseCompression(opts.compression)
	if err != nil {
		return nil, err
	}
	manifestDesc, err := filesystem.SaveModel(ctx, localRepo, kitfile, ignore, compression)
	if err != nil {
		return nil, err
	}
	return manifestDesc, nil
}

func readKitfile(modelFile string) (*artifact.KitFile, error) {
	// 1. Read the model file
	kitfile := &artifact.KitFile{}
	kitfileContentReader, err := readerForKitfile(modelFile)
	if err != nil {
		return nil, err
	}
	defer kitfileContentReader.Close()
	if err := kitfile.LoadModel(kitfileContentReader); err != nil {
		return nil, err
	}
	if err := kfutils.ValidateKitfile(kitfile); err != nil {
		return nil, err
	}
	return kitfile, nil
}

// readerForKitfile returns a reader for the Kitfile specified by the modelFile argument.
// If modelFile is "-", the function returns a reader for stdin. If modelFile is a file path,
// the function returns a reader for the file. If the file cannot be opened, the function returns
// an error.
// it is up-to the caller to close the reader.
func readerForKitfile(modelFile string) (io.ReadCloser, error) {
	var modelfile io.ReadCloser
	if modelFile == "-" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			modelfile = os.Stdin
		} else {
			return nil, fmt.Errorf("No input file specified and no data piped")
		}
	} else {
		var err error
		modelfile, err = os.Open(modelFile)
		if err != nil {
			return nil, err
		}
	}
	return modelfile, nil
}
