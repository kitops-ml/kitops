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

package constants

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPathUsesKitopsHome(t *testing.T) {
	t.Setenv(KitopsHomeEnvVar, "custom-kitops-home")

	configPath, err := ConfigPath()
	require.NoError(t, err)

	expected, err := filepath.Abs("custom-kitops-home")
	require.NoError(t, err)
	assert.Equal(t, expected, configPath)
}

func TestConfigPathFallsBackToDefaultConfigPath(t *testing.T) {
	t.Setenv(KitopsHomeEnvVar, "")

	configPath, err := ConfigPath()
	require.NoError(t, err)

	defaultPath, err := DefaultConfigPath()
	require.NoError(t, err)
	assert.Equal(t, defaultPath, configPath)
}
