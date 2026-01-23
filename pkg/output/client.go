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

package output

import (
	"net/http"
	"time"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type LoggingClient struct {
	remote.Client
}

func (c *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := c.Client.Do(req)
	duration := float64(time.Since(start)) / float64(time.Millisecond)
	if err != nil {
		SafeLogf(LogLevelTrace, "%s %s -> ERROR -- duration %.2f ms", req.Method, req.URL, duration)
	} else {
		SafeLogf(LogLevelTrace, "%s %s -> %d -- duration %.2f ms", req.Method, req.URL, resp.StatusCode, duration)
	}
	return resp, err
}

type LoggingTransport struct {
	http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	resp, err := rt.RoundTrip(req)
	duration := float64(time.Since(start)) / float64(time.Millisecond)
	if err != nil {
		SafeLogf(LogLevelTrace, "%s %s -> ERROR -- duration %.2f ms", req.Method, req.URL, duration)
	} else {
		SafeLogf(LogLevelTrace, "%s %s -> %d -- duration %.2f ms", req.Method, req.URL, resp.StatusCode, duration)
	}
	return resp, err
}

// WrapClient returns a remote.Client that logs every request at a 'trace' level.
// If the currently set logging level would not print 'trace' logs, this is a no-op.
func WrapClient(c remote.Client) remote.Client {
	if !logLevel.shouldPrint(LogLevelTrace) {
		return c
	}

	// If the client is an *auth.Client, we need to wrap the underlying http.Client Transport
	if authClient, ok := c.(*auth.Client); ok {
		if authClient.Client == nil {
			authClient.Client = &http.Client{}
		}
		authClient.Client.Transport = &LoggingTransport{
			RoundTripper: authClient.Client.Transport,
		}
		return authClient
	}

	return &LoggingClient{
		Client: c,
	}
}
