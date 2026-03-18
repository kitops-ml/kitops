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

package mlflow

import (
	"fmt"
	"os"
	"strings"
)

// ParsedURI holds the parsed components of a mlflow:// import URI.
type ParsedURI struct {
	// TrackingURI is the URL of the MLFlow tracking server.
	// Sourced from: host in the URI, MLFLOW_TRACKING_URI env var, or the default.
	TrackingURI  string
	ExperimentID string
	RunID        string
}

// ParseMLFlowURI parses URIs in the following forms:
//
//	mlflow://mlflow.company.com/experiments/42/runs/abc123
//	mlflow://experiments/42/runs/abc123   (uses MLFLOW_TRACKING_URI env var)
//	mlflow://runs/abc123                  (short form, uses MLFLOW_TRACKING_URI)
//
// If an explicit host is present it is used as the tracking URI (with http:// scheme).
// Otherwise MLFLOW_TRACKING_URI is consulted and finally defaultTrackingURI.
func ParseMLFlowURI(rawURI string) (*ParsedURI, error) {
	if !strings.HasPrefix(rawURI, "mlflow://") {
		return nil, fmt.Errorf("not an mlflow:// URI: %q", rawURI)
	}

	// Strip the scheme to get the authority + path portion
	rest := strings.TrimPrefix(rawURI, "mlflow://")

	// Determine whether the first segment looks like a host or a known keyword
	parts := strings.SplitN(rest, "/", 2)
	firstSeg := parts[0]

	var host, pathPart string
	if firstSeg == "experiments" || firstSeg == "runs" {
		// No host embedded, e.g.  mlflow://experiments/42/runs/abc123
		host = ""
		pathPart = rest
	} else {
		// Host is the first segment, e.g.  mlflow://mlflow.company.com/experiments/…
		host = firstSeg
		if len(parts) > 1 {
			pathPart = parts[1]
		}
	}

	trackingURI := resolveTrackingURI(host)

	expID, runID, err := parsePath(pathPart)
	if err != nil {
		return nil, fmt.Errorf("invalid mlflow:// URI %q: %w", rawURI, err)
	}

	return &ParsedURI{
		TrackingURI:  trackingURI,
		ExperimentID: expID,
		RunID:        runID,
	}, nil
}

// resolveTrackingURI turns a raw host into a full tracking URI, falling back to
// the MLFLOW_TRACKING_URI env var and finally the default.
func resolveTrackingURI(host string) string {
	if host != "" {
		if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
			return host
		}
		return "http://" + host
	}
	if envURI := os.Getenv("MLFLOW_TRACKING_URI"); envURI != "" {
		return envURI
	}
	return defaultTrackingURI
}

// parsePath extracts experimentID and runID from path strings of the forms:
//
//	runs/{run_id}
//	experiments/{exp_id}/runs/{run_id}
func parsePath(p string) (experimentID, runID string, err error) {
	p = strings.Trim(p, "/")
	segments := strings.Split(p, "/")

	switch {
	case len(segments) == 2 && segments[0] == "runs":
		// mlflow://runs/{run_id}
		return "", segments[1], nil

	case len(segments) == 4 && segments[0] == "experiments" && segments[2] == "runs":
		// mlflow://experiments/{exp_id}/runs/{run_id}
		return segments[1], segments[3], nil

	default:
		return "", "", fmt.Errorf(
			"unsupported path format %q: expected 'runs/{run_id}' or 'experiments/{exp_id}/runs/{run_id}'",
			p,
		)
	}
}
