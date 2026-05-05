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

package hf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	modelRevisionURLFmt   = "https://huggingface.co/api/models/%s/revision/%s"
	datasetRevisionURLFmt = "https://huggingface.co/api/datasets/%s/revision/%s"
)

// PinCommit resolves ref to an immutable commit SHA via HF's revision
// metadata endpoint. HF accepts commit SHAs anywhere a ref is expected, so
// callers can thread the returned SHA through ListFiles and DownloadFiles to
// bind the entire import to one snapshot even if the branch moves mid-run.
func PinCommit(ctx context.Context, repo, ref, token string, repoType RepositoryType) (string, error) {
	var apiURL string
	if repoType == RepoTypeDataset {
		apiURL = fmt.Sprintf(datasetRevisionURLFmt, repo, ref)
	} else {
		apiURL = fmt.Sprintf(modelRevisionURLFmt, repo, ref)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("revision lookup got status %d", resp.StatusCode)
	}

	var info struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("parsing revision response: %w", err)
	}
	if info.SHA == "" {
		return "", fmt.Errorf("revision response missing sha for %s@%s", repo, ref)
	}
	return info.SHA, nil
}
