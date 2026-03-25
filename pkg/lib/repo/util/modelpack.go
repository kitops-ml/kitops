// Copyright 2026 The KitOps Authors.
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

package util

import (
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"
	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// GenerateKitfileForModelPack generates a minimal Kitfile for a ModelPack manifest.
// This fallback is suitable for unpack and inspect workflows when no Kitfile is
// available in annotations.
func GenerateKitfileForModelPack(manifest *ocispec.Manifest) (*artifact.KitFile, error) {
	if format, err := mediatype.ModelFormatForManifest(manifest); err != nil || format != mediatype.ModelPackFormat {
		return nil, fmt.Errorf("not a modelpack artifact")
	}
	kf := &artifact.KitFile{
		Model: &artifact.Model{},
	}
	for _, desc := range manifest.Layers {
		if desc.Annotations == nil || desc.Annotations[modelspecv1.AnnotationFilepath] == "" {
			return nil, fmt.Errorf("unknown file path for layer: no %s annotation", modelspecv1.AnnotationFilepath)
		}
		filepath := desc.Annotations[modelspecv1.AnnotationFilepath]
		mt, err := mediatype.ParseMediaType(desc.MediaType)
		if err != nil {
			return nil, err
		}
		switch mt.Base() {
		case mediatype.ModelBaseType:
			kf.Model.Path = filepath
		case mediatype.ModelPartBaseType:
			kf.Model.Parts = append(kf.Model.Parts, artifact.ModelPart{Path: filepath})
		case mediatype.CodeBaseType:
			kf.Code = append(kf.Code, artifact.Code{Path: filepath})
		case mediatype.DatasetBaseType:
			kf.DataSets = append(kf.DataSets, artifact.DataSet{Path: filepath})
		case mediatype.DocsBaseType:
			kf.Docs = append(kf.Docs, artifact.Docs{Path: filepath})
		default:
			return nil, fmt.Errorf("unknown layer type: %s", mt)
		}
	}
	return kf, nil
}
