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

package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	modelspecv1 "github.com/modelpack/model-spec/specs-go/v1"
	"github.com/opencontainers/go-digest"
	"go.yaml.in/yaml/v3"
)

const modelPartTypeMaxLen = 64

var modelPartTypeRegexp = regexp.MustCompile(`^[\w][\w.-]*$`)

// PathType represents different types of paths in Kitfile fields, e.g. local on-disk paths or remote S3 URLs.
type PathType int

const (
	UnknownPathType PathType = iota
	LocalPathType
	ModelReferencePathType
	S3PathType
)

type (
	KitFile struct {
		ManifestVersion string    `json:"manifestVersion" yaml:"manifestVersion"`
		Package         Package   `json:"package,omitempty" yaml:"package,omitempty"`
		Model           *Model    `json:"model,omitempty" yaml:"model,omitempty"`
		Code            []Code    `json:"code,omitempty" yaml:"code,omitempty"`
		DataSets        []DataSet `json:"datasets,omitempty" yaml:"datasets,omitempty"`
		Docs            []Docs    `json:"docs,omitempty" yaml:"docs,omitempty"`
		Prompts         []Prompt  `json:"prompts,omitempty" yaml:"prompts,omitempty"`
	}

	Package struct {
		Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
		Version     string   `json:"version,omitempty" yaml:"version,omitempty"`
		Description string   `json:"description,omitempty" yaml:"description,omitempty"`
		License     string   `json:"license,omitempty" yaml:"license,omitempty"`
		Authors     []string `json:"authors,omitempty" yaml:"authors,omitempty,flow"`
	}

	Docs struct {
		Path        string `json:"path" yaml:"path"`
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		*LayerInfo  `json:",inline" yaml:",inline"`
	}

	Model struct {
		Name        string      `json:"name,omitempty" yaml:"name,omitempty"`
		Path        string      `json:"path,omitempty" yaml:"path,omitempty"`
		License     string      `json:"license,omitempty" yaml:"license,omitempty"`
		Framework   string      `json:"framework,omitempty" yaml:"framework,omitempty"`
		Format      string      `json:"format,omitempty" yaml:"format,omitempty"`
		Version     string      `json:"version,omitempty" yaml:"version,omitempty"`
		Description string      `json:"description,omitempty" yaml:"description,omitempty"`
		Parts       []ModelPart `json:"parts,omitempty" yaml:"parts,omitempty"`
		// Parameters is an arbitrary section of yaml that can be used to store any additional
		// data that may be relevant to the current model, with a few caveats:
		//  * Only a json-compatible subset of yaml is supported
		//  * Strings will be serialized without flow parameters, etc.
		//  * Numbers will be converted to decimal representations (0xFF -> 255, 1.2e+3 -> 1200)
		//  * Maps will be sorted alphabetically by key
		Parameters any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		*LayerInfo `json:",inline" yaml:",inline"`
	}

	ModelPart struct {
		Name       string `json:"name,omitempty" yaml:"name,omitempty"`
		Path       string `json:"path,omitempty" yaml:"path,omitempty"`
		License    string `json:"license,omitempty" yaml:"license,omitempty"`
		Type       string `json:"type,omitempty" yaml:"type,omitempty"`
		*LayerInfo `json:",inline" yaml:",inline"`
	}

	Code struct {
		Path        string `json:"path,omitempty" yaml:"path,omitempty"`
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		License     string `json:"license,omitempty" yaml:"license,omitempty"`
		*LayerInfo  `json:",inline" yaml:",inline"`
	}

	DataSet struct {
		Name        string `json:"name,omitempty" yaml:"name,omitempty"`
		Path        string `json:"path,omitempty" yaml:"path,omitempty"`
		RemotePath  string `json:"remotePath,omitempty" yaml:"remotePath,omitempty"`
		RemoteHash  string `json:"remoteHash,omitempty" yaml:"remoteHash,omitempty"`
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		License     string `json:"license,omitempty" yaml:"license,omitempty"`
		// Parameters is an arbitrary section of yaml that can be used to store any additional
		// metadata relevant to the dataset, with a few caveats:
		//  * Only a json-compatible subset of yaml is supported
		//  * Strings will be serialized without flow parameters, etc.
		//  * Numbers will be converted to decimal representations
		//  * Maps will be sorted alphabetically by key
		//  * It's recommended to store metadata like preprocessing steps, formats, etc.
		Parameters any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		*LayerInfo `json:",inline" yaml:",inline"`
	}

	Prompt struct {
		Name        string `json:"name,omitempty" yaml:"name,omitempty"`
		Path        string `json:"path,omitempty" yaml:"path,omitempty"`
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		*LayerInfo  `json:",inline" yaml:",inline"`
	}

	LayerInfo struct {
		// Digest for the layer corresponding to this element
		Digest string `json:"digest,omitempty" yaml:"-"`
		// Diff ID (uncompressed digest) for the layer corresponding to this element
		DiffId string `json:"diffId,omitempty" yaml:"-"`
	}
)

