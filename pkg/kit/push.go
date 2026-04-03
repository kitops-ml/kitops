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
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

type PushOptions struct {
	options.NetworkOptions
	ConfigHome   string
	SrcModelRef  *registry.Reference
	DestModelRef *registry.Reference
}

type PushResult struct {
	Descriptor ocispec.Descriptor
}

func Push(ctx context.Context, opts *PushOptions) (*PushResult, error) {
	if opts.NetworkOptions.Concurrency < 1 {
		opts.NetworkOptions.Concurrency = 5
	}
	if opts.NetworkOptions.CredentialsPath == "" {
		opts.NetworkOptions.CredentialsPath = constants.CredentialsPath(opts.ConfigHome)
		if !opts.NetworkOptions.PlainHTTP && !opts.NetworkOptions.TLSVerify {
			opts.NetworkOptions.TLSVerify = true
		}
	}

	remoteRepo, err := remote.NewRepository(
		ctx,
		opts.DestModelRef.Registry,
		opts.DestModelRef.Repository,
		&opts.NetworkOptions,
	)
	if err != nil {
		return nil, err
	}

	localRepo, err := local.NewLocalRepo(constants.StoragePath(opts.ConfigHome), opts.SrcModelRef)
	if err != nil {
		return nil, err
	}

	desc, err := pushModel(ctx, localRepo, remoteRepo, opts)
	if err != nil {
		return nil, err
	}
	return &PushResult{Descriptor: desc}, nil
}

func pushModel(ctx context.Context, localRepo local.LocalRepo, repo registry.Repository, opts *PushOptions) (ocispec.Descriptor, error) {
	trackedRepo, logger := output.WrapTarget(repo)
	srcTag := opts.SrcModelRef.Reference
	destTag := opts.DestModelRef.Reference
	copyOpts := oras.CopyOptions{}
	copyOpts.Concurrency = opts.Concurrency
	desc, err := oras.Copy(ctx, localRepo, srcTag, trackedRepo, destTag, copyOpts)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to copy to remote: %w", err)
	}
	logger.Wait()

	return desc, err
}
