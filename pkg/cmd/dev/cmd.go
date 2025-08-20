// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package dev

import (
	"net"
	"strconv"

	"github.com/kitops-ml/kitops/pkg/lib/harness"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
)

func DevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dev <directory|registry/repository[:tag|@digest]> [flags]",
		Short:   devShortDesc,
		Long:    devLongDesc,
		Example: devExample,
	}
	cmd.AddCommand(DevStartCommand())
	cmd.AddCommand(DevStopCommand())
	cmd.AddCommand(DevLogsCommand())
	return cmd
}

func DevStartCommand() *cobra.Command {
	opts := &DevStartOptions{}
	cmd := &cobra.Command{
		Use:     "start [directory|registry/repository[:tag|@digest]] [flags]",
		Short:   devStartShortDesc,
		Long:    devStartLongDesc,
		Example: devStartExample,
		RunE:    runStartCommand(opts),
	}
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Flags().StringVarP(&opts.modelFile, "file", "f", "", "Path to the kitfile")
	cmd.Flags().StringVar(&opts.host, "host", "127.0.0.1", "Host for the development server")
	cmd.Flags().IntVar(&opts.port, "port", 0, "Port for development server to listen on")
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func DevStopCommand() *cobra.Command {
	opts := &DevBaseOptions{}
	cmd := &cobra.Command{
		Use:   "stop",
		Short: devStopShortDesc,
		Long:  devStopLongDesc,
		RunE:  runStopCommand(opts),
	}
	return cmd
}

func DevLogsCommand() *cobra.Command {
	opts := &DevLogsOptions{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: devLogsShortDesc,
		Long:  devLogsLongDesc,
		RunE:  runLogsCommand(opts),
	}
	cmd.Flags().BoolVarP(&opts.follow, "follow", "f", false, "Stream the log file")
	return cmd
}

func runStartCommand(opts *DevStartOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("failed to complete options: %s", err)
		}

		err := runDev(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to start dev server: %s", err)
		}
		output.Infof("Development server started at http://%s:%d", opts.host, opts.port)
		output.Infof("Use \"kit dev stop\" to stop the development server")
		return nil
	}
}

func runStopCommand(opts *DevBaseOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("failed to complete options: %s", err)
		}

		output.Infoln("Stopping development server...")
		err := stopDev(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to stop dev server: %s", err)
		}
		output.Infoln("Development server stopped")
		return nil
	}
}

func runLogsCommand(opts *DevLogsOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("failed to complete options: %s", err)
		}

		err := harness.PrintLogs(opts.configHome, cmd.OutOrStdout(), opts.follow)
		if err != nil {
			return output.Fatalln(err)
		}
		return nil
	}
}

func findAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}
