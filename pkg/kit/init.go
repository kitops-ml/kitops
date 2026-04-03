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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/external/hf"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
)

type InitOptions struct {
	ConfigHome  string
	Path        string // local directory or remote repo URL
	Remote      bool
	RepoRef     string // branch or tag for remote; defaults to "main"
	Token       string // auth token for remote
	OutputPath  string // file path to write; "-" = return bytes only; "" = default
	Overwrite   bool
	Name        string
	Description string
	Author      string
}

type InitResult struct {
	KitfileBytes []byte
	WrittenPath  string // populated when written to disk
}

func Init(ctx context.Context, opts *InitOptions) (*InitResult, error) {
	if opts.RepoRef == "" {
		opts.RepoRef = "main"
	}

	var dirContents *kfgen.DirectoryListing
	var repo string
	if opts.Remote {
		parsedRepo, repoType, err := hf.ParseHuggingFaceRepo(opts.Path)
		if err != nil {
			return nil, fmt.Errorf("invalid HuggingFace repository: %w", err)
		}
		repo = parsedRepo
		dirContents, err = hf.ListFiles(ctx, repo, opts.RepoRef, opts.Token, repoType)
		if err != nil {
			return nil, fmt.Errorf("error fetching remote repository: %w", err)
		}
	} else {
		var err error
		dirContents, err = kfgen.DirectoryListingFromFS(opts.Path)
		if err != nil {
			return nil, fmt.Errorf("error processing directory: %w", err)
		}
	}

	modelPackage := BuildPackage(repo, opts.Name, opts.Description, opts.Author)
	kitfile, err := kfgen.GenerateKitfile(dirContents, modelPackage)
	if err != nil {
		return nil, fmt.Errorf("error generating Kitfile: %w", err)
	}
	bytes, err := kitfile.MarshalToYAML()
	if err != nil {
		return nil, fmt.Errorf("error formatting Kitfile: %w", err)
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		if opts.Remote {
			outputPath = "-"
		} else {
			outputPath = filepath.Join(opts.Path, constants.DefaultKitfileName)
		}
	}

	if outputPath == "-" {
		return &InitResult{KitfileBytes: bytes}, nil
	}

	if _, err := os.Stat(outputPath); err == nil {
		if !opts.Overwrite {
			return nil, fmt.Errorf("Kitfile already exists at %s; set Overwrite to replace it", outputPath)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error checking for existing Kitfile: %w", err)
	}
	if err := os.WriteFile(outputPath, bytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write Kitfile: %w", err)
	}
	return &InitResult{KitfileBytes: bytes, WrittenPath: outputPath}, nil
}

func BuildPackage(repo, name, description, author string) *artifact.Package {
	sections := strings.Split(repo, "/")
	pkg := &artifact.Package{}
	if name != "" {
		pkg.Name = name
	} else if len(sections) >= 2 {
		pkg.Name = sections[len(sections)-1]
	}
	if description != "" {
		pkg.Description = description
	}
	if author != "" {
		pkg.Authors = append(pkg.Authors, author)
	} else if len(sections) >= 2 {
		pkg.Authors = append(pkg.Authors, sections[len(sections)-2])
	}
	return pkg
}
