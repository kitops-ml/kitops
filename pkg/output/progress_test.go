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

package output

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vbauerster/mpb/v8"
)

func TestDownloadProgressDoneWithZeroByteFile(t *testing.T) {
	pb := &DownloadProgressBar{progress: mpb.New(mpb.WithOutput(io.Discard))}
	rc := pb.TrackDownload(io.NopCloser(strings.NewReader("")), "empty", 0)
	_, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.NoError(t, rc.Close())

	done := make(chan struct{})
	go func() {
		pb.Done()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Done() did not return for a zero-byte download")
	}
}
