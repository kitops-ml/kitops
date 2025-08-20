// Copyright 2025 The KitOps Authors.
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

package completions

import (
	"strings"

	"github.com/spf13/cobra"
)

// WithStaticArgCompletions adds static argument completions to a command
func WithStaticArgCompletions(cmd *cobra.Command, lazyLoadCompletions func() ([]string, error), maxCompletionsCount int) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > (maxCompletionsCount - 1) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var suggestions []string
		completions, err := lazyLoadCompletions()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		for _, c := range completions {
			if strings.HasPrefix(c, toComplete) {
				suggestions = append(suggestions, c)
			}
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
}
