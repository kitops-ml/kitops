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

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

func installShellCompletions(rootCmd *cobra.Command, opts *rootOptions) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil
	}

	configHome, err := getConfigHome(opts)
	if err != nil {
		return err
	}

	completionsDir := filepath.Join(configHome, "completions")
	err = os.MkdirAll(completionsDir, 0755)
	if err != nil {
		return err
	}

	shell := detectShell()
	if shell == "" {
		return fmt.Errorf("failed to detect shell, is $SHELL set?")
	}

	completionsFilePath := filepath.Join(completionsDir, "kitops."+shell)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shellRcPath := ""
	switch shell {
	case "bash":
		shellRcPath = filepath.Join(userHomeDir, ".bashrc")
		if err := rootCmd.GenBashCompletionFileV2(completionsFilePath, true); err != nil {
			return err
		}
	case "zsh":
		shellRcPath = filepath.Join(userHomeDir, ".zshrc")
		if err := rootCmd.GenZshCompletionFile(completionsFilePath); err != nil {
			return err
		}
	case "fish":
		shellRcPath = filepath.Join(userHomeDir, ".config", "fish", "config.fish")
		if err := rootCmd.GenFishCompletionFile(completionsFilePath, true); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if err := writeActivationScriptIfMissing(shellRcPath, completionsFilePath); err != nil {
		return err
	}

	return nil
}

func writeActivationScriptIfMissing(pathOfShellRC, pathOfCompletionFile string) error {
	quotedPath := strconv.Quote(pathOfCompletionFile)
	sourceCmd := "source " + quotedPath

	file, err := os.OpenFile(pathOfShellRC, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), sourceCmd) {
		return nil // already present
	}

	if _, err := file.WriteString("\n# Kitops completions\n" + sourceCmd + "\n"); err != nil {
		return err
	}

	output.Infof("Shell completions installed. Run 'source %s' or restart your shell.", pathOfCompletionFile)
	return nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	parts := strings.Split(shell, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
