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
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/kitops-ml/kitops/pkg/output"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// callAndPrintError is a wrapper to print an error message for a function that
// may return an error. The error is printed and then discarded.
func callAndPrintError(f func() error, msg string) {
	if err := f(); err != nil {
		output.Errorf(msg, err)
	}
}

func fillDescAnnotations(desc *ocispec.Descriptor, path string, filemeta fs.FileInfo) error {
	if desc.Annotations == nil {
		desc.Annotations = map[string]string{}
	}
	desc.Annotations[modelspecv1.AnnotationFilepath] = path
	if filemeta != nil {
		// This requires an idiosyncratic handling for mode bits -- Mode is _just_ the permission bits in an int32
		// while TypeFlag is _just_ the type byte from the mode
		meta := modelspecv1.FileMetadata{
			Name:    filemeta.Name(),
			Mode:    uint32(filemeta.Mode().Perm()),
			Uid:     0,
			Gid:     0,
			Size:    filemeta.Size(),
			ModTime: filemeta.ModTime(),
			// TODO: This doesn't handle endianess; handling that raises more questions than we're ready to handle anyways.
			Typeflag: byte(filemeta.Mode().Type() >> 24 & 0xFF),
		}
		metabytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal file metadata: %w", err)
		}
		desc.Annotations[modelspecv1.AnnotationFileMetadata] = string(metabytes)
	}
	return nil
}
