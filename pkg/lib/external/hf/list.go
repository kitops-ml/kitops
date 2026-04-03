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
	"net/url"
	"path"
	"strings"
	"time"

	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	"github.com/kitops-ml/kitops/pkg/output"
)

const (
	modelTreeURLFmt   = "https://huggingface.co/api/models/%s/tree/%s"
	datasetTreeURLFmt = "https://huggingface.co/api/datasets/%s/tree/%s"
)

type hfTreeResponse []struct {
	Type string `json:"type"`
	OID  string `json:"oid"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type hfErrorResponse struct {
	Error string `json:"error"`
}

func ListFiles(ctx context.Context, modelRepo, ref string, token string, repoType RepositoryType) (*kfgen.DirectoryListing, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var treeURL string
	if repoType == RepoTypeDataset {
		treeURL = fmt.Sprintf(datasetTreeURLFmt, modelRepo, ref)
	} else {
		treeURL = fmt.Sprintf(modelTreeURLFmt, modelRepo, ref)
	}

	baseURL, err := url.Parse(treeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return walkRepoTree(ctx, client, token, baseURL, ".")
}

func walkRepoTree(ctx context.Context, client *http.Client, token string, repoBaseUrl *url.URL, subDir string) (*kfgen.DirectoryListing, error) {
	curUrl := *repoBaseUrl
	curUrl.Path = path.Join(curUrl.Path, subDir)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, curUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	repoTree, err := processTreeResponse(resp)
	if closeErr := resp.Body.Close(); closeErr != nil {
		output.Logf(output.LogLevelWarn, "failed to close response body: %s", closeErr)
	}
	if err != nil {
		return nil, err
	}

	dirListing := &kfgen.DirectoryListing{
		Name: path.Base(subDir),
		Path: subDir,
	}
	for _, elem := range *repoTree {
		switch elem.Type {
		case "directory":
			subDirListing, err := walkRepoTree(ctx, client, token, repoBaseUrl, elem.Path)
			if err != nil {
				return nil, err
			}
			dirListing.Subdirs = append(dirListing.Subdirs, *subDirListing)
		case "file":
			name := path.Base(elem.Path)
			if name == ".gitignore" || name == ".gitattributes" {
				continue
			}
			dirListing.Files = append(dirListing.Files, kfgen.FileListing{
				Name: name,
				Path: elem.Path,
				Size: elem.Size,
			})
		default:
			return nil, fmt.Errorf("unknown type in repository tree: %s", elem.Type)
		}
	}

	return dirListing, nil
}

func processTreeResponse(resp *http.Response) (*hfTreeResponse, error) {
	if resp.StatusCode != http.StatusOK {
		errResp := &hfErrorResponse{}
		if err := json.NewDecoder(resp.Body).Decode(errResp); err != nil {
			return nil, fmt.Errorf("failed to parse API error response: %w", err)
		}
		if resp.StatusCode == http.StatusNotFound && strings.HasPrefix(errResp.Error, "Invalid rev id") {
			// Handle case where provided reference (tag) is not found to avoid confusing response
			ref := errResp.Error[strings.LastIndex(errResp.Error, " ")+1:]
			return nil, fmt.Errorf("reference '%s' not found", ref)
		}
		return nil, fmt.Errorf("got error code %d from API: %s", resp.StatusCode, errResp.Error)
	}

	repoTree := &hfTreeResponse{}
	if err := json.NewDecoder(resp.Body).Decode(repoTree); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	return repoTree, nil
}
