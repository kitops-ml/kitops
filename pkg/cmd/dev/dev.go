// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem"
	"github.com/kitops-ml/kitops/pkg/lib/harness"
	kfutils "github.com/kitops-ml/kitops/pkg/lib/kitfile"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

func runDev(ctx context.Context, options *DevStartOptions) error {

	kitfile := &artifact.KitFile{}

	modelfile, err := os.Open(options.modelFile)
	if err != nil {
		return err
	}
	defer modelfile.Close()
	if err := kitfile.LoadModel(modelfile); err != nil {
		return err
	}
	output.Infof("Loaded Kitfile: %s", options.modelFile)
	if util.IsModelKitReference(kitfile.Model.Path) {
		resolvedKitfile, err := kfutils.ResolveKitfile(ctx, options.configHome, kitfile.Model.Path, kitfile.Model.Path)
		if err != nil {
			return fmt.Errorf("failed to resolve referenced modelkit %s: %w", kitfile.Model.Path, err)
		}
		kitfile.Model.Path = resolvedKitfile.Model.Path
		kitfile.Model.Parts = append(kitfile.Model.Parts, resolvedKitfile.Model.Parts...)
	}
	modelPath, _, err := filesystem.VerifySubpath(options.contextDir, kitfile.Model.Path)
	if err != nil {
		return err
	}

	llmHarness := &harness.LLMHarness{}
	llmHarness.Host = options.host
	llmHarness.Port = options.port
	llmHarness.ConfigHome = options.configHome
	if err := llmHarness.Init(); err != nil {
		return err
	}

	if err := llmHarness.Start(modelPath); err != nil {
		return err
	}

	return nil
}

func stopDev(_ context.Context, options *DevBaseOptions) error {

	llmHarness := &harness.LLMHarness{}
	llmHarness.ConfigHome = options.configHome

	if err := llmHarness.Init(); err != nil {
		return err
	}

	if err := llmHarness.Stop(); err != nil {
		return err
	}
	return nil
}
