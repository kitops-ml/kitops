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
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldUseColor(t *testing.T) {
	assert.False(t, shouldUseColor(&bytes.Buffer{}), "a non-file writer is never a terminal")

	devNull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer devNull.Close()
	assert.False(t, shouldUseColor(devNull), "a file that is not a terminal must not be colored")
}

func TestGetPrefixOmitsColorForNonTerminalOutput(t *testing.T) {
	origOut, origErr, origLevel := stdout, stderr, logLevel
	t.Cleanup(func() {
		stdout, stderr, logLevel = origOut, origErr, origLevel
	})

	SetOut(&bytes.Buffer{})
	SetErr(&bytes.Buffer{})
	// Any level other than LogLevelInfo, so that the [INFO ] prefix is printed.
	SetLogLevel(LogLevelTrace)

	levels := []LogLevel{LogLevelTrace, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelSystem}
	for _, level := range levels {
		assert.NotContains(t, level.getPrefix(), "\033", "level %d must not emit color codes to a non-terminal writer", level)
	}
}
