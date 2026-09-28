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

package options

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/spf13/cobra"
)

// NetworkOptions represent common networking-related flags that are used by multiple commands.
// The flags should be added to the command via AddNetworkFlags before running.
type NetworkOptions struct {
	PlainHTTP         bool
	TLSVerify         bool
	TLSTrustCertPaths []string
	CredentialsPath   string
	ClientCertPath    string
	ClientCertKeyPath string
	Concurrency       int
	Proxy             string
}

func (o *NetworkOptions) AddNetworkFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.PlainHTTP, "plain-http", false, "Use plain HTTP when connecting to remote registries")
	cmd.Flags().BoolVar(&o.TLSVerify, "tls-verify", true, "Require TLS and verify certificates when connecting to remote registries")
	cmd.Flags().StringSliceVar(&o.TLSTrustCertPaths, "tls-cert", nil, "Path to TLS cert to add to trust store (flag can be repeated)")
	cmd.Flags().StringVar(&o.ClientCertPath, "cert", "",
		fmt.Sprintf("Path to client certificate used for authentication (can also be set via environment variable %s)", constants.ClientCertEnvVar))
	cmd.Flags().StringVar(&o.ClientCertKeyPath, "key", "",
		fmt.Sprintf("Path to client certificate key used for authentication (can also be set via environment variable %s)", constants.ClientCertKeyEnvVar))
	cmd.Flags().IntVar(&o.Concurrency, "concurrency", 5, "Maximum number of simultaneous uploads/downloads")
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Proxy to use for connections (overrides proxy set by environment)")
}

var networkFlagNames = []string{"plain-http", "tls-verify", "tls-cert", "cert", "key", "concurrency", "proxy"}

// NetworkFlagsChanged reports whether any flag added by AddNetworkFlags was set
func NetworkFlagsChanged(cmd *cobra.Command) bool {
	for _, f := range networkFlagNames {
		if cmd.Flags().Changed(f) {
			return true
		}
	}
	return false
}

func (o *NetworkOptions) Complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	o.CredentialsPath = constants.CredentialsPath(configHome)

	if certPath := os.Getenv(constants.ClientCertEnvVar); certPath != "" {
		o.ClientCertPath = certPath
	}
	if certKeyPath := os.Getenv(constants.ClientCertKeyEnvVar); certKeyPath != "" {
		o.ClientCertKeyPath = certKeyPath
	}
	// Commands such as pack and unpack change the working directory, so relative paths must be resolved first
	certPaths := []*string{&o.ClientCertPath, &o.ClientCertKeyPath}
	for i := range o.TLSTrustCertPaths {
		certPaths = append(certPaths, &o.TLSTrustCertPaths[i])
	}
	for _, path := range certPaths {
		if *path == "" {
			continue
		}
		absPath, err := filepath.Abs(*path)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", *path, err)
		}
		*path = absPath
	}
	if o.Concurrency < 1 {
		return fmt.Errorf("invalid argument for concurrency (%d): must be at least 1", o.Concurrency)
	}

	return nil
}

func DefaultNetworkOptions(configHome string) *NetworkOptions {
	return &NetworkOptions{
		PlainHTTP:       false,
		TLSVerify:       true,
		CredentialsPath: constants.CredentialsPath(configHome),
	}
}
