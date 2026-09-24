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

package output

import (
	"io"
	"os"
	"runtime"

	"golang.org/x/term"
)

type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelSystem // always printed, routed to stderr
)

const colorNone = "\033[0m"

var levelPrefixes = map[LogLevel]struct {
	label string
	color string
}{
	LogLevelTrace:  {"[TRACE]", "\033[0m"},    // No color
	LogLevelDebug:  {"[DEBUG]", "\033[0;34m"}, // Blue
	LogLevelInfo:   {"[INFO ]", "\033[0;32m"}, // Green
	LogLevelWarn:   {"[WARN ]", "\033[0;93m"}, // Yellow
	LogLevelError:  {"[ERROR]", "\033[0;31m"}, // Red
	LogLevelSystem: {"[INFO ]", "\033[0;32m"}, // Green
}

// shouldUseColor reports whether color codes can be written to w. Writers that
// are not an *os.File (e.g. those set via SetOut) are never a terminal.
func shouldUseColor(w io.Writer) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (l LogLevel) shouldPrint(atLevel LogLevel) bool {
	return l <= atLevel
}

func (l LogLevel) getOutput() io.Writer {
	switch l {
	case LogLevelError, LogLevelWarn, LogLevelSystem:
		return stderr
	default:
		return stdout
	}
}

func (l LogLevel) getPrefix() string {
	// At the default log level, only warnings and errors are prefixed.
	if logLevel == LogLevelInfo && l != LogLevelWarn && l != LogLevelError {
		return ""
	}
	prefix, ok := levelPrefixes[l]
	if !ok {
		return ""
	}
	if !shouldUseColor(l.getOutput()) {
		return prefix.label + " "
	}
	return prefix.color + prefix.label + " " + colorNone
}
