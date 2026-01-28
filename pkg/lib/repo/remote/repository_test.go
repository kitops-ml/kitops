// Copyright 2026 The KitOps Authors.
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
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type testClient struct {
	responses   []func(*http.Request) (*http.Response, error) // Responses should be either a *http.Response, error or a Do() function
	responseIdx int
}

func (tc *testClient) Do(req *http.Request) (*http.Response, error) {
	if tc.responseIdx < 0 || tc.responseIdx >= len(tc.responses) {
		return nil, fmt.Errorf("test error! invalid request: not in responses slice")
	}
	resp := tc.responses[tc.responseIdx]
	tc.responseIdx++
	return resp(req)
}

func setup(logbuf *bytes.Buffer) (teardown func()) {
	origLogLevel := output.GetLogLevel()
	origOut := output.GetOut()
	origErr := output.GetErr()
	output.SetLogLevel(output.LogLevelTrace)
	output.SetOut(logbuf)
	output.SetErr(logbuf)
	output.SetProgressBars("none")
	return func() {
		output.SetLogLevel(origLogLevel)
		output.SetOut(origOut)
		output.SetErr(origErr)
		output.SetProgressBars("plain")
	}
}

func TestUploadBlobChunked(t *testing.T) {
	var logbuf bytes.Buffer
	teardown := setup(&logbuf)
	defer teardown()

	startUrl, err := url.Parse("http://127.0.0.1/one")
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := uploadChunkDefaultSize*2 + 1<<10
	expectedDigest := ocispec.DescriptorEmptyJSON.Digest
	expectedDesc := ocispec.Descriptor{Digest: expectedDigest, Size: expectedSize}
	// Test content: "upload" in 3 chunks total
	testContent := io.LimitReader(rand.Reader, expectedSize)

	responses := []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/two",
				"Range":    fmt.Sprintf("0-%d", uploadChunkDefaultSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/two"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/three",
				"Range":    fmt.Sprintf("0-%d", 2*uploadChunkDefaultSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/three"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/four",
				"Range":    fmt.Sprintf("0-%d", expectedSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPut, "/four"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusCreated, "", map[string]string{
				"Location": "finalLocation",
			})
		},
	}

	tc := &testClient{
		responses: responses,
	}

	testRepo := Repository{
		Repository:      nil, // Not testing library functionality here!
		Reference:       registry.Reference{Registry: "testreg", Repository: "testrepo", Reference: "testtag"},
		PlainHttp:       true,
		Client:          tc,
		uploadChunkSize: uploadChunkDefaultSize,
	}

	finalLocation, tErr := testRepo.uploadBlobChunked(t.Context(), startUrl, "", expectedDesc, testContent)
	t.Logf("Function output:\n%s\n", logbuf.String())
	if !assert.NoError(t, tErr) {
		return
	}
	assert.Equal(t, "http://127.0.0.1/finalLocation", finalLocation, "Should return location in last response")
}

