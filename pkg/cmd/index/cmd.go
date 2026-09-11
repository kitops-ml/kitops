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
	"github.com/spf13/cobra"
)

const (
	shortDesc = `Manage ModelKit indexes`
	longDesc  = `Create and manage ModelKit indexes.

A ModelKit index groups variants of a model under a single reference. Use
'kit pull -l' or 'kit unpack -l' to select a ModelKit from an index by label.`

	example = `# Create an index in local storage
kit index create my-org/my-model:all

# Add ModelKits to the index
kit index add my-org/my-model:all my-org/my-model:q4_0
kit index add my-org/my-model:all my-org/my-model:q8_0

# Remove a ModelKit from the index
kit index remove my-org/my-model:all my-org/my-model:q4_0

# List indexes in local storage
kit index list

# List the ModelKits in the index
kit index info my-org/my-model:all

# Print the index manifest
kit index inspect my-org/my-model:all

# Push the index to a remote registry
kit index push my-org/my-model:all registry.example.com/my-org/my-model:all

# Pull an index from a remote registry
kit index pull registry.example.com/my-org/my-model:all

# Delete the index
kit index delete my-org/my-model:all`
)

func IndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "index",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
	}

	cmd.AddCommand(indexCreateCommand())
	cmd.AddCommand(indexDeleteCommand())
	cmd.AddCommand(indexAddCommand())
	cmd.AddCommand(indexRemoveCommand())
	cmd.AddCommand(indexListCommand())
	cmd.AddCommand(indexInfoCommand())
	cmd.AddCommand(indexInspectCommand())
	cmd.AddCommand(indexPullCommand())
	cmd.AddCommand(indexPushCommand())

	return cmd
}
