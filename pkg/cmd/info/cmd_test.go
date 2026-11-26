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

package info

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/artifact"
)

func TestFilterKitfile(t *testing.T) {
	config := &artifact.KitFile{
		ManifestVersion: "1.0.0",
		Package:         artifact.Package{Name: "my-modelkit"},
		Code:            []artifact.Code{{Path: "src"}},
	}

	tests := []struct {
		filter  string
		want    string
		wantErr bool
	}{
		{filter: "package.name", want: "my-modelkit\n"},
		{filter: ".manifestVersion", want: "1.0.0\n"},
		{filter: "model", want: "null\n"},
		{filter: "package.unknown", wantErr: true},
		// Descending into a nil pointer, a list, or a scalar must return an error, not panic.
		{filter: "model.name", wantErr: true},
		{filter: "code.path", wantErr: true},
		{filter: "package.name.foo", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			got, err := filterKitfile(config, tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got output %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
