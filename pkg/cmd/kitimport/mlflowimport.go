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

package kitimport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/filesystem/cache"
	kfgen "github.com/kitops-ml/kitops/pkg/lib/kitfile/generate"
	mlflowclient "github.com/kitops-ml/kitops/pkg/lib/mlflow"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
)

// importUsingMLFlow implements the mlflow:// import flow:
//  1. Parse the mlflow:// URI
//  2. Validate run status (only FINISHED runs are importable)
//  3. Download artifacts from the tracking server
//  4. Generate / use provided Kitfile
//  5. Augment Kitfile with MLFlow provenance
//  6. Pack the ModelKit
func importUsingMLFlow(ctx context.Context, opts *importOptions) error {
	parsedURI, err := mlflowclient.ParseMLFlowURI(opts.repo)
	if err != nil {
		return err
	}

	token := opts.token
	if token == "" {
		token = os.Getenv("MLFLOW_TRACKING_TOKEN")
	}

	client := mlflowclient.NewClient(parsedURI.TrackingURI, token)

	output.Infof("Connecting to MLFlow tracking server %s", parsedURI.TrackingURI)

	run, err := client.GetRun(ctx, parsedURI.RunID)
	if err != nil {
		return fmt.Errorf("failed to fetch MLFlow run %s: %w", parsedURI.RunID, err)
	}

	// Only import completed runs to avoid partial artifacts
	if run.Info.Status != mlflowclient.RunStatusFinished {
		return fmt.Errorf(
			"MLFlow run %s has status %q — only %s runs can be imported",
			parsedURI.RunID, run.Info.Status, mlflowclient.RunStatusFinished,
		)
	}

	output.Infof("Found run %q (experiment %s) — status %s", run.Info.RunName, run.Info.ExperimentID, run.Info.Status)

	// List all artifact files for this run
	artifacts, err := client.ListArtifacts(ctx, parsedURI.RunID, "")
	if err != nil {
		return fmt.Errorf("failed to list artifacts for run %s: %w", parsedURI.RunID, err)
	}
	if len(artifacts) == 0 {
		return fmt.Errorf("MLFlow run %s has no artifacts — nothing to import", parsedURI.RunID)
	}
	output.Infof("Found %d artifact file(s) for run %s", len(artifacts), parsedURI.RunID)

	// Create temp directory for downloads
	tmpDir, cleanupTmp, err := cache.MkCacheDir(cache.CacheImportSubdir, "")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	doCleanup := true
	defer func() {
		if doCleanup {
			cleanupTmp()
		}
	}()

	// Download artifacts to temp dir
	if err := mlflowclient.DownloadArtifacts(ctx, client, parsedURI.RunID, tmpDir, artifacts, opts.concurrency); err != nil {
		return fmt.Errorf("failed to download MLFlow artifacts: %w", err)
	}

	// Build the directory listing from the downloaded files
	dirListing, err := kfgen.DirectoryListingFromFS(tmpDir)
	if err != nil {
		return fmt.Errorf("error processing downloaded artifacts: %w", err)
	}

	// Determine the Kitfile to use
	var kitfile *artifact.KitFile
	switch {
	case opts.kitfilePath == "-":
		kitfile = &artifact.KitFile{}
		if err := kitfile.LoadModel(os.Stdin); err != nil {
			return fmt.Errorf("failed to read Kitfile from stdin: %w", err)
		}
	case opts.kitfilePath != "":
		kf, err := readExistingKitfile(opts.kitfilePath)
		if err != nil {
			return err
		}
		kitfile = kf
	default:
		// Auto-generate Kitfile from the downloaded artifacts
		repoLabel := buildRepoLabel(run, parsedURI)
		kf, err := generateKitfile(dirListing, repoLabel, tmpDir)
		if err != nil {
			return err
		}
		kitfile = kf

		if util.IsInteractiveSession() {
			newKitfile, err := promptToEditKitfile(tmpDir, kf)
			if err != nil {
				if errors.Is(err, ErrNoEditorFound) {
					doCleanup = false
					kfPath := filepath.Join(tmpDir, constants.DefaultKitfileName)
					output.Logf(output.LogLevelWarn, "Could not determine default editor from $EDITOR environment variable")
					output.Logf(output.LogLevelWarn, "Please manually edit Kitfile at path")
					output.Logf(output.LogLevelWarn, "    %s", kfPath)
					output.Logf(output.LogLevelWarn, "and run command")
					output.Logf(output.LogLevelWarn, "    kit import %s -t %s -f %s", opts.repo, opts.tag, kfPath)
					output.Logf(output.LogLevelWarn, "to complete process")
					return err
				}
				return err
			}
			kitfile = newKitfile
		}
	}

	// Enrich Kitfile with MLFlow provenance metadata
	injectMLFlowProvenance(kitfile, run, parsedURI)

	// Re-write the Kitfile with provenance so it is packed into the ModelKit
	kitfileBytes, err := kitfile.MarshalToYAML()
	if err != nil {
		return fmt.Errorf("failed to serialize Kitfile: %w", err)
	}
	kitfilePath := filepath.Join(tmpDir, constants.DefaultKitfileName)
	if err := os.WriteFile(kitfilePath, kitfileBytes, 0644); err != nil {
		return fmt.Errorf("failed to write Kitfile: %w", err)
	}
	output.Infof("Updated Kitfile with MLFlow provenance:\n\n%s\n", string(kitfileBytes))

	// Pack
	output.Infof("Packing model to %s", opts.tag)
	if err := packDirectory(ctx, opts.configHome, tmpDir, kitfile, opts.modelKitRef); err != nil {
		return fmt.Errorf("failed to pack ModelKit: %w", err)
	}
	output.Infof("Model is packed as %s", opts.tag)

	if err := cache.CleanCacheDir(cache.CacheImportSubdir); err != nil {
		output.Logf(output.LogLevelWarn, "Failed to clean cache directory: %s", err)
	}

	return nil
}

