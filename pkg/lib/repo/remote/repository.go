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

package remote

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type Repository struct {
	registry.Repository
	Reference registry.Reference
	PlainHttp bool
	Client    remote.Client
}

func (r *Repository) Untag(ctx context.Context, reference string) error {
	if err := r.Reference.ValidateReferenceAsDigest(); err == nil {
		return fmt.Errorf("cannot untag using digest")
	}
	ctx = auth.AppendRepositoryScope(ctx, r.Reference, auth.ActionPull, auth.ActionDelete)
	apiURL := buildRepositoryManifestsURL(r.PlainHttp, r.Reference, reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL, nil)
	if err != nil {
		return fmt.Errorf("error generating request: %w", err)
	}
	resp, err := r.client().Do(req)
	if err != nil {
		return fmt.Errorf("failed to untag: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusMethodNotAllowed:
		return fmt.Errorf("remote registry does not support untagging")
	case http.StatusNotFound:
		return fmt.Errorf("reference %s not found in remote registry", reference)
	case http.StatusAccepted:
		return nil
	default:
		return fmt.Errorf("unexpected response code from remote registry: %s", resp.Status)
	}
}

// Push pushes the content, matching the expected descriptor.
func (r *Repository) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	if expected.MediaType == ocispec.MediaTypeImageManifest {
		// If it's a manifest, we can just use the regular implementation
		return r.Repository.Push(ctx, expected, content)
	}

	// Otherwise, push a blob according to the OCI spec
	ctx = auth.AppendRepositoryScope(ctx, r.Reference, auth.ActionPull, auth.ActionPush)
	sessionURL, postResp, err := r.initiateUploadSession(ctx)
	if err != nil {
		return err
	}

	blobUrl, err := r.uploadBlob(ctx, sessionURL, postResp, expected, content)
	if err != nil {
		return err
	}
	output.SafeDebugf("Blob uploaded, available at url %s", blobUrl)

	return nil
}

func (r *Repository) initiateUploadSession(ctx context.Context) (*url.URL, *http.Response, error) {
	uploadUrl := buildRepositoryBlobUploadURL(r.PlainHttp, r.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadUrl, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Handle warnings from remote
	// References:
	//   - https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc4/spec.md#warnings
	//   - https://www.rfc-editor.org/rfc/rfc7234#section-5.5
	resp, err := r.client().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initiate upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return nil, nil, handleRemoteError(resp)
	}
	location, err := resp.Location()
	if err != nil {
		return nil, nil, fmt.Errorf("registry did not respond with upload location")
	}

	// Workaround for https://github.com/oras-project/oras-go/issues/177 -- sometimes
	// location header does not include port, causing auth client to mismatch the context
	locationHostname := location.Hostname()
	locationPort := location.Port()
	origHostname := req.URL.Hostname()
	origPort := req.URL.Port()
	if origPort == "443" && locationHostname == origHostname && locationPort == "" {
		location.Host = locationHostname + ":" + origPort
	}
	output.SafeDebugf("Using location %s for blob upload", path.Join(location.Hostname(), location.Path))

	return location, resp, nil
}

func (r *Repository) uploadBlob(ctx context.Context, location *url.URL, postResp *http.Response, expected ocispec.Descriptor, content io.Reader) (string, error) {
	output.SafeDebugf("Size: %d", expected.Size)
	uploadFormat := getUploadFormat(location.Hostname(), expected.Size)
	switch uploadFormat {
	case uploadMonolithicPut:
		return r.uploadBlobMonolithic(ctx, location, postResp, expected, content)
	case uploadChunkedPatch:
		return r.uploadBlobChunked(ctx, location, postResp, expected, content)
	default:
		return "", fmt.Errorf("unknown registry %s, cannot upload", location.Hostname())
	}
}

// uploadBlobMonolithic performs a monolithic blob upload as per the distribution spec. The content of the blob is uploaded
// in one PUT request at the provided location.
func (r *Repository) uploadBlobMonolithic(ctx context.Context, location *url.URL, postResp *http.Response, expected ocispec.Descriptor, content io.Reader) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, location.String(), content)
	if err != nil {
		return "", err
	}
	// Set Content-Length header
	if req.GetBody != nil && req.ContentLength != expected.Size {
		// short circuit a size mismatch for built-in types.
		return "", fmt.Errorf("mismatch content length %d: expect %d", req.ContentLength, expected.Size)
	}
	req.ContentLength = expected.Size

	// Set Content-Type to required 'application/octet-stream'
	req.Header.Set("Content-Type", "application/octet-stream")

	// Set digest query to mark this as completing the upload
	q := req.URL.Query()
	q.Set("digest", expected.Digest.String())
	req.URL.RawQuery = q.Encode()

	// Reuse credentials from POST request that initiated upload
	if auth := postResp.Request.Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	output.SafeDebugf("Uploading blob as one chunk")
	// TODO: Handle warnings from remote
	// References:
	//   - https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc4/spec.md#warnings
	//   - https://www.rfc-editor.org/rfc/rfc7234#section-5.5
	resp, err := r.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", handleRemoteError(resp)
	}

	blobLocation, err := resp.Location()
	if err != nil {
		output.Errorf("Warning: remote registry did not return blob location")
	}

	return blobLocation.String(), nil
}

