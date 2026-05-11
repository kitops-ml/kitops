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

package kit

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/network"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

// LoginOptions contains configuration for logging in to a registry.
type LoginOptions struct {
	options.NetworkOptions
	ConfigHome string
	Registry   string
	Credential auth.Credential
}

// Login logs in to a registry and saves credentials.
func Login(ctx context.Context, opts *LoginOptions) error {
	applyNetworkDefaults(&opts.NetworkOptions, opts.ConfigHome)
	credentialsStorePath := constants.CredentialsPath(opts.ConfigHome)
	store, err := network.NewCredentialStore(credentialsStorePath)
	if err != nil {
		return err
	}
	reg, err := remote.NewRegistry(opts.Registry, &opts.NetworkOptions)
	if err != nil {
		return fmt.Errorf("could not resolve registry %s: %w", opts.Registry, err)
	}
	if err := credentials.Login(ctx, store, reg, opts.Credential); err != nil {
		return err
	}
	output.Infoln("Log in successful")
	return nil
}

// LogoutOptions contains configuration for logging out from a registry.
type LogoutOptions struct {
	ConfigHome string
	Registry   string
}

// Logout logs out from a registry and removes saved credentials.
func Logout(ctx context.Context, opts *LogoutOptions) error {
	credentialsPath := constants.CredentialsPath(opts.ConfigHome)
	store, err := network.NewCredentialStore(credentialsPath)
	if err != nil {
		return err
	}
	if err := credentials.Logout(ctx, store, opts.Registry); err != nil {
		return err
	}
	output.Infof("Successfully logged out from %s", opts.Registry)
	return nil
}
