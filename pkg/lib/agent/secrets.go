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

package agent

import (
	"math"
	"regexp"
	"strings"
)

// This file decides, conservatively, whether a value from a developer's config
// is a credential that must be dropped rather than written into the Kitfile.

// credentialNameFragments mark a name as credential-shaped. Matched as
// lowercase substrings so that GITHUB_TOKEN, apiKey and DB_PASSWORD all hit.
var credentialNameFragments = []string{
	"token", "key", "secret", "pass", "pwd",
	"credential", "auth", "apikey", "access_id", "session",
	"signature", "private", "cert",
}

// credentialValuePrefixes are issuer prefixes that identify a value as a
// credential regardless of the name it is stored under or its entropy.
var credentialValuePrefixes = []string{
	"sk-", "sk_live_", "sk_test_", "pk_live_", "rk_live_",
	"ghp_", "gho_", "ghu_", "ghs_", "ghr_", "github_pat_",
	"xoxb-", "xoxp-", "xoxa-", "xoxs-", "xapp-",
	"glpat-", "gsk_", "AKIA", "ASIA", "ya29.",
	"AIza", "hf_", "dop_v1_", "npm_", "SG.", "shpat_",
	"Bearer ", "Basic ",
}

// secretCharset matches values made only of characters that appear in encoded
// credentials. A value containing anything else (spaces, braces, colons) is not
// treated as high-entropy.
var secretCharset = regexp.MustCompile(`^[A-Za-z0-9_\-+/=.~]+$`)

// credentialURLRegexp matches a URL carrying credentials in its userinfo, such
// as a database connection string. These are the most common way a password
// reaches a config file without a credential-shaped name to give it away, and
// they are excluded from the entropy test by the URL rule, so they are matched
// on their own.
var credentialURLRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.\-]*://[^/@\s]+:[^/@\s]*@`)

// envVarRefRegexp matches a value that is entirely a shell-style variable
// reference, e.g. "${GITHUB_TOKEN}" or "$GITHUB_TOKEN".
var envVarRefRegexp = regexp.MustCompile(`^\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?$`)

// minSecretLength is the shortest value considered high-entropy. Shorter values
// are matched by name or prefix only; a short random-looking string is more
// often a version or an id than a credential.
const minSecretLength = 20

// minSecretEntropy is the Shannon entropy per character, in bits, at which a
// value is treated as a credential. Encoded credentials sit well above 3.5;
// English words, identifiers and dotted paths sit below it.
const minSecretEntropy = 3.5

// EnvVarReference returns the variable name a value refers to, if the value is
// entirely a shell-style variable reference. A config that already points at an
// environment variable tells kit exactly which need to declare.
func EnvVarReference(value string) (string, bool) {
	match := envVarRefRegexp.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return "", false
	}
	return match[1], true
}

// IsCredentialName reports whether a name looks like it holds a credential.
func IsCredentialName(name string) bool {
	lower := strings.ToLower(name)
	for _, fragment := range credentialNameFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

// LooksLikeSecret reports whether a name/value pair must be dropped rather than
// recorded in the Kitfile: the name is credential-shaped, or the value carries
// a known issuer prefix or is long and high-entropy.
func LooksLikeSecret(name, value string) bool {
	if IsCredentialName(name) {
		return true
	}
	return IsSecretValue(value)
}

// IsSecretValue reports whether a value on its own looks like a credential.
func IsSecretValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, prefix := range credentialValuePrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	if credentialURLRegexp.MatchString(trimmed) {
		return true
	}
	return isHighEntropy(trimmed)
}

// isHighEntropy applies the length, shape and entropy tests that identify an
// encoded credential. Each early return excludes a class of value that is long
// and varied but is plainly configuration: paths, URLs, and lists.
func isHighEntropy(value string) bool {
	if len(value) < minSecretLength {
		return false
	}
	// Paths and URLs are configuration, however long.
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "./") ||
		strings.HasPrefix(value, "../") || strings.HasPrefix(value, "~") ||
		strings.Contains(value, "://") {
		return false
	}
	// A value made only of characters that appear in encoded credentials; a
	// space, comma or brace means it is prose, a list, or a template.
	if !secretCharset.MatchString(value) {
		return false
	}
	// Encoded credentials mix letters and digits. An all-letter value of any
	// length is a name, a package, or a dotted identifier.
	var hasLetter, hasDigit bool
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return false
	}
	return shannonEntropy(value) >= minSecretEntropy
}

// shannonEntropy returns the Shannon entropy of s in bits per character.
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}
	counts := map[rune]int{}
	total := 0
	for _, r := range s {
		counts[r]++
		total++
	}
	entropy := 0.0
	for _, count := range counts {
		p := float64(count) / float64(total)
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// SplitAuthScheme separates a leading auth scheme from a header value, so that
// "Bearer ghp_xxx" can be rewritten as "Bearer ${needs.NAME}" with the scheme
// preserved and only the credential dropped.
func SplitAuthScheme(value string) (scheme, credential string) {
	for _, s := range []string{"Bearer ", "Basic ", "Token ", "token "} {
		if rest, found := strings.CutPrefix(value, s); found {
			return s, rest
		}
	}
	return "", value
}
