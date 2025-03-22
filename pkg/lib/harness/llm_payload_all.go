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

package harness

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
)

func extractFile(fs fs.FS, file, harnessHome string) (err error) {
	srcFile, err := fs.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", file, err)
	}
	defer srcFile.Close()

	srcReader := io.Reader(srcFile)
	destFileName := file

	if strings.HasSuffix(file, ".tar.gz") {
		gzr, err := gzip.NewReader(srcReader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader for %s: %w", file, err)
		}
		defer gzr.Close()
		tarReader := tar.NewReader(gzr)
		return extractTar(tarReader, harnessHome)
	}

	if strings.HasSuffix(file, ".gz") {
		srcReader, err = gzip.NewReader(srcReader)
		if err != nil {
			return fmt.Errorf("failed to decompress payload %s: %w", file, err)
		}
		destFileName = strings.TrimSuffix(file, ".gz")
	}

	destFile := filepath.Join(harnessHome, filepath.Base(destFileName))
	dest, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0755)) // Keep executable permissions
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destFile, err)
	}
	defer closeFileErrCheck(dest, &err, "failed to close destination file")

	if _, err := io.Copy(dest, srcReader); err != nil {
		return fmt.Errorf("failed to copy payload to %s: %w", destFile, err)
	}
	return nil
}

func extractTar(tr *tar.Reader, dir string) (err error) {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}
		// Sanitize the file name to prevent path traversal
		sanitizedName := filepath.Clean(header.Name)
		if strings.Contains(sanitizedName, "..") || filepath.IsAbs(sanitizedName) {
			return fmt.Errorf("invalid file path in archive: %s", sanitizedName)
		}
		outPath := filepath.Join(dir, sanitizedName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", outPath, err)
			}

		case tar.TypeReg:
			if fi, exists := filesystem.PathExists(outPath); exists {
				if !fi.Mode().IsRegular() {
					return fmt.Errorf("path '%s' already exists and is not a regular file", outPath)
				}
			}
			file, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outPath, err)
			}
			defer closeFileErrCheck(file, &err, "failed to close gzip reader")

			written, err := io.Copy(file, tr)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", outPath, err)
			}
			if written != header.Size {
				return fmt.Errorf("could not unpack file %s", outPath)
			}

		default:
			return fmt.Errorf("unrecognized type in archive: %s", header.Name)
		}
	}
	return nil
}

func closeFileErrCheck(f *os.File, err *error, msg string) {
	if cerr := f.Close(); cerr != nil {
		if *err == nil {
			*err = fmt.Errorf("%s: %w", msg, cerr)
		} else {
			*err = fmt.Errorf("%v; %s: %w", *err, msg, cerr)
		}
	}
}
