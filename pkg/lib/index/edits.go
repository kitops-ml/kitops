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

package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"strings"
)

// KeyValueEdits is a set of values to set and keys to remove on an index entry's labels or
// annotations.
type KeyValueEdits[V any] struct {
	Set    map[string]V
	Remove []string
}

// labelKeyPattern matches Kubernetes-style qualified names, without their length limits: an
// optional DNS subdomain prefix and '/', followed by a name.
var labelKeyPattern = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?$`)

// keyValueArg is one parsed key=value, key:=json, or key- argument.
type keyValueArg struct {
	key, value     string
	isJSON, remove bool
}

func parseKeyValueArgs(args []string, kind string) ([]keyValueArg, error) {
	var parsed []keyValueArg
	seen := map[string]bool{}
	for _, arg := range args {
		var kv keyValueArg
		key, value, isAssignment := strings.Cut(arg, "=")
		switch {
		case isAssignment && strings.HasSuffix(key, ":"):
			kv = keyValueArg{key: strings.TrimSuffix(key, ":"), value: value, isJSON: true}
		case isAssignment:
			kv = keyValueArg{key: key, value: value}
		case strings.HasSuffix(arg, "-"):
			kv = keyValueArg{key: strings.TrimSuffix(arg, "-"), remove: true}
		default:
			return nil, fmt.Errorf("invalid %s %q: expected key=value to set, or key- to remove", kind, arg)
		}
		if kv.key == "" {
			return nil, fmt.Errorf("invalid %s %q: key cannot be empty", kind, arg)
		}
		if seen[kv.key] {
			return nil, fmt.Errorf("invalid %s %q: each key can be specified only once", kind, arg)
		}
		seen[kv.key] = true
		parsed = append(parsed, kv)
	}
	return parsed, nil
}

// ParseLabelEdits reads key=value (a string value), key:=json (any JSON value), and key-
// (remove) arguments into label edits.
func ParseLabelEdits(args []string) (KeyValueEdits[json.RawMessage], error) {
	parsed, err := parseKeyValueArgs(args, "label")
	if err != nil {
		return KeyValueEdits[json.RawMessage]{}, err
	}
	edits := KeyValueEdits[json.RawMessage]{Set: map[string]json.RawMessage{}}
	for _, kv := range parsed {
		if kv.remove {
			edits.Remove = append(edits.Remove, kv.key)
			continue
		}
		if !labelKeyPattern.MatchString(kv.key) {
			return KeyValueEdits[json.RawMessage]{}, fmt.Errorf("invalid label key %q: must be a name of letters, digits, '-', '_', and '.' that starts and ends with a letter or digit, optionally prefixed by a DNS subdomain and '/' (e.g. example.com/gpu)", kv.key)
		}
		if !kv.isJSON {
			value, err := json.Marshal(kv.value)
			if err != nil {
				return KeyValueEdits[json.RawMessage]{}, fmt.Errorf("failed to encode label %q: %w", kv.key, err)
			}
			edits.Set[kv.key] = value
			continue
		}
		var compacted bytes.Buffer
		if err := json.Compact(&compacted, []byte(kv.value)); err != nil {
			return KeyValueEdits[json.RawMessage]{}, fmt.Errorf("invalid label %s:=%s: value is not valid JSON", kv.key, kv.value)
		}
		edits.Set[kv.key] = compacted.Bytes()
	}
	return edits, nil
}

// ParseAnnotationEdits reads key=value and key- (remove) arguments into annotation edits.
// Annotation values are always strings.
func ParseAnnotationEdits(args []string) (KeyValueEdits[string], error) {
	parsed, err := parseKeyValueArgs(args, "annotation")
	if err != nil {
		return KeyValueEdits[string]{}, err
	}
	edits := KeyValueEdits[string]{Set: map[string]string{}}
	for _, kv := range parsed {
		switch {
		case kv.remove:
			edits.Remove = append(edits.Remove, kv.key)
		case kv.isJSON:
			return KeyValueEdits[string]{}, fmt.Errorf("invalid annotation %s:=%s: annotation values are strings; use key=value", kv.key, kv.value)
		default:
			edits.Set[kv.key] = kv.value
		}
	}
	return edits, nil
}

// ApplyEdits merges edits into current, returning nil rather than an empty map so that the
// field is omitted from the index.
func ApplyEdits[M ~map[string]V, V any](current M, edits KeyValueEdits[V]) M {
	merged := maps.Clone(current)
	if merged == nil {
		merged = M{}
	}
	maps.Copy(merged, edits.Set)
	for _, key := range edits.Remove {
		delete(merged, key)
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
