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

package login

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/network"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	"oras.land/oras-go/v2/registry/remote/credentials"
)

func login(ctx context.Context, opts *loginOptions) error {
	credentialsStorePath := constants.CredentialsPath(opts.configHome)
	store, err := network.NewCredentialStore(credentialsStorePath)
	if err != nil {
		return err
	}
	registry, err := remote.NewRegistry(opts.registry, &opts.NetworkOptions)
	if err != nil {
		return fmt.Errorf("could not resolve registry %s: %w", opts.registry, err)
	}
	if err := credentials.Login(ctx, store, registry, opts.credential); err != nil {
		return err
	}
	output.Infoln("Log in successful")
	return nil
}
