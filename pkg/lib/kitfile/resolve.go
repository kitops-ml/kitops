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

package kitfile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/local"
	"github.com/kitops-ml/kitops/pkg/lib/repo/remote"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"

	"oras.land/oras-go/v2/registry"
)

func GetKitfileForRefString(ctx context.Context, configHome string, ref string) (*artifact.KitFile, error) {
	modelRef, _, err := util.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	return GetKitfileForRef(ctx, configHome, modelRef)
}

func GetKitfileForRef(ctx context.Context, configHome string, ref *registry.Reference) (*artifact.KitFile, error) {
	storageRoot := constants.StoragePath(configHome)
	localRepo, err := local.NewLocalRepo(storageRoot, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to read local storage: %w", err)
	}
	_, _, localKitfile, err := util.ResolveManifestAndConfig(ctx, localRepo, ref.Reference)
	if err == nil {
		return localKitfile, nil
	} else if errors.Is(err, util.ErrNoKitfile) || errors.Is(err, util.ErrNotAModelKit) {
		// We found an artifact but it does not have a Kitfile
		return nil, err
	}

	repository, err := remote.NewRepository(ctx, ref.Registry, ref.Repository, options.DefaultNetworkOptions(configHome))
	if err != nil {
		return nil, err
	}
	_, _, remoteKitfile, err := util.ResolveManifestAndConfig(ctx, repository, ref.Reference)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Kitfile for %s: %w", ref, err)
	}
	return remoteKitfile, nil
}

// ResolveKitfile returns the Kitfile for a reference. Any references to other modelkits
// are fetched and included in the resolved Kitfile, giving the equivalent Kitfile including
// the model, datasets, and code from those referenced modelkits.
func ResolveKitfile(ctx context.Context, configHome, kitfileRef, baseRef string) (*artifact.KitFile, error) {
	resolved := &artifact.KitFile{}
	refChain := []string{baseRef, kitfileRef}
	for i := 0; i < constants.MaxModelRefChain; i++ {
		kitfile, err := GetKitfileForRefString(ctx, configHome, kitfileRef)
		if err != nil {
			return nil, err
		}
		resolved = mergeKitfiles(resolved, kitfile)
		if resolved.Model == nil || !util.IsModelKitReference(resolved.Model.Path) {
			if err := ValidateKitfile(resolved); err != nil {
				return nil, err
			}
			return resolved, nil
		}
		if idx := getIndex(refChain, resolved.Model.Path); idx != -1 {
			cycleStr := fmt.Sprintf("[%s=>%s]", strings.Join(refChain[idx:], "=>"), resolved.Model.Path)
			return nil, fmt.Errorf("Found cycle in modelkit references: %s", cycleStr)
		}
		refChain = append(refChain, resolved.Model.Path)
		kitfileRef = resolved.Model.Path
	}
	return nil, fmt.Errorf("reached maximum number of model references: [%s]", strings.Join(refChain, "=>"))
}

func mergeKitfiles(into, from *artifact.KitFile) *artifact.KitFile {
	firstNonEmpty := func(strs ...string) string {
		for _, s := range strs {
			if s != "" {
				return s
			}
		}
		return ""
	}

	result := &artifact.KitFile{}
	result.ManifestVersion = firstNonEmpty(into.ManifestVersion, from.ManifestVersion)
	result.Package.Name = firstNonEmpty(into.Package.Name, from.Package.Name)
	result.Package.Description = firstNonEmpty(into.Package.Description, from.Package.Description)
	result.Package.License = firstNonEmpty(into.Package.License, from.Package.License)
	result.Package.Version = firstNonEmpty(into.Package.Version, from.Package.Version)
	result.Package.Authors = append(into.Package.Authors, from.Package.Authors...)

	if into.Model != nil || from.Model != nil {
		result.Model = &artifact.Model{}
		intoModel := into.Model
		fromModel := from.Model
		if intoModel == nil {
			intoModel = &artifact.Model{}
		}
		if fromModel == nil {
			fromModel = &artifact.Model{}
		}
		result.Model.Path = fromModel.Path
		result.Model.Name = firstNonEmpty(intoModel.Name, fromModel.Name)
		result.Model.Description = firstNonEmpty(intoModel.Description, fromModel.Description)
		result.Model.Framework = firstNonEmpty(intoModel.Framework, fromModel.Framework)
		result.Model.Version = firstNonEmpty(intoModel.Version, fromModel.Version)
		result.Model.Parts = append(intoModel.Parts, fromModel.Parts...)
	}

	result.Code = into.Code
	result.Code = append(result.Code, from.Code...)
	result.DataSets = into.DataSets
	result.DataSets = append(result.DataSets, from.DataSets...)
	result.Docs = into.Docs
	result.Docs = append(result.Docs, from.Docs...)
	return result
}

func getIndex(list []string, s string) int {
	for idx, item := range list {
		if s == item {
			return idx
		}
	}
	return -1
}