// buildRepoLabel builds a human-friendly repository label used to populate
// the package metadata when generating a Kitfile automatically.
func buildRepoLabel(run *mlflowclient.Run, parsedURI *mlflowclient.ParsedURI) string {
	name := run.Info.RunName
	if name == "" {
		name = run.Info.RunID
	}
	expID := run.Info.ExperimentID
	if expID == "" {
		expID = parsedURI.ExperimentID
	}
	if expID == "" {
		return "mlflow/" + name
	}
	return fmt.Sprintf("mlflow-exp%s/%s", expID, name)
}

// injectMLFlowProvenance enriches the kitfile's package and model sections with
// metadata sourced from the MLFlow run so that provenance is captured in the ModelKit.
func injectMLFlowProvenance(kf *artifact.KitFile, run *mlflowclient.Run, parsedURI *mlflowclient.ParsedURI) {
	// Populate Package section if sparse
	if kf.Package.Name == "" {
		name := run.Info.RunName
		if name == "" {
			name = run.Info.RunID
		}
		kf.Package.Name = sanitizeName(name)
	}
	if kf.Package.Description == "" {
		kf.Package.Description = buildDescription(run, parsedURI)
	}
	if kf.Package.Version == "" && run.Info.EndTime > 0 {
		t := time.UnixMilli(run.Info.EndTime).UTC()
		kf.Package.Version = t.Format("20060102150405")
	}

	// Embed provenance into model parameters so it is queryable later
	if kf.Model != nil {
		provenance := buildProvenance(run, parsedURI)
		switch existing := kf.Model.Parameters.(type) {
		case nil:
			kf.Model.Parameters = provenance
		case map[string]any:
			for k, v := range provenance {
				existing[k] = v
			}
			kf.Model.Parameters = existing
		}
	}
}

// buildDescription generates a human-readable description from run metadata.
func buildDescription(run *mlflowclient.Run, parsedURI *mlflowclient.ParsedURI) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Imported from MLFlow run %s", run.Info.RunID))
	if run.Info.RunName != "" {
		parts = append(parts, fmt.Sprintf("(%s)", run.Info.RunName))
	}
	if parsedURI.ExperimentID != "" {
		parts = append(parts, fmt.Sprintf("experiment %s", parsedURI.ExperimentID))
	}
	parts = append(parts, fmt.Sprintf("tracking server %s", parsedURI.TrackingURI))
	return strings.Join(parts, " — ")
}

// buildProvenance assembles the mlflow_provenance map written into model.parameters.
func buildProvenance(run *mlflowclient.Run, parsedURI *mlflowclient.ParsedURI) map[string]any {
	p := map[string]any{
		"mlflow_tracking_uri":  parsedURI.TrackingURI,
		"mlflow_run_id":        run.Info.RunID,
		"mlflow_run_name":      run.Info.RunName,
		"mlflow_experiment_id": run.Info.ExperimentID,
		"mlflow_run_status":    run.Info.Status,
	}
	if run.Info.StartTime > 0 {
		p["mlflow_start_time"] = time.UnixMilli(run.Info.StartTime).UTC().Format(time.RFC3339)
	}
	if run.Info.EndTime > 0 {
		p["mlflow_end_time"] = time.UnixMilli(run.Info.EndTime).UTC().Format(time.RFC3339)
	}

	// Capture logged metrics as provenance
	if len(run.Data.Metrics) > 0 {
		metrics := make(map[string]any, len(run.Data.Metrics))
		for _, m := range run.Data.Metrics {
			metrics[m.Key] = m.Value
		}
		p["mlflow_metrics"] = metrics
	}

	// Capture logged params as provenance
	if len(run.Data.Params) > 0 {
		params := make(map[string]any, len(run.Data.Params))
		for _, param := range run.Data.Params {
			params[param.Key] = param.Value
		}
		p["mlflow_params"] = params
	}

	return map[string]any{"mlflow_provenance": p}
}

// sanitizeName converts arbitrary run names into valid package names.
func sanitizeName(name string) string {
	// Replace characters that are likely invalid in a package name
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "@", "-")
	return strings.ToLower(replacer.Replace(name))
}
