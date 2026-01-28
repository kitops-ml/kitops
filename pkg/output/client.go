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
)

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

func WrapHTTPTransport(rt http.RoundTripper) http.RoundTripper {
	if !logLevel.shouldPrint(LogLevelTrace) {
		return rt
	}
	return &LoggingTransport{rt}
}
