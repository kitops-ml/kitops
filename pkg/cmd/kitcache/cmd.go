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

package kitcache

import (
	"fmt"
	fscache "github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	"github.com/kitops-ml/kitops/pkg/output"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func CacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: `Manage temporary files cached by Kit`,
		Long: `Manage files stored in the temporary KitOps cache dir ($KITOPS_HOME/cache)

Normally, this directory is empty, but may contain leftover files from resumable
downloads or files that were not cleaned up due to the command being cancelled.

The $KITOPS_HOME location is system dependent:
	- Linux: $XDG_DATA_HOME/kitops with a fall back to $HOME/.local/share/kitops
	- MacOS: ~/Library/Caches/kitops
	- Windows: %LOCALAPPDATA%\kitops
`,
		Example: `# Get information about size of cached files
kit cache info

# Clear files in cache
kit cache clear
		`,
	}
	cmd.AddCommand(cacheInfoCommand())
	cmd.AddCommand(cacheClearCommand())

	return cmd
}

func cacheInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: `Get information about cache disk usage`,
		Long:  `Print the total size of temporary files in the cache directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			totalSize, stats, err := fscache.StatCache()
			if err != nil {
				return output.Fatalln(err)
			}
			if totalSize == 0 {
				output.Infof("Cache is currently empty")
				return nil
			}
			printCacheInfo(cmd.OutOrStdout(), totalSize, stats)
			return nil
		},
	}
	return cmd
}

func cacheClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: `Clear temporary cache storage`,
		Long:  `Clear temporary files from cache storage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := fscache.ClearCache(); err != nil {
				return output.Fatalln(err)
			}
			return nil
		},
	}
	return cmd
}

func printCacheInfo(w io.Writer, totalSize int64, subpaths map[string]int64) {
	var subpathNames []string
	for s := range subpaths {
		subpathNames = append(subpathNames, s)
	}
	sort.Strings(subpathNames)
	fmt.Fprintf(w, "Total size of cache directory: %s\n", output.FormatBytes(totalSize))
	fmt.Fprintln(w, "Cache contents:")
	tw := tabwriter.NewWriter(w, 0, 2, 4, ' ', 0)
	for _, subpath := range subpathNames {
		fmt.Fprintf(tw, "  ./%s\t%s\n", subpath, output.FormatBytes(subpaths[subpath]))
	}
	tw.Flush()
}