// uploadBlobChunked performs a chunked blob upload as per the distribution spec. The blob is divided into chunks of maximum 100MiB
// in size and uploaded sequentially through PATCH requests. Once entire blob is uploaded, a PUT request marks the upload as complete.
// Note that the distribution spec 1) requires blobs to uploaded in-order, and 2) does not have a way of specifying maximum blob
// size.
func (r *Repository) uploadBlobChunked(ctx context.Context, location *url.URL, postResp *http.Response, expected ocispec.Descriptor, content io.Reader) (string, error) {
	// TODO: Handle 'OCI-Chunk-Min-Length' header in post response
	numChunks := int(math.Ceil(float64(expected.Size) / float64(uploadChunkDefaultSize)))
	authHeader := postResp.Request.Header.Get("Authorization")

	rangeStart := int64(0)
	rangeEnd := min(uploadChunkDefaultSize-1, expected.Size-1)
	nextLocation := location
	for i := 0; i < numChunks; i++ {
		output.SafeDebugf("Uploading chunk %d/%d, range %d-%d", i+1, numChunks, rangeStart, rangeEnd)

		bodyLength := rangeEnd - rangeStart + 1
		lr := io.LimitReader(content, int64(bodyLength))

		// Set up request reading from the LimitReader
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, nextLocation.String(), lr)
		if err != nil {
			return "", err
		}
		req.ContentLength = bodyLength
		req.Header.Set("Content-Range", fmt.Sprintf("%d-%d", rangeStart, rangeEnd))
		req.Header.Set("Content-Type", "application/octet-stream")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		// Submit the chunk as a PATCH
		// TODO: Handle 416 response code (range not satisfiable)
		resp, err := r.client().Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to upload blob chunk: %w", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			defer resp.Body.Close()
			return "", handleRemoteError(resp)
		}
		resp.Body.Close()

		// Parse and verify data out of response
		// Location should be the next upload location
		respLocation, err := resp.Location()
		if err != nil {
			return "", fmt.Errorf("missing Location header in response")
		}
		nextLocation = respLocation

		// Verify Range header in response matches what we expect
		respRange := resp.Header.Get("Range")
		if respRange == "" {
			return "", fmt.Errorf("missing Range header in response")
		}
		startEnd := strings.Split(respRange, "-")
		if len(startEnd) != 2 || startEnd[0] != "0" {
			return "", fmt.Errorf("server returned invalid Range header: %s", respRange)
		}
		curEnd, err := strconv.ParseInt(startEnd[1], 10, 0)
		if err != nil {
			return "", fmt.Errorf("server returned invalid Range header: %s", respRange)
		}
		if curEnd != rangeEnd {
			return "", fmt.Errorf("mismatch in range header: expected 0-%d, actual 0-%d", rangeEnd, curEnd)
		}

		// Prepare next range
		rangeStart = rangeEnd + 1
		rangeEnd = min(expected.Size-1, rangeEnd+uploadChunkDefaultSize)
	}

	// Final PUT request to mark upload as completed for server. Note that the final chunk _could_ be included in this
	// PUT but isn't for simplicity
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, nextLocation.String(), nil)
	if err != nil {
		return "", err
	}
	// Set digest query to mark this as completing the upload
	q := req.URL.Query()
	q.Set("digest", expected.Digest.String())
	req.URL.RawQuery = q.Encode()
	// Reuse credentials from POST request that initiated upload
	if auth := postResp.Request.Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	output.SafeDebugf("Finalizing upload")
	resp, err := r.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to finalize blob upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", handleRemoteError(resp)
	}

	blobLocation, err := resp.Location()
	if err != nil {
		output.Errorf("Warning: remote registry did not return blob location")
	}

	return blobLocation.String(), nil
}

// client returns an HTTP client used to access the remote repository.
// A default HTTP client is return if the client is not configured.
func (r *Repository) client() remote.Client {
	if r.Client == nil {
		return auth.DefaultClient
	}
	return r.Client
}

func buildRepositoryBlobUploadURL(plainHTTP bool, ref registry.Reference) string {
	scheme := "https"
	if plainHTTP {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/", scheme, ref.Host(), ref.Repository)
}

func buildRepositoryManifestsURL(plainHTTP bool, registryRef registry.Reference, manifestRef string) string {
	scheme := "https"
	if plainHTTP {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/v2/%s/manifests/%s", scheme, registryRef.Host(), registryRef.Repository, manifestRef)
}
