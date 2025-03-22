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

//go:build !embed_harness
// +build !embed_harness

package harness

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kitops-ml/kitops/pkg/output"
)

const (
	llamafileDownloadURL = "https://jozu.ml/downloads/?file=llamafile.tar.gz&version=" + LlamaFileVersion
	uiDownloadURL        = "https://jozu.ml/downloads/?file=ui.tar.gz&version=" + LlamaFileVersion
	checksumURL          = "https://jozu.ml/downloads/?file=checksums.txt&version=" + LlamaFileVersion
)

func extractServer(harnessHome string) error {
	if err := os.MkdirAll(harnessHome, os.FileMode(0755)); err != nil {
		return fmt.Errorf("error creating directory %s: %w", harnessHome, err)
	}
	tmpFolder, err := os.MkdirTemp("", "kitops_tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpFolder)

	output.Infoln("downloading harness binaries")
	err = downloadFile(llamafileDownloadURL, tmpFolder, "llamafile.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to extract llamafile: %w", err)
	}
	// Download and verify checksum
	checksums, err := downloadAndParseChecksums(tmpFolder)
	if err != nil {
		return fmt.Errorf("failed to download and parse checksums: %w", err)
	}
	err = verifyChecksum(tmpFolder, "llamafile.tar.gz", checksums)
	if err != nil {
		return fmt.Errorf("checksum verification failed for llamafile.tar.gz: %w", err)
	}

	localFS := os.DirFS(tmpFolder)

	err = extractFile(localFS, "llamafile.tar.gz", harnessHome)
	if err != nil {
		return fmt.Errorf("failed to unpack llamafile: %w", err)
	}

	llamaFilePath := filepath.Join(harnessHome, "llamafile")
	if runtime.GOOS == "windows" {
		llamaExePath := filepath.Join(harnessHome, "llamafile.exe")
		if err := os.Rename(llamaFilePath, llamaExePath); err != nil {
			return fmt.Errorf("error renaming file to executable: %w", err)
		}
	} else {
		if err := os.Chmod(llamaFilePath, os.FileMode(0755)); err != nil {
			return fmt.Errorf("error setting executable permission: %w", err)
		}
	}

	return nil
}

func extractUI(harnessHome string) error {
	tmpFolder, err := os.MkdirTemp("", "kitops_tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpFolder)

	output.Infoln("Updating harness UI components")
	err = downloadFile(uiDownloadURL, tmpFolder, "ui.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to extract UI: %w", err)
	}

	// Download checksum.txt
	checksums, err := downloadAndParseChecksums(tmpFolder)
	if err != nil {
		return fmt.Errorf("failed to download and parse checksums: %w", err)
	}
	// Verify checksum
	err = verifyChecksum(tmpFolder, "ui.tar.gz", checksums)
	if err != nil {
		return fmt.Errorf("checksum verification failed for ui.tar.gz: %w", err)
	}

	uiHome := filepath.Join(harnessHome, "ui")
	if err := os.MkdirAll(uiHome, os.FileMode(0755)); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", uiHome, err)
	}
	localFS := os.DirFS(tmpFolder)

	return extractFile(localFS, "ui.tar.gz", uiHome)
}

func downloadFile(url string, folder string, filename string) error {

	err := os.MkdirAll(folder, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to create folder %s: %w", folder, err)
	}

	filePath := filepath.Join(folder, filename)

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", folder, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download from url %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading file from %s: %s", url, resp.Status)

	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	return nil

}
func downloadAndParseChecksums(tmpFolder string) (map[string]string, error) {
	err := downloadFile(checksumURL, tmpFolder, "checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums.txt: %w", err)
	}

	checksumFilePath := filepath.Join(tmpFolder, "checksums.txt")
	checksums, err := parseChecksumFile(checksumFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse checksums.txt: %w", err)
	}

	return checksums, nil
}

func verifyChecksum(folder, filename string, checksums map[string]string) error {
	expectedChecksum, ok := checksums[filename]
	if !ok {
		return fmt.Errorf("checksum not found for file: %s", filename)
	}

	filePath := filepath.Join(folder, filename)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	hash := sha256.Sum256(fileData)
	computedChecksum := hex.EncodeToString(hash[:])

	if computedChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filename, expectedChecksum, computedChecksum)
	}

	output.Infoln(fmt.Sprintf("Checksum verified for %s", filename))
	return nil
}

func parseChecksumFile(filePath string) (map[string]string, error) {
	checksums := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open checksum file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid checksum line: %s", line)
		}
		checksum := parts[0]
		filename := filepath.Base(parts[1]) // Get the base name of the file (e.g., 'ui.tar.gz')
		checksums[filename] = checksum
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading checksum file: %w", err)
	}

	return checksums, nil
}
