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

package list

import (
	"fmt"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	listTableHeader = "REPOSITORY\tTAG\tMAINTAINER\tNAME\tSIZE\tDIGEST"
	listTableFmt    = "%s\t%s\t%s\t%s\t%s\t%s"
)

type modelInfo struct {
	Repo      string   `json:"repo"`
	Digest    string   `json:"digest"`
	Tags      []string `json:"tags"`
	ModelName string   `json:"modelName"`
	Size      string   `json:"size"`
	Author    string   `json:"author"`
}

func (m *modelInfo) format() []string {
	if len(m.Tags) == 0 {
		line := fmt.Sprintf(listTableFmt, m.Repo, "<none>", m.Author, m.ModelName, m.Size, m.Digest)
		return []string{line}
	}
	var lines []string
	for _, tag := range m.Tags {
		line := fmt.Sprintf(listTableFmt, m.Repo, tag, m.Author, m.ModelName, m.Size, m.Digest)
		lines = append(lines, line)
	}
	return lines
}

// fill adds information pulled from a manifest and kitfile into the modelInfo. Handles
// cases where the Kitfile is nil
func (m *modelInfo) fill(manifest *ocispec.Manifest, kitfile *artifact.KitFile) {
	m.Size = getModelSize(manifest)
	m.Author = getModelAuthor(kitfile)
	m.ModelName = getModelName(kitfile)
}

func getModelSize(manifest *ocispec.Manifest) string {
	var size int64
	for _, layer := range manifest.Layers {
		size += layer.Size
	}
	return output.FormatBytes(size)
}

func getModelAuthor(kitfile *artifact.KitFile) string {
	if kitfile != nil && len(kitfile.Package.Authors) > 0 {
		return kitfile.Package.Authors[0]
	} else {
		return "<none>"
	}
}

func getModelName(kitfile *artifact.KitFile) string {
	name := ""
	if kitfile != nil {
		name = kitfile.Package.Name
	}
	if name == "" {
		name = "<none>"
	}
	return name
}