func (kf *KitFile) LoadModel(kitfileContent io.ReadCloser) error {
	decoder := yaml.NewDecoder(kitfileContent)
	decoder.KnownFields(true)
	if err := decoder.Decode(kf); err != nil {
		return err
	}
	if err := kf.Validate(); err != nil {
		return err
	}
	return nil
}

func (kf *KitFile) MarshalToJSON() ([]byte, error) {
	jsonData, err := json.Marshal(kf)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (kf *KitFile) MarshalToYAML() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(kf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (kf *KitFile) Validate() error {
	var errs []string
	addErr := func(format string, a ...any) {
		s := fmt.Sprintf(format, a...)
		errs = append(errs, fmt.Sprintf("  * %s", s))
	}
	if kf.ManifestVersion != "1.0.0" {
		addErr("invalid manifestVersion: expect 1.0.0 but got %s", kf.ManifestVersion)
	}

	// Map of paths to the component that uses them; used to detect duplicate paths
	paths := map[string][]string{}
	addPath := func(path, source string) {
		if path == "" {
			path = "."
		}
		path = filepath.Clean(path)
		paths[path] = append(paths[path], source)
	}

	if kf.Model != nil {
		addPath(kf.Model.Path, fmt.Sprintf("model %s", kf.Model.Name))
		modelPathType, err := GetPathType(kf.Model.Path)
		if err != nil {
			addErr("invalid path for model: %s", err)
		}
		if modelPathType != LocalPathType && modelPathType != ModelReferencePathType {
			addErr("invalid path for model: only local paths and ModelKit references are permitted")
		}
		for _, part := range kf.Model.Parts {
			addPath(part.Path, fmt.Sprintf("modelpart %s", part.Name))
			partPathType, err := GetPathType(part.Path)
			if err != nil {
				addErr("invalid path for model part (%s): %s", part.Path, err)
			}
			if partPathType != LocalPathType {
				addErr("invalid path for model part (%s): only local paths are permitted", part.Path)
			}
			if part.Type != "" {
				if !modelPartTypeRegexp.MatchString(part.Type) {
					addErr("modelpart %s has invalid type (must be alphanumeric with dots, dashes, and underscores)", part.Name)
				}
				if len(part.Type) > modelPartTypeMaxLen {
					addErr("modelpart %s type is too long (must be fewer than %d characters)", part.Name, modelPartTypeMaxLen)
				}
			}
		}
	}
	for idx, code := range kf.Code {
		addPath(code.Path, fmt.Sprintf("code layer %d", idx))
		pathType, err := GetPathType(code.Path)
		if err != nil {
			addErr("invalid path for code (%s): %s", code.Path, err)
		}
		if pathType != LocalPathType {
			addErr("invalid path for code (%s): only local paths are permitted", code.Path)
		}
	}
	for idx, dataset := range kf.DataSets {
		addPath(dataset.Path, fmt.Sprintf("dataset layer %d", idx))
		pathType, err := GetPathType(dataset.Path)
		if err != nil {
			addErr("invalid path for dataset (%s): %s", dataset.Path, err)
		}
		if pathType != LocalPathType {
			addErr("invalid path for dataset (%s): only local paths are permitted", dataset.Path)
		}
		if dataset.RemotePath != "" {
			remotePathType, err := GetPathType(dataset.RemotePath)
			if err != nil {
				addErr("invalid remote path for dataset (%s): %s", dataset.RemotePath, err)
			}
			if remotePathType != S3PathType {
				addErr("only S3 URLs are supported for remote dataset paths (%s)", dataset.RemotePath)
			}
			if dataset.RemoteHash == "" {
				addErr("remoteHash is required when remote dataset paths are used (%s)", dataset.RemotePath)
			}
		} else {
			if dataset.RemoteHash != "" {
				addErr("remote hash is only applicable when remotePath is set")
			}
		}
	}
	for idx, doc := range kf.Docs {
		addPath(doc.Path, fmt.Sprintf("docs layer %d", idx))
		pathType, err := GetPathType(doc.Path)
		if err != nil {
			addErr("invalid path for doc (%s): %s", doc.Path, err)
		}
		if pathType != LocalPathType {
			addErr("invalid path for doc (%s): only local paths are permitted", doc.Path)
		}
	}
	for _, prompt := range kf.Prompts {
		pathType, err := GetPathType(prompt.Path)
		if err != nil {
			addErr("invalid path for prompt (%s): %s", prompt.Path, err)
		}
		if pathType != LocalPathType {
			addErr("invalid path for prompt (%s): only local paths are permitted", prompt.Path)
		}
	}

	for layerPath, layerIds := range paths {
		if len := len(layerIds); len > 1 {
			addErr("%s and %s use the same path %s", strings.Join(layerIds[:len-1], ", "), layerIds[len-1], layerPath)
		}
		if path.IsAbs(layerPath) || filepath.IsAbs(layerPath) {
			addErr("absolute paths are not supported in a Kitfile (path %s in %s)", layerPath, layerIds[0])
		}
	}

	if len(errs) > 0 {
		// Iterating through the paths map is random; sort to get a consistent message
		slices.Sort(errs)
		return fmt.Errorf("errors while validating Kitfile: \n%s", strings.Join(errs, "\n"))
	}

	return nil
}

func (kf *KitFile) ToModelPackConfig(diffIDs []digest.Digest) modelspecv1.Model {
	// Fill fields as best as we can, depending on what's available in the Kitfile
	now := time.Now()

	modelDescriptor := modelspecv1.ModelDescriptor{
		CreatedAt:   &now,
		Authors:     kf.Package.Authors,
		Name:        kf.Package.Name,
		Version:     kf.Package.Version,
		Licenses:    kf.collectLicenses(),
		Title:       kf.Package.Name,
		Description: kf.Package.Description,
	}

	modelFS := modelspecv1.ModelFS{
		Type:    "layers",
		DiffIDs: diffIDs,
	}

	modelConfig := modelspecv1.ModelConfig{}
	if kf.Model != nil {
		modelConfig.Format = kf.Model.Format
	}

	model := modelspecv1.Model{
		Descriptor: modelDescriptor,
		ModelFS:    modelFS,
		Config:     modelConfig,
	}

	return model
}

func (kf *KitFile) collectLicenses() []string {
	var licenses []string
	appendNotEmpty := func(l []string, s string) []string {
		if s == "" {
			return l
		}
		return append(l, s)
	}
	licenses = appendNotEmpty(licenses, kf.Package.License)
	if kf.Model != nil {
		licenses = appendNotEmpty(licenses, kf.Model.License)
		for _, modelpart := range kf.Model.Parts {
			licenses = appendNotEmpty(licenses, modelpart.License)
		}
	}
	for _, ds := range kf.DataSets {
		licenses = appendNotEmpty(licenses, ds.License)
	}
	for _, code := range kf.Code {
		licenses = appendNotEmpty(licenses, code.License)
	}
	slices.Sort(licenses)
	licenses = slices.Compact(licenses)

	return licenses
}

func GetPathType(path string) (PathType, error) {
	if IsModelKitReference(path) {
		return ModelReferencePathType, nil
	}

	// Treat things that don't parse as URLs as local paths
	parsed, err := url.Parse(path)
	if err != nil || (parsed.Scheme == "" && parsed.Host == "") {
		return LocalPathType, nil
	}

	switch parsed.Scheme {
	case "s3":
		return S3PathType, nil
	case "http", "https":
		return UnknownPathType, fmt.Errorf("HTTP urls are not supported in paths (%s)", path)
	}

	return UnknownPathType, fmt.Errorf("unrecognized URL in path: %s", path)
}
