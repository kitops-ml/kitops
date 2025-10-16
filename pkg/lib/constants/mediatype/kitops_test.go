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

package mediatype

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKitopsMediaType(t *testing.T) {
	mediaTypes := []string{
		"application/vnd.kitops.modelkit.config.v1+json",
		"application/vnd.kitops.modelkit.model.v1.tar",
		"application/vnd.kitops.modelkit.model.v1.tar+gzip",
		"application/vnd.kitops.modelkit.modelpart.v1.tar",
		"application/vnd.kitops.modelkit.modelpart.v1.tar+gzip",
		"application/vnd.kitops.modelkit.dataset.v1.tar",
		"application/vnd.kitops.modelkit.dataset.v1.tar+gzip",
		"application/vnd.kitops.modelkit.code.v1.tar",
		"application/vnd.kitops.modelkit.code.v1.tar+gzip",
		"application/vnd.kitops.modelkit.docs.v1.tar",
		"application/vnd.kitops.modelkit.docs.v1.tar+gzip",
	}

	for _, mediaType := range mediaTypes {
		t.Run(fmt.Sprintf("Parsing %s", mediaType), func(t *testing.T) {
			parsedType, err := ParseMediaType(mediaType)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, mediaType, parsedType.String(), "Parsed media type should match input")
		})
	}
}

func TestParseKitopsInvalidType(t *testing.T) {
	tests := []struct {
		mediaType string
		errRegexp string
	}{
		{mediaType: "application/vnd.kitops.modelkit.badbase.v1.tar", errRegexp: "invalid base type"},
		{mediaType: "application/vnd.kitops.modelkit.model.v1.tar+badCompression", errRegexp: "invalid compression"},
		{mediaType: "application/vnd.kitops.modelkit.model.v1.badFormat", errRegexp: "unrecognized media type"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Parsing %s", tt.mediaType), func(t *testing.T) {
			_, err := ParseMediaType(tt.mediaType)
			if !assert.Error(t, err) {
				return
			}
			assert.Regexp(t, tt.errRegexp, err.Error(), "Should return error for invalid media type")
		})
	}

}
