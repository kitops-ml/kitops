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

package index

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

const (
	listShortDesc = `List ModelKit indexes in local storage`
	listLongDesc  = `List ModelKit indexes in local storage, with the number of ModelKits each one
references.`

	listExample = `# List local indexes
kit index list`
)

type listOptions struct {
	configHome string
}

func indexListCommand() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   listShortDesc,
		Long:    listLongDesc,
		Example: listExample,
		RunE:    runListCommand(opts),
		Args:    cobra.NoArgs,
	}
	cmd.Flags().SortFlags = false

	return cmd
}

func runListCommand(opts *listOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context()); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		return output.Fatalf("Not implemented: kit index list")
	}
}

func (opts *listOptions) complete(ctx context.Context) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	return nil
}
