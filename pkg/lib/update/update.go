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

package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"golang.org/x/mod/semver"
)

const versionCheckUrl = "https://kitops.org/version"

// Regexp for a semver version -- taken from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
// We've added an optional 'v' to the start (e.g. v1.2.3) since using a 'v' prefix is common (and used, in our case)
// Capture groups are:
//
//	[1] - Major version
//	[2] - Minor version
//	[3] - Bugfix/z-stream version
//	[4] - Pre-release identifiers (1.2.3-<info>), if present
//	[5] - Build metadata (1.2.3+<metadata>), if present
var versionTagRegexp = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type ghReleaseInfo struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	Url        string `json:"html_url"`
}

func CheckForUpdate(configHome, command string) {
	// If this isn't a release version of kit, don't nag the user unnecessarily
	if constants.Version == "unknown" || !versionTagRegexp.MatchString(constants.Version) {
		return
	}
	if !shouldShowNotification(configHome) {
		return
	}

	info, err := getLatestReleaseInfo(command)
	if err != nil {
		output.Debugf("Error checking for CLI updates: %s", err)
		return
	}
	if info.Prerelease || info.Draft {
		// This isn't a full release; for now just don't notify users, even if there is a newer full release we don't know about
		return
	}

	// The Go semver package requires versions start with a 'v' (contrary to the spec)
	currentVersion := fmt.Sprintf("v%s", strings.TrimPrefix(constants.Version, "v"))
	latestVersion := fmt.Sprintf("v%s", strings.TrimPrefix(info.TagName, "v"))
	if semver.Compare(currentVersion, latestVersion) < 0 {
		output.SystemInfof("Note: A new version of Kit is available! You are using Kit %s. The latest version is %s.", currentVersion, latestVersion)
		output.SystemInfof("      To see a list of changes, visit %s", info.Url)
		output.SystemInfof("      To disable this notification, use 'kit version --show-update-notifications=false'")
		output.SystemInfof("") // Add a newline to not confuse it with regular output
	}
}

func SetShowNotifications(configHome string, shouldShow bool) error {
	flagFile := filepath.Join(configHome, constants.UpdateNotificationsConfigFilename)
	if shouldShow {
		if err := os.Remove(flagFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("error enabling update notifications: %w", err)
		}
	} else {
		f, err := os.Create(flagFile)
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				return nil
			}
			return fmt.Errorf("error disabling update notifications: %w", err)
		}
		f.Close()
	}
	return nil
}

func shouldShowNotification(configHome string) bool {
	flagFile := filepath.Join(configHome, constants.UpdateNotificationsConfigFilename)
	_, err := os.Stat(flagFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true
		}
		output.Debugf("Error checking if update notifications should be shown: %s", err)
	}
	return false
}

func getLatestReleaseInfo(command string) (*ghReleaseInfo, error) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, versionCheckUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error while generating HTTP request: %w", err)
	}
	req.Header.Set("User-Agent", "kitops-cli/"+constants.Version)
	req.Header.Set("X-Command", command)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GitHub response body: %w", err)
	}
	info := &ghReleaseInfo{}
	if err := json.Unmarshal(respBody, info); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response body: %w", err)
	}
	return info, nil
}
