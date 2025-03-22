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

package version

import (
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/update"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

const (
	shortDesc = `Display the version information for the CLI`
	longDesc  = `The version command prints detailed version information.

This information includes the current version of the tool, the Git commit that
the version was built from, the build time, and the version of Go it was
compiled with.`
)

const versionNotifFlag = "show-update-notifications"

type versionOpts struct {
	shouldShowNotifications bool
}

func VersionCommand() *cobra.Command {
	opts := &versionOpts{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: shortDesc,
		Long:  longDesc,
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Changed(versionNotifFlag) {
				configHome, ok := cmd.Context().Value(constants.ConfigKey{}).(string)
				if !ok {
					output.Fatalln("default config path not set on command context")
				}
				if err := update.SetShowNotifications(configHome, opts.shouldShowNotifications); err != nil {
					output.Fatalln(err)
				}
			} else {
				output.Infof("Version: %s\nCommit: %s\nBuilt: %s\nGo version: %s\n", constants.Version, constants.GitCommit, constants.BuildTime, constants.GoVersion)
			}
		},
	}
	cmd.Flags().BoolVar(&opts.shouldShowNotifications, versionNotifFlag, false, "Enable or disable update notifications for the Kit CLI")

	return cmd
}

func init() {}
