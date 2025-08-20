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

package unpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		filter          string
		expectError     bool
		errorContains   string
		expectedTypes   []string
		expectedFilters []string
	}{
		{
			name:          "simple model filter",
			filter:        "model",
			expectedTypes: []string{"model"},
		},
		{
			name:            "model with specific filters",
			filter:          "model:llama-7b,gpt-3",
			expectedTypes:   []string{"model"},
			expectedFilters: []string{"llama-7b", "gpt-3"},
		},
		{
			name:          "multiple types",
			filter:        "model,datasets,code",
			expectedTypes: []string{"model", "dataset", "code"},
		},
		{
			name:            "multiple types with filters",
			filter:          "datasets:training-data,validation",
			expectedTypes:   []string{"dataset"},
			expectedFilters: []string{"training-data", "validation"},
		},
		{
			name:          "too many colons",
			filter:        "model:filter1:filter2:extra",
			expectError:   true,
			errorContains: "invalid filter: should be in format",
		},
		{
			name:          "invalid type",
			filter:        "invalidtype",
			expectError:   true,
			errorContains: "invalid filter type",
		},
		{
			name:          "empty filter",
			filter:        "",
			expectError:   true,
			errorContains: "invalid filter",
		},
		{
			name:          "filter with empty type",
			filter:        ":filter1,filter2",
			expectError:   true,
			errorContains: "invalid filter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFilter(tt.filter)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify base types
			if tt.expectedTypes != nil {
				assert.Equal(t, tt.expectedTypes, result.BaseTypes)
			}

			// Verify filters
			if tt.expectedFilters != nil {
				assert.Equal(t, tt.expectedFilters, result.Filters)
			} else {
				// If no specific filters expected, should be empty
				assert.Empty(t, result.Filters)
			}
		})
	}
}

func TestFiltersFromUnpackConf(t *testing.T) {
	tests := []struct {
		name              string
		unpackKitfile     bool
		unpackModels      bool
		unpackCode        bool
		unpackDatasets    bool
		unpackDocs        bool
		expectedTypeCount int
		expectedTypes     []string
	}{
		{
			name:              "only models",
			unpackModels:      true,
			expectedTypeCount: 1,
			expectedTypes:     []string{"model"},
		},
		{
			name:              "models and datasets",
			unpackModels:      true,
			unpackDatasets:    true,
			expectedTypeCount: 2,
			expectedTypes:     []string{"model", "dataset"},
		},
		{
			name:              "all types",
			unpackKitfile:     true,
			unpackModels:      true,
			unpackCode:        true,
			unpackDatasets:    true,
			unpackDocs:        true,
			expectedTypeCount: 5,
			expectedTypes:     []string{"config", "model", "docs", "dataset", "code"},
		},
		{
			name:              "none selected",
			expectedTypeCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FiltersFromUnpackConf(
				tt.unpackKitfile,
				tt.unpackModels,
				tt.unpackCode,
				tt.unpackDatasets,
				tt.unpackDocs,
			)

			require.Len(t, result, 1, "Should return exactly one filter config")

			filterConf := result[0]
			assert.Len(t, filterConf.BaseTypes, tt.expectedTypeCount)

			if tt.expectedTypes != nil {
				// Check that all expected types are present (order may vary)
				for _, expectedType := range tt.expectedTypes {
					assert.Contains(t, filterConf.BaseTypes, expectedType)
				}
			}

			// Filters should be empty for deprecated config conversion
			assert.Empty(t, filterConf.Filters)
		})
	}
}
