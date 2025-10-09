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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/cmd/dev"
	"github.com/kitops-ml/kitops/pkg/cmd/diff"
	"github.com/kitops-ml/kitops/pkg/cmd/info"
	"github.com/kitops-ml/kitops/pkg/cmd/inspect"
	"github.com/kitops-ml/kitops/pkg/cmd/kitcache"
	"github.com/kitops-ml/kitops/pkg/cmd/kitimport"
	"github.com/kitops-ml/kitops/pkg/cmd/kitinit"
	"github.com/kitops-ml/kitops/pkg/cmd/list"
	"github.com/kitops-ml/kitops/pkg/cmd/login"
	"github.com/kitops-ml/kitops/pkg/cmd/logout"
	"github.com/kitops-ml/kitops/pkg/cmd/pack"
	"github.com/kitops-ml/kitops/pkg/cmd/pull"
	"github.com/kitops-ml/kitops/pkg/cmd/push"
	"github.com/kitops-ml/kitops/pkg/cmd/remove"
	"github.com/kitops-ml/kitops/pkg/cmd/tag"
	"github.com/kitops-ml/kitops/pkg/cmd/unpack"
	"github.com/kitops-ml/kitops/pkg/cmd/version"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/update"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

var (
	shortDesc = `Streamline the lifecycle of AI/ML models`
	longDesc  = `Kit is a tool for efficient AI/ML model lifecycle management.

Find more information at: http://kitops.ml`
)

type rootOptions struct {
	configHome   string
	verbosity    int
	loglevel     string
	progressBars string
}

func RunCommand() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   `kit`,
		Short: shortDesc,
		Long:  longDesc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			output.SetOut(cmd.OutOrStdout())
			output.SetErr(cmd.ErrOrStderr())
			if err := output.SetLogLevelFromString(opts.loglevel); err != nil {
				return output.Fatalln(err)
			}
			output.SetProgressBars(opts.progressBars)

			switch opts.verbosity {
			case 0:
				break
			case 1:
				output.Debugf("Setting verbosity to %s", output.LogLevelDebug)
				output.SetLogLevel(output.LogLevelDebug)
			case 2:
				output.Debugf("Setting verbosity to %s", output.LogLevelTrace)
				output.SetLogLevel(output.LogLevelTrace)
			default:
				output.Debugf("Setting verbosity to %s and disabling progress bars", output.LogLevelTrace)
				output.SetLogLevel(output.LogLevelTrace)
				output.SetProgressBars("none")
			}

			configHome, err := getConfigHome(opts)
			if err != nil {
				output.Errorf("Failed to read base config directory")
				output.Infof("Use the --config flag or set the $%s environment variable to provide a default", constants.KitopsHomeEnvVar)
				output.Debugf("Error: %s", err)
				return errors.New("exit")
			}
			ctx := context.WithValue(cmd.Context(), constants.ConfigKey{}, configHome)
			cache.SetCacheHome(constants.CachePath(configHome))
			cmd.SetContext(ctx)

			update.CheckForUpdate(configHome, cmd.CalledAs())

			// At this point, we've parsed the command tree and args; the CLI is being correctly
			// so we don't want to print usage. Each subcommand should print its error message before
			// returning
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			storagePath := constants.StoragePath(configHome)
			needsMigration, err := local.NeedsMigrate(storagePath)
			if err != nil {
				return output.Fatalf("Failed to determine if local modelkit needs to be migrated")
			} else if needsMigration {
				output.Infof("Migrating local storage to new format")
				if err := local.MigrateStorage(ctx, storagePath); err != nil {
					return output.Fatalf("Error migrating storage: %s", err)
				}
			}
			return nil
		},
	}
	addSubcommands(cmd)
	cmd.PersistentFlags().StringVar(&opts.loglevel, "log-level", "info", "Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info')")
	cmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"trace", "debug", "info", "warn", "error"}, cobra.ShellCompDirectiveDefault
	})
	cmd.PersistentFlags().StringVar(&opts.progressBars, "progress", "plain", "Configure progress bars for longer operations (options: none, plain, fancy)")
	cmd.RegisterFlagCompletionFunc("progress", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"none", "plain", "fancy"}, cobra.ShellCompDirectiveDefault
	})
	cmd.PersistentFlags().StringVar(&opts.configHome, "config", "", "Alternate path to root storage directory for CLI")
	cmd.PersistentFlags().CountVarP(&opts.verbosity, "verbose", "v", "Increase verbosity of output (use -vv for more)")
	cmd.PersistentFlags().SortFlags = false
	cmd.Flags().SortFlags = false

	cmd.SetHelpTemplate(helpTemplate)
	cmd.SetUsageTemplate(usageTemplate)
	cobra.AddTemplateFunc("indent", indentBlock)
	cobra.AddTemplateFunc("sectionHead", sectionHead)
	cobra.AddTemplateFunc("ensureTrailingNewline", ensureTrailingNewline)

	cmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveNoFileComp)
	return cmd
}

func addSubcommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(pack.PackCommand())
	rootCmd.AddCommand(unpack.UnpackCommand())
	rootCmd.AddCommand(push.PushCommand())
	rootCmd.AddCommand(pull.PullCommand())
	rootCmd.AddCommand(tag.TagCommand())
	rootCmd.AddCommand(list.ListCommand())
	rootCmd.AddCommand(inspect.InspectCommand())
	rootCmd.AddCommand(info.InfoCommand())
	rootCmd.AddCommand(remove.RemoveCommand())
	rootCmd.AddCommand(login.LoginCommand())
	rootCmd.AddCommand(logout.LogoutCommand())
	rootCmd.AddCommand(version.VersionCommand())
	rootCmd.AddCommand(dev.DevCommand())
	rootCmd.AddCommand(kitinit.InitCommand())
	rootCmd.AddCommand(diff.DiffCommand())
	rootCmd.AddCommand(kitimport.ImportCommand())
	rootCmd.AddCommand(kitcache.CacheCommand())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RunCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getConfigHome(opts *rootOptions) (string, error) {
	if opts.configHome != "" {
		output.Debugf("Using config directory from flag: %s", opts.configHome)
		absHome, err := filepath.Abs(opts.configHome)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", opts.configHome, err)
		}
		return absHome, nil
	}

	envHome := os.Getenv(constants.KitopsHomeEnvVar)
	if envHome != "" {
		output.Debugf("Using config directory from environment variable: %s", envHome)
		absHome, err := filepath.Abs(envHome)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", constants.KitopsHomeEnvVar, err)
		}
		return absHome, nil
	}

	defaultHome, err := constants.DefaultConfigPath()
	if err != nil {
		return "", err
	}
	output.Debugf("Using default config directory: %s", defaultHome)
	return defaultHome, nil
}
