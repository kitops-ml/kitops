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

package inspect

import (
	"context"
	"errors"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

// Utility struct for formatting output of inspect
type inspectInfo struct {
	Digest     digest.Digest     `json:"digest,omitempty" yaml:"digest,omitempty"`
	CLIVersion string            `json:"cliVersion,omitempty" yaml:"cliVersion,omitempty"`
	Kitfile    *artifact.KitFile `json:"kitfile,omitempty" yaml:"kitfile,omitempty"`
	Manifest   *ocispec.Manifest `json:"manifest,omitempty" yaml:"manifest,omitempty"`
}

func inspectReference(ctx context.Context, opts *inspectOptions) (*inspectInfo, error) {
	if opts.checkRemote {
		return getRemoteInspect(ctx, opts)
	} else {
		return getLocalInspect(ctx, opts)
	}
}

func getLocalInspect(ctx context.Context, opts *inspectOptions) (*inspectInfo, error) {
	storageRoot := constants.StoragePath(opts.configHome)
	localRepo, err := local.NewLocalRepo(storageRoot, opts.modelRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}
	return getInspectInfo(ctx, localRepo, opts.modelRef.Reference)
}

func getRemoteInspect(ctx context.Context, opts *inspectOptions) (*inspectInfo, error) {
	repository, err := remote.NewRepository(ctx, opts.modelRef.Registry, opts.modelRef.Repository, &opts.NetworkOptions)
	if err != nil {
		return nil, err
	}
	return getInspectInfo(ctx, repository, opts.modelRef.Reference)
}

func getInspectInfo(ctx context.Context, repository oras.Target, ref string) (*inspectInfo, error) {
	desc, manifest, kitfile, err := util.ResolveManifestAndConfig(ctx, repository, ref)
	if err != nil && !errors.Is(err, util.ErrNoKitfile) {
		return nil, err
	}
	version := "unknown"
	if manifest.Annotations != nil && manifest.Annotations[constants.CliVersionAnnotation] != "" {
		version = manifest.Annotations[constants.CliVersionAnnotation]
	}
	return &inspectInfo{
		Digest:     desc.Digest,
		CLIVersion: version,
		Kitfile:    kitfile,
		Manifest:   manifest,
	}, nil
}