func TestUploadBlobChunkedVerifyRequestHeaders(t *testing.T) {
	var logbuf bytes.Buffer
	teardown := setup(&logbuf)
	defer teardown()

	startUrl, err := url.Parse("http://127.0.0.1/start")
	if err != nil {
		t.Fatal(err)
	}

	expectedAuthHeader := "test-auth"
	expectedSize := uploadChunkDefaultSize * 3
	expectedDigest := ocispec.DescriptorEmptyJSON.Digest
	expectedDesc := ocispec.Descriptor{Digest: expectedDigest, Size: expectedSize}
	// Test content: "upload" in 3 chunks total
	testContent := io.LimitReader(rand.Reader, expectedSize)
	completedErr := errors.New("test complete")

	responses := []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) {
			processRequest(req, http.MethodPatch, "/start")
			expectedContentRange := fmt.Sprintf("0-%d", uploadChunkDefaultSize-1)
			assert.Equal(t, expectedContentRange, req.Header.Get("Content-Range"), "Content range should match expected")
			assert.Equal(t, "application/octet-stream", req.Header.Get("Content-Type"), "Incorrect content type")
			assert.Equal(t, expectedAuthHeader, req.Header.Get("Authorization"), "Should pass auth header in subrequests")

			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/next",
				"Range":    fmt.Sprintf("0-%d", uploadChunkDefaultSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			processRequest(req, http.MethodPatch, "/next")
			expectedContentRange := fmt.Sprintf("%d-%d", uploadChunkDefaultSize, uploadChunkDefaultSize*2-1)
			assert.Equal(t, expectedContentRange, req.Header.Get("Content-Range"), "Content range should match expected")
			assert.Equal(t, "application/octet-stream", req.Header.Get("Content-Type"), "Incorrect content type")
			assert.Equal(t, expectedAuthHeader, req.Header.Get("Authorization"), "Should pass auth header in subrequests")

			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/next",
				"Range":    fmt.Sprintf("0-%d", 2*uploadChunkDefaultSize-1),
			})
		},
		func(_ *http.Request) (*http.Response, error) {
			// return an error to stop the upload; ensure this error is returned from function
			return nil, completedErr
		},
	}

	tc := &testClient{
		responses: responses,
	}

	testRepo := Repository{
		Repository:      nil, // Not testing library functionality here!
		Reference:       registry.Reference{Registry: "testreg", Repository: "testrepo", Reference: "testtag"},
		PlainHttp:       true,
		Client:          tc,
		uploadChunkSize: uploadChunkDefaultSize,
	}

	_, tErr := testRepo.uploadBlobChunked(t.Context(), startUrl, expectedAuthHeader, expectedDesc, testContent)
	t.Logf("Function output:\n%s\n", logbuf.String())
	assert.ErrorIs(t, tErr, completedErr, "Unexpected error returned")
}

func TestUploadBlobChunkedRetries(t *testing.T) {
	var logbuf bytes.Buffer
	teardown := setup(&logbuf)
	defer teardown()

	retryPolicy = &retry.GenericPolicy{
		Retryable: retry.DefaultPredicate,
		Backoff:   retry.DefaultBackoff,
		MinWait:   100 * time.Millisecond,
		MaxWait:   300 * time.Millisecond,
		MaxRetry:  5,
	}
	defer func() {
		retryPolicy = retry.DefaultPolicy
	}()

	startUrl, err := url.Parse("http://127.0.0.1/one")
	if err != nil {
		t.Fatal(err)
	}

	var testChunkSize int64 = 5

	expectedSize := 3 * testChunkSize
	expectedDigest := ocispec.DescriptorEmptyJSON.Digest
	expectedDesc := ocispec.Descriptor{Digest: expectedDigest, Size: expectedSize}
	// Test content: "upload" in 3 chunks total
	var buf bytes.Buffer
	io.CopyN(&buf, rand.Reader, expectedSize)
	testContent := bytes.NewReader(buf.Bytes())

	responses := []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			// Should retry on HTTP 500
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/two",
				"Range":    fmt.Sprintf("0-%d", testChunkSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/two"); err != nil {
				return nil, err
			}
			// Should retry on HTTP 408
			return makeResponse(req, http.StatusRequestTimeout, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/two"); err != nil {
				return nil, err
			}
			// Should retry on HTTP 429; also tests multiple retries
			return makeResponse(req, http.StatusTooManyRequests, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/two"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/three",
				"Range":    fmt.Sprintf("0-%d", 2*testChunkSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/three"); err != nil {
				return nil, err
			}
			// Should retry on timeout error
			return nil, &net.DNSError{IsTimeout: true}
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/three"); err != nil {
				return nil, err
			}
			// Should retry on other 5xx statuses
			return makeResponse(req, http.StatusServiceUnavailable, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/three"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusAccepted, "", map[string]string{
				"Location": "/four",
				"Range":    fmt.Sprintf("0-%d", expectedSize-1),
			})
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPut, "/four"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusCreated, "", map[string]string{
				"Location": "finalLocation",
			})
		},
	}

	tc := &testClient{
		responses: responses,
	}

	testRepo := Repository{
		Repository:      nil, // Not testing library functionality here!
		Reference:       registry.Reference{Registry: "testreg", Repository: "testrepo", Reference: "testtag"},
		PlainHttp:       true,
		Client:          tc,
		uploadChunkSize: testChunkSize,
	}

	finalLocation, tErr := testRepo.uploadBlobChunked(t.Context(), startUrl, "", expectedDesc, testContent)
	t.Logf("Function output:\n%s\n", logbuf.String())
	if !assert.NoError(t, tErr) {
		return
	}
	assert.Equal(t, "http://127.0.0.1/finalLocation", finalLocation, "Should return location in last response")
}

