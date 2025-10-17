// Copyright 2025 The KitOps Authors.
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

package mediatype

import (
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	ArtifactTypeKitManifest   = "application/vnd.kitops.modelkit.manifest.v1+json"
	ArtifactTypeModelManifest = "application/vnd.cncf.model.manifest.v1+json"
)

func ModelFormatForManifest(manifest *ocispec.Manifest) (ModelFormat, error) {
	if manifest.ArtifactType == ArtifactTypeKitManifest || manifest.Config.MediaType == KitConfigMediaType.String() {
		return KitFormat, nil
	}
	if manifest.ArtifactType == ArtifactTypeModelManifest || manifest.Config.MediaType == ModelPackConfigMediaType.String() {
		return ModelPackFormat, nil
	}
	return UnknownModelFormat, fmt.Errorf("manifest is not a Model manifest: artifactType is %s, config mediaType is %s", manifest.ArtifactType, manifest.Config.MediaType)
}
