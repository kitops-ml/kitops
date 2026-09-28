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

package filesystem

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/output"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

// saveFileToRepo saves the content at filePath to target under desc. If target is a local.LocalRepo,
// the file is moved directly into local storage to avoid copying a potentially very large file;
// otherwise (e.g. a remote registry target), the file is streamed to the target via Push.
func saveFileToRepo(ctx context.Context, filePath string, desc ocispec.Descriptor, target oras.Target) error {
	mt, err := mediatype.ParseMediaType(desc.MediaType)
	if err != nil {
		return err
	}

	if exists, err := target.Exists(ctx, desc); err != nil {
		return err
	} else if exists {
		output.Infof("Already saved %s layer: %s", mt.UserString(), desc.Digest)
		return nil
	}

	if localRepo, ok := target.(local.LocalRepo); ok {
		// Workaround to avoid copying a potentially very large file: move it to the expected path
		// and verify that it exists afterwards.
		if err := localRepo.EnsureDirs(desc); err != nil {
			return err
		}
		blobPath := localRepo.BlobPath(desc)
		if err := os.Rename(filePath, blobPath); err != nil {
			// This may fail on some systems (e.g. linux where / and /home are different partitions)
			// Fallback to regular push which is basically a copy
			output.Debugf("Failed to move temp file into storage (will copy instead): %s", err)
			if err := pushFileToTarget(ctx, filePath, desc, target); err != nil {
				return fmt.Errorf("failed to add layer to storage: %w", err)
			}
		}
	} else if err := pushFileToTarget(ctx, filePath, desc, target); err != nil {
		return fmt.Errorf("failed to push layer: %w", err)
	}

	// Verify blob is in store now
	exists, err := target.Exists(ctx, desc)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("failed to save layer to target: blob is not stored")
	}
	output.Infof("Saved %s layer: %s", mt.UserString(), desc.Digest)
	return nil
}

func pushFileToTarget(ctx context.Context, filePath string, desc ocispec.Descriptor, target oras.Target) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer file.Close()
	return target.Push(ctx, desc, file)
}

func fillDescAnnotations(desc *ocispec.Descriptor, targetPath string, fi fs.FileInfo) error {
	if desc.Annotations == nil {
		desc.Annotations = map[string]string{}
	}

	desc.Annotations[modelspecv1.AnnotationFilepath] = filepath.ToSlash(targetPath)
	if fi != nil {
		tarHeader, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return fmt.Errorf("error processing file metadata: %w", err)
		}
		// This requires an idiosyncratic handling for mode bits -- Mode is _just_ the permission bits in an int32
		// while TypeFlag is _just_ a "type" byte
		meta := modelspecv1.FileMetadata{
			Name:    fi.Name(),
			Mode:    uint32(fi.Mode().Perm()),
			Uid:     0,
			Gid:     0,
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
			// It's not currently specified what this byte should be; for now use the typeflag from tar headers
			Typeflag: tarHeader.Typeflag,
		}
		metabytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal file metadata: %w", err)
		}
		desc.Annotations[modelspecv1.AnnotationFileMetadata] = string(metabytes)
	}
	return nil
}

// closeAll closes all Closers one by one, returning any errors that occure while closing. To support a chain of closers
// that is used for a write operation, the list of io.Closers is closed in reverse order, assuming that the list is in
// the order the Closers were created.
func closeAll(toClose []io.Closer) error {
	var errs error
	for _, tc := range slices.Backward(toClose) {
		errs = errors.Join(errs, tc.Close())
	}
	return errs
}
