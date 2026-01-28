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
	"regexp"
	"strings"

	"github.com/kitops-ml/kitops/pkg/output"
)

type uploadFormat int

const (
	uploadMonolithicPut uploadFormat = iota
	uploadChunkedPatch
	uploadUndefined
)

const (
	uploadChunkDefaultSize int64 = 100 << 20
)

var (
	googleArtifactRegistryRegexp         = regexp.MustCompile(`.*\.pkg\.dev$`)
	googleContainerRegistryRegexp        = regexp.MustCompile(`.*\.?gcr\.io$`)
	amazonElasticContainerRegistryRegexp = regexp.MustCompile(`.*\.?amazonaws\.com(\.cn)?$`)
)

func getUploadFormat(registry string, size int64, chunkSize int64) uploadFormat {
	output.SafeDebugf("Getting upload format for: %s", registry)
	registry = strings.ToLower(registry)
	switch {
	case registry == "ghcr.io":
		// ghcr.io returns 416 is a PATCH has Content-Length greater than 4.0 MiB for some reason
		// Transfer-Encoding: chunked is supported by the registry, but not implemented yet.
		return uploadMonolithicPut
	case googleArtifactRegistryRegexp.MatchString(registry) || googleContainerRegistryRegexp.MatchString(registry):
		// Google Artifact Registry does not support chunked uploads and instead requires monolithic
		// uploads.
		// docs: https://cloud.google.com/artifact-registry/docs/docker/pushing-and-pulling#pushing
		return uploadMonolithicPut
	case amazonElasticContainerRegistryRegexp.MatchString(registry):
		// Amazon ECR returns an HTTP 201 after the first piece in a multi-part upload, which causes
		// uploads to fail.
		return uploadMonolithicPut
	default:
		// No matches above, use heuristic
		if size < chunkSize {
			return uploadMonolithicPut
		} else {
			return uploadChunkedPatch
		}
	}
}
