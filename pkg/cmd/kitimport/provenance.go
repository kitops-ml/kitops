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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	BuildType = "https://kitops.org/import/v1"
	BuilderID = "https://kitops.org/kit"
)

// ProvenanceData is the internal aggregate populated as the import runs.
//
// We deliberately do not enumerate per-file resolved dependencies. The source
// commit digest pins the entire input tree byte-for-byte; the Kitfile config
// digest pins the build recipe; and the manifest digest (which cosign attests
// from the OCI ref) pins the output bytes via per-layer DiffIDs. That chain is
// sufficient for verification, so the SLSA-RECOMMENDED resolvedDependencies
// array is emitted empty rather than restating what those digests already cover.
type ProvenanceData struct {
	KitfileConfigDigest digest.Digest
	ManifestDigest      digest.Digest
	SourceURI           string
	SourceCommitSHA     string
	InvocationID        string
	StartedOn           time.Time
	FinishedOn          time.Time
	KitVersion          string
	GoVersion           string
}

// newProvenanceData constructs a ProvenanceData with the importer-invariant
// fields filled in (start time, invocation ID, builder version).
func newProvenanceData() *ProvenanceData {
	return &ProvenanceData{
		StartedOn:    time.Now().UTC(),
		InvocationID: newInvocationID(),
		KitVersion:   constants.Version,
		GoVersion:    constants.GoVersion,
	}
}

// finalize records the digests produced by packing and stamps the finish time.
// Call once, immediately after a successful packDirectory. The Kitfile config
// digest is recomputed here from the same MarshalToJSON() pack uses for the
// OCI config blob, so the predicate is byte-identical to what cosign attests.
func (p *ProvenanceData) finalize(manifestDesc *ocispec.Descriptor, kitfile *artifact.KitFile) error {
	cfgBytes, err := kitfile.MarshalToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal Kitfile for provenance: %w", err)
	}
	p.KitfileConfigDigest = digest.FromBytes(cfgBytes)
	p.ManifestDigest = manifestDesc.Digest
	p.FinishedOn = time.Now().UTC()
	return nil
}

// validate ensures every field a verifier needs has been populated. Emitting a
// predicate with an empty source digest, kitfile digest, or manifest digest
// produces an attestation that cannot be checked back to its inputs — that's
// worse than no attestation, so we refuse rather than silently emit it.
func (p *ProvenanceData) validate() error {
	if p.SourceURI == "" {
		return fmt.Errorf("provenance: source URI is empty")
	}
	if p.SourceCommitSHA == "" {
		return fmt.Errorf("provenance: source commit SHA missing for %s; predicate would be unverifiable", p.SourceURI)
	}
	if p.KitfileConfigDigest == "" {
		return fmt.Errorf("provenance: Kitfile config digest is empty")
	}
	if p.ManifestDigest == "" {
		return fmt.Errorf("provenance: manifest digest is empty")
	}
	return nil
}

// Wire-format predicate types mirror the SLSA Provenance v1 predicate body.
// The top-level is the predicate body (no _type/subject/predicateType wrapper);
// `cosign attest --predicate` derives the in-toto Statement subject from the OCI ref.

type Predicate struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

type BuildDefinition struct {
	BuildType            string               `json:"buildType"`
	ExternalParameters   ExternalParameters   `json:"externalParameters"`
	ResolvedDependencies []ResourceDescriptor `json:"resolvedDependencies"`
}

type ExternalParameters struct {
	Source  ResourceDescriptor `json:"source"`
	Kitfile ResourceDescriptor `json:"kitfile"`
}

type ResourceDescriptor struct {
	URI       string            `json:"uri,omitempty"`
	Digest    map[string]string `json:"digest"`
	MediaType string            `json:"mediaType,omitempty"`
}

type RunDetails struct {
	Builder  Builder  `json:"builder"`
	Metadata Metadata `json:"metadata"`
}

type Builder struct {
	ID      string            `json:"id"`
	Version map[string]string `json:"version"`
}

type Metadata struct {
	InvocationID string `json:"invocationId"`
	StartedOn    string `json:"startedOn"`
	FinishedOn   string `json:"finishedOn"`
}

// BuildPredicate maps the internal aggregate to the wire-format predicate.
func BuildPredicate(p *ProvenanceData) Predicate {
	source := ResourceDescriptor{
		URI:    p.SourceURI,
		Digest: map[string]string{},
	}
	if p.SourceCommitSHA != "" {
		// in-toto's digest set uses `gitCommit` for git commit IDs; the HF
		// X-Repo-Commit header is a git commit ID since HF repos are git-backed.
		// See https://github.com/in-toto/attestation/blob/main/spec/v1/digest_set.md
		source.Digest["gitCommit"] = p.SourceCommitSHA
	}

	kitfileDesc := ResourceDescriptor{
		Digest: map[string]string{
			"sha256": p.KitfileConfigDigest.Encoded(),
		},
	}

	return Predicate{
		BuildDefinition: BuildDefinition{
			BuildType: BuildType,
			ExternalParameters: ExternalParameters{
				Source:  source,
				Kitfile: kitfileDesc,
			},
			ResolvedDependencies: []ResourceDescriptor{},
		},
		RunDetails: RunDetails{
			Builder: Builder{
				ID: BuilderID,
				Version: map[string]string{
					"kit": p.KitVersion,
					"go":  p.GoVersion,
				},
			},
			Metadata: Metadata{
				InvocationID: p.InvocationID,
				StartedOn:    p.StartedOn.UTC().Format(time.RFC3339Nano),
				FinishedOn:   p.FinishedOn.UTC().Format(time.RFC3339Nano),
			},
		},
	}
}

// newInvocationID returns a UUIDv4 derived from 16 cryptographically-random bytes.
func newInvocationID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failures are catastrophic; fall back to a timestamp-based ID.
		return fmt.Sprintf("%016x-%016x", time.Now().UnixNano(), 0)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
