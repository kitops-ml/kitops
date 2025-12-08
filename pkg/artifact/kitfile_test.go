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

package artifact

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v3"
)

type parameterTestCase struct {
	Name        string
	Description string `yaml:"description"`
	KitfileYaml string `yaml:"kitfileYaml"`
	KitfileJson string `yaml:"kitfileJson"`
}

func (tc parameterTestCase) withName(name string) parameterTestCase {
	tc.Name = name
	return tc
}

func TestParameterMarshalUnmarshal(t *testing.T) {
	tests := loadAllTestCasesOrPanic[parameterTestCase](t, filepath.Join("testdata", "parameters"))
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s)", tt.Name, tt.Description), func(t *testing.T) {
			kf := &KitFile{}
			rc := io.NopCloser(strings.NewReader(tt.KitfileYaml))
			err := kf.LoadModel(rc)
			if !assert.NoError(t, err) {
				return
			}

			unmarshalledYaml, err := kf.MarshalToYAML()
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.KitfileYaml, string(unmarshalledYaml))

			unmarshalledJson, err := kf.MarshalToJSON()
			if !assert.NoError(t, err) {
				return
			}
			if tt.KitfileJson != "" {
				assert.Equal(t, tt.KitfileJson, string(unmarshalledJson))
			}
		})
	}
}

func loadAllTestCasesOrPanic[T interface{ withName(string) T }](t *testing.T, testsPath string) []T {
	files, err := os.ReadDir(testsPath)
	if err != nil {
		t.Fatal(err)
	}
	var tests []T
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join(testsPath, file.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var testcase T
		if err := yaml.Unmarshal(bytes, &testcase); err != nil {
			t.Fatal(err)
		}
		testcase = testcase.withName(file.Name())
		tests = append(tests, testcase)
	}
	return tests
}

func TestCollectLicenses(t *testing.T) {
	testcases := []struct {
		desc     string
		kitfile  string
		licenses []string
	}{
		{
			desc: "Empty kitfile",
			kitfile: `
manifestVersion: 1.0.0
`,
			licenses: nil,
		},
		{
			desc: "Kitfile with no licenses",
			kitfile: `
manifestVersion: 1.0.0
package:
  name: "test-package"
model:
  path: model-files
`,
			licenses: nil,
		},
		{
			desc: "Kitfile with package license",
			kitfile: `
manifestVersion: 1.0.0
package:
  name: "test-package"
  license: "Apache-2.0"
model:
  path: model-files
`,
			licenses: []string{"Apache-2.0"},
		},
		{
			desc: "Kitfile with multiple licenses, to be sorted",
			kitfile: `
manifestVersion: 1.0.0
package:
  name: "test-package"
  license: "license-g"
model:
  path: model-files
  license: "license-h"
  parts:
  - path: part-files
    license: "license-f"
  - path: part-files
    license: "license-e"
datasets:
- path: dataset
  license: "license-c"
- path: dataset-extra
  license: "license-d"
code:
- path: code
  license: "license-b"
- path: code-extra
  license: "license-a"
`,
			licenses: []string{"license-a", "license-b", "license-c", "license-d", "license-e", "license-f", "license-g", "license-h"},
		},
		{
			desc: "Kitfile with multiple licenses, to be deduplicated",
			kitfile: `
manifestVersion: 1.0.0
package:
  name: "test-package"
  license: Apache-2.0
model:
  path: model-files
  license: MIT
  parts:
  - path: part-files
    license: Apache-2.0
  - path: part-files
    license: MIT
datasets:
- path: dataset
  license: Apache-2.0
- path: dataset-extra
  license: MIT
code:
- path: code
  license: MIT
- path: code-extra
  license: MIT
`,
			licenses: []string{"Apache-2.0", "MIT"},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.desc, func(t *testing.T) {
			kf := &KitFile{}
			err := kf.LoadModel(io.NopCloser(strings.NewReader(tt.kitfile)))
			if !assert.NoError(t, err, "Unexpected error loading testcase Kitfile") {
				return
			}
			actualLicenses := kf.collectLicenses()
			assert.Equal(t, actualLicenses, tt.licenses, "Licenses should match, including order")
		})
	}
}
