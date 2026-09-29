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

package push

import (
	"testing"
)

func TestPushCommandUploadChunkSizeDefault(t *testing.T) {
	t.Parallel()

	got, err := PushCommand().Flags().GetInt64("chunk-size")
	if err != nil {
		t.Fatalf("get chunk-size flag: %v", err)
	}
	if got != 32 {
		t.Fatalf("chunk-size default = %d MiB, want 32 MiB", got)
	}
}

func TestUploadChunkSizeBytes(t *testing.T) {
	t.Parallel()

	for _, sizeMiB := range []int64{16, 32} {
		got, err := uploadChunkSizeBytes(sizeMiB)
		if err != nil {
			t.Fatalf("uploadChunkSizeBytes(%d): %v", sizeMiB, err)
		}
		if want := sizeMiB * bytesPerMiB; got != want {
			t.Fatalf("uploadChunkSizeBytes(%d) = %d, want %d", sizeMiB, got, want)
		}
	}

	for _, sizeMiB := range []int64{-1, 0, 15, 33} {
		if _, err := uploadChunkSizeBytes(sizeMiB); err == nil {
			t.Fatalf("uploadChunkSizeBytes(%d) should fail", sizeMiB)
		}
	}
}

// AGENT_MODIFIED: Human review required before merge
