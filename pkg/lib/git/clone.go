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

package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/kitops-ml/kitops/pkg/output"
)

const DefaultGitRef = "main"

func CloneRepository(repo, repoRef, dest, token string) error {
	if err := checkGit(); err != nil {
		return err
	}
	if err := checkDestination(dest); err != nil {
		return err
	}

	cloneEnv := os.Environ()
	if token != "" {
		// Set up additional git config settings without overriding user's actual gitconfig.
		// Git documentation for this "feature": https://git-scm.com/docs/git-config#ENVIRONMENT
		cloneEnv = append(cloneEnv, fmt.Sprintf(`KIT_IMPORT_TOKEN=%s`, token))
		cloneEnv = append(cloneEnv,
			"GIT_CONFIG_COUNT=2",
			"GIT_CONFIG_KEY_0=credential.username",
			"GIT_CONFIG_VALUE_0=token",
			"GIT_CONFIG_KEY_1=credential.helper",
			`GIT_CONFIG_VALUE_1=!f() { test "$1" = get && echo "password=${KIT_IMPORT_TOKEN}"; }; f`,
		)
	} else {
		// This is a workaround to disable interactive password prompts
		cloneEnv = append(cloneEnv, "GIT_ASKPASS=true")
		cloneEnv = append(cloneEnv, "GIT_TERMINAL_PROMPT=0")
	}

	// Clone without LFS enabled to get metadata about repo first
	var cloneCmd *exec.Cmd
	if repoRef != DefaultGitRef {
		cloneCmd = exec.Command("git", "clone", "--depth", "1", "--branch", repoRef, "-c", "advice.detachedHead=false", repo, dest)
	} else {
		cloneCmd = exec.Command("git", "clone", "--depth", "1", repo, dest)
	}
	cloneCmd.Env = cloneEnv
	cloneCmd.Env = append(cloneCmd.Env, "GIT_LFS_SKIP_SMUDGE=1")
	cloneCmd.Stderr = os.Stderr
	cloneCmd.Stdout = os.Stdout

	output.Infof("Cloning repository %s", repo)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %w", err)
	}

	// Pull LFS files
	lfsCmd := exec.Command("git", "lfs", "pull")
	lfsCmd.Dir = dest
	lfsCmd.Env = cloneEnv
	lfsCmd.Stdout = os.Stdout
	lfsCmd.Stderr = os.Stderr
	output.Infof("Pulling large files")
	if err := lfsCmd.Run(); err != nil {
		return fmt.Errorf("error downloading LFS files: %w", err)
	}
	// LFS pull prints progress by overwriting a line repeatedly; once it's done
	// we need to print a newline to avoid overwriting the last line
	fmt.Printf("\n")

	return nil
}

func checkGit() error {
	gitCmd := exec.Command("git", "version")
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git is not installed")
	}

	gitLFSCmd := exec.Command("git", "lfs", "version")
	if err := gitLFSCmd.Run(); err != nil {
		return fmt.Errorf("git-lfs is not installed")
	}

	return nil
}

func checkDestination(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("path %s exists and is not a directory", path)
	}
	// TODO: probably don't need to read the _whole_ directory
	contents, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to inspect directory %s: %w", path, err)
	}
	if len(contents) > 0 {
		return fmt.Errorf("cannot clone to a non-empty directory")
	}
	return nil
}