func TestUploadBlobChunkedRetriesLimit(t *testing.T) {
	var logbuf bytes.Buffer
	teardown := setup(&logbuf)
	defer teardown()

	retryPolicy = &retry.GenericPolicy{
		Retryable: retry.DefaultPredicate,
		Backoff:   retry.DefaultBackoff,
		MinWait:   100 * time.Millisecond,
		MaxWait:   300 * time.Millisecond,
		MaxRetry:  5,
	}
	defer func() {
		retryPolicy = retry.DefaultPolicy
	}()

	startUrl, err := url.Parse("http://127.0.0.1/one")
	if err != nil {
		t.Fatal(err)
	}

	var testChunkSize int64 = 5

	expectedSize := 3 * testChunkSize
	expectedDigest := ocispec.DescriptorEmptyJSON.Digest
	expectedDesc := ocispec.Descriptor{Digest: expectedDigest, Size: expectedSize}
	// Test content: "upload" in 3 chunks total
	var buf bytes.Buffer
	io.CopyN(&buf, rand.Reader, expectedSize)
	testContent := bytes.NewReader(buf.Bytes())

	responses := []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusInternalServerError, "", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			if err := processRequest(req, http.MethodPatch, "/one"); err != nil {
				return nil, err
			}
			return makeResponse(req, http.StatusServiceUnavailable, "end of test", nil)
		},
		func(req *http.Request) (*http.Response, error) {
			// Should not reach this point, so this request should never be processed
			return nil, fmt.Errorf("Unexpected request")
		},
	}

	tc := &testClient{
		responses: responses,
	}

	testRepo := Repository{
		Repository:      nil, // Not testing library functionality here!
		Reference:       registry.Reference{Registry: "testreg", Repository: "testrepo", Reference: "testtag"},
		PlainHttp:       true,
		Client:          tc,
		uploadChunkSize: testChunkSize,
	}

	_, tErr := testRepo.uploadBlobChunked(t.Context(), startUrl, "", expectedDesc, testContent)
	t.Logf("Function output:\n%s\n", logbuf.String())
	if !assert.Error(t, tErr, "Expected an error to be returned") {
		return
	}
	assert.ErrorContains(t, tErr, "end of test", "Unexpected error returned")
}

func gobbleBody(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	defer req.Body.Close()
	n, err := io.Copy(io.Discard, req.Body)
	if err != nil {
		return err
	}
	if n != req.ContentLength {
		return fmt.Errorf("request body length does not match content-length: %d vs %d", n, req.ContentLength)
	}
	return nil
}

func processRequest(req *http.Request, expectMethod, expectPath string) error {
	if err := gobbleBody(req); err != nil {
		return err
	}
	if req.Method != expectMethod {
		return fmt.Errorf("expected %s method", expectMethod)
	}
	if req.URL.Path != expectPath {
		return fmt.Errorf("Unexpected path in request: expected '%s' but got '%s'", expectPath, req.URL.Path)
	}
	return nil
}

func makeResponse(req *http.Request, status int, body string, headers map[string]string) (*http.Response, error) {
	actualHeaders := map[string][]string{}
	for k, v := range headers {
		actualHeaders[k] = []string{v}
	}
	return &http.Response{
		Request:    req,
		StatusCode: status,
		Header:     actualHeaders,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}, nil
}
