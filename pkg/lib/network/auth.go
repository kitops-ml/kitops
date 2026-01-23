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

package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// NewCredentialStore returns a credential store from @storePath and falls back to Docker's native store for reads only.
func NewCredentialStore(storePath string) (credentials.Store, error) {
	existingCredStore, err := credentials.NewStore(storePath, credentials.StoreOptions{
		DetectDefaultNativeStore: true,
		AllowPlaintextPut:        true,
	})
	if err != nil {
		return nil, err
	}

	storeOpts := credentials.StoreOptions{}
	dockerCredStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return nil, err
	}

	return credentials.NewStoreWithFallbacks(existingCredStore, dockerCredStore), nil
}

// ClientWithAuth returns a default *auth.Client using the provided credentials
// store
func ClientWithAuth(store credentials.Store, opts *options.NetworkOptions) (*auth.Client, error) {
	client, err := DefaultClient(opts)
	if err != nil {
		return nil, err
	}
	client.Credential = credentials.Credential(store)

	return client, nil
}

// DefaultClient returns an *auth.Client with a default User-Agent header and TLS
// configured from opts (optionally disabling TLS verification)
func DefaultClient(opts *options.NetworkOptions) (*auth.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = !opts.TLSVerify

	if len(opts.TLSTrustCertPaths) > 0 {
		if opts.PlainHTTP {
			output.Logf(output.LogLevelWarn, "Flag '--tls-cert' has no effect when --plain-http is specified")
		}
		if !opts.TLSVerify {
			output.Logf(output.LogLevelWarn, "Flag '--tls-cert' has no effect when --tls-verify=false is specified")
		}
		certPool, err := getCertsTrust(opts)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig.RootCAs = certPool
	}

	if opts.Proxy != "" {
		proxyURL, err := url.Parse(opts.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	if opts.ClientCertKeyPath != "" && opts.ClientCertPath != "" {
		cert, err := tls.LoadX509KeyPair(opts.ClientCertPath, opts.ClientCertKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate: %w", err)
		}
		transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
	}

	client := &auth.Client{
		Client: &http.Client{
			Transport: retry.NewTransport(transport),
		},
		Cache: auth.NewCache(),
		Header: http.Header{
			"User-Agent": {"kitops-cli/" + constants.Version},
		},
	}

	return client, nil
}

func getCertsTrust(opts *options.NetworkOptions) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		output.Logf(output.LogLevelWarn, "Error reading system certificates: %s", err)
		rootCAs = x509.NewCertPool()
	}

	for _, path := range opts.TLSTrustCertPaths {
		certsPEM, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate at path %s: %w", path, err)
		}
		if !rootCAs.AppendCertsFromPEM(certsPEM) {
			output.Logf(output.LogLevelWarn, "Failed to add certificate at path %s to pool", path)
		}
	}
	return rootCAs, nil
}
