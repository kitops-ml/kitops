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

package output

// Logger receives all log output. Embedders implement this to redirect logs.
// When set via SetLogger, Kit bypasses its own level filtering and
// formatting — the Logger controls both.
// format+args follow fmt.Sprintf conventions and are passed through
// unformatted.
type Logger interface {
	Log(level LogLevel, format string, args ...any)
}
