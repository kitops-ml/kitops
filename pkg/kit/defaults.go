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
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
)

// applyNetworkDefaults sets safe defaults for NetworkOptions when callers leave
// them at zero values. This ensures library consumers get TLS verification
// enabled and credentials resolved without explicit configuration.
func applyNetworkDefaults(opts *options.NetworkOptions, configHome string) {
	if opts.Concurrency < 1 {
		opts.Concurrency = 5
	}
	if opts.CredentialsPath == "" {
		opts.CredentialsPath = constants.CredentialsPath(configHome)
		if !opts.PlainHTTP && !opts.TLSVerify {
			opts.TLSVerify = true
		}
	}
}
