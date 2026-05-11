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

package mlflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	// defaultTrackingURI is used when no tracking URI is embedded in the mlflow:// URI
	// and MLFLOW_TRACKING_URI is not set.
	defaultTrackingURI = "http://localhost:5000"

	// MLFlow REST API paths
	apiGetRun        = "api/2.0/mlflow/runs/get"
	apiListArtifacts = "api/2.0/mlflow/artifacts/list"

	// Run lifecycle states
	RunStatusFinished = "FINISHED"
)

// RunInfo contains metadata for a MLFlow run, as returned by the Tracking API.
type RunInfo struct {
	RunID        string `json:"run_id"`
	RunName      string `json:"run_name"`
	ExperimentID string `json:"experiment_id"`
	Status       string `json:"status"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
}

// RunData contains params and metrics for a run.
type RunData struct {
	Params  []RunParam  `json:"params"`
	Metrics []RunMetric `json:"metrics"`
}

// RunParam is a key-value parameter logged to a run.
type RunParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunMetric is a logged metric value.
type RunMetric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

// Run is the full run object from the MLFlow API.
type Run struct {
	Info RunInfo `json:"info"`
	Data RunData `json:"data"`
}

// ArtifactFileInfo describes a single file or directory in the artifact store.
type ArtifactFileInfo struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}

// Client calls the MLFlow REST Tracking API.
type Client struct {
	trackingURI string
	token       string
	httpClient  *http.Client
}

// NewClient creates a new MLFlow REST client.
// trackingURI should be the base URL of the MLFlow tracking server (e.g. "http://localhost:5000").
// token is optional; if non-empty it is sent as a Bearer token.
func NewClient(trackingURI, token string) *Client {
	return &Client{
		trackingURI: strings.TrimRight(trackingURI, "/"),
		token:       token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRun fetches the run identified by runID from the MLFlow tracking server.
func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	params := url.Values{}
	params.Set("run_id", runID)

	resp, err := c.doGet(ctx, apiGetRun, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp)
	}

	var body struct {
		Run *Run `json:"run"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode MLFlow run response: %w", err)
	}
	if body.Run == nil {
		return nil, fmt.Errorf("MLFlow API returned empty run object for run_id=%s", runID)
	}
	return body.Run, nil
}

// ListArtifacts lists artifact entries under path prefix for the given run.
// Pass an empty artifactPath to list the root artifact directory.
// It recursively expands subdirectories and returns only file entries.
func (c *Client) ListArtifacts(ctx context.Context, runID, artifactPath string) ([]ArtifactFileInfo, error) {
	return c.listArtifactsRecursive(ctx, runID, artifactPath)
}

func (c *Client) listArtifactsRecursive(ctx context.Context, runID, artifactPath string) ([]ArtifactFileInfo, error) {
	params := url.Values{}
	params.Set("run_id", runID)
	if artifactPath != "" {
		params.Set("path", artifactPath)
	}

	resp, err := c.doGet(ctx, apiListArtifacts, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp)
	}

	var body struct {
		Files []ArtifactFileInfo `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode MLFlow artifacts list response: %w", err)
	}

	var results []ArtifactFileInfo
	for _, f := range body.Files {
		if f.IsDir {
			// Recurse into subdirectory
			children, err := c.listArtifactsRecursive(ctx, runID, f.Path)
			if err != nil {
				return nil, err
			}
			results = append(results, children...)
		} else {
			results = append(results, f)
		}
	}
	return results, nil
}

// DownloadArtifactURL returns the URL to download an artifact file.
// This delegates to the tracking server's artifact proxy endpoint.
func (c *Client) DownloadArtifactURL(runID, artifactPath string) (string, error) {
	u, err := url.Parse(c.trackingURI)
	if err != nil {
		return "", fmt.Errorf("invalid tracking URI: %w", err)
	}
	u.Path = path.Join(u.Path, "get-artifact")
	q := u.Query()
	q.Set("run_id", runID)
	q.Set("path", artifactPath)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// doGet executes an authenticated GET request against the MLFlow REST API.
func (c *Client) doGet(ctx context.Context, apiPath string, params url.Values) (*http.Response, error) {
	u, err := url.Parse(c.trackingURI)
	if err != nil {
		return nil, fmt.Errorf("invalid tracking URI: %w", err)
	}
	u.Path = path.Join(u.Path, apiPath)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to MLFlow tracking server failed: %w", err)
	}
	return resp, nil
}

// mlflowErrorResponse is the standard JSON error payload from MLFlow.
type mlflowErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func parseAPIError(resp *http.Response) error {
	var errBody mlflowErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		return fmt.Errorf("MLFlow API error (status %d): unable to parse error body", resp.StatusCode)
	}
	if errBody.Message != "" {
		return fmt.Errorf("MLFlow API error (status %d, code %s): %s", resp.StatusCode, errBody.ErrorCode, errBody.Message)
	}
	return fmt.Errorf("MLFlow API error: status %d", resp.StatusCode)
}
