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

import "testing"

func TestWithUploadChunkSize(t *testing.T) {
	t.Parallel()

	repo := &Repository{uploadChunkSize: DefaultUploadChunkSize}
	if err := WithUploadChunkSize(16 << 20)(repo); err != nil {
		t.Fatalf("WithUploadChunkSize(): %v", err)
	}
	if repo.uploadChunkSize != 16<<20 {
		t.Fatalf("uploadChunkSize = %d, want %d", repo.uploadChunkSize, 16<<20)
	}

	if err := WithUploadChunkSize(0)(repo); err == nil {
		t.Fatal("WithUploadChunkSize(0) should fail")
	}
}

// AGENT_MODIFIED: Human review required before merge
