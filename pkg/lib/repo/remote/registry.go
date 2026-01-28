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

package remote

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/network"
	"github.com/kitops-ml/kitops/pkg/output"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// NewRegistry returns a new *remote.Registry for hostname, with credentials and TLS
// configured.
func NewRegistry(hostname string, opts *options.NetworkOptions) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(hostname)
	if err != nil {
		return nil, err
	}

	reg.PlainHTTP = opts.PlainHTTP
	credentialStore, err := network.NewCredentialStore(opts.CredentialsPath)
	if err != nil {
		return nil, err
	}
	authClient, err := network.ClientWithAuth(credentialStore, opts)
	if err != nil {
		return nil, err
	}
	authClient.Client.Transport = output.WrapHTTPTransport(authClient.Client.Transport)
	reg.Client = authClient

	return reg, nil
}

func NewRepository(ctx context.Context, hostname, repository string, opts *options.NetworkOptions) (registry.Repository, error) {
	reg, err := NewRegistry(hostname, opts)
	if err != nil {
		return nil, fmt.Errorf("could not resolve registry: %w", err)
	}
	repo, err := reg.Repository(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	ref := registry.Reference{
		Registry:   hostname,
		Repository: repository,
	}

	return &Repository{
		Repository: repo,
		Reference:  ref,
		PlainHttp:  opts.PlainHTTP,
		Client:     reg.Client,
	}, nil
}
