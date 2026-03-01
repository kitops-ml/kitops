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
	"fmt"
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

var (
	colorNone  = "\033[0m"
	colorTrace = "\033[0m"    // No color
	colorDebug = "\033[0;34m" // Blue
	colorInfo  = "\033[0;32m" // Green
	colorWarn  = "\033[0;93m" // Yellow
	colorError = "\033[0;31m" // Red
)

func init() {
	if runtime.GOOS == "windows" || !term.IsTerminal(int(os.Stdout.Fd())) {
		colorNone = ""
		colorError = ""
		colorWarn = ""
		colorDebug = ""
		colorTrace = ""
	}
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
	if logLevel == LogLevelInfo {
		switch l {
		case LogLevelInfo, LogLevelSystem:
			return ""
		case LogLevelWarn:
			return fmt.Sprintf("%s[WARN ] %s", colorWarn, colorNone)
		case LogLevelError:
			return fmt.Sprintf("%s[ERROR] %s", colorError, colorNone)
		}
	} else {
		switch l {
		case LogLevelTrace:
			return fmt.Sprintf("%s[TRACE] %s", colorTrace, colorNone)
		case LogLevelDebug:
			return fmt.Sprintf("%s[DEBUG] %s", colorDebug, colorNone)
		case LogLevelInfo, LogLevelSystem:
			return fmt.Sprintf("%s[INFO ] %s", colorInfo, colorNone)
		case LogLevelWarn:
			return fmt.Sprintf("%s[WARN ] %s", colorWarn, colorNone)
		case LogLevelError:
			return fmt.Sprintf("%s[ERROR] %s", colorError, colorNone)
		}
	}
	return ""
}
