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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/vbauerster/mpb/v8"
)

func Infoln(s any) {
	Logln(LogLevelInfo, s)
}

func Infof(s string, args ...any) {
	Logf(LogLevelInfo, s, args...)
}

func Errorln(s any) {
	Logln(LogLevelError, s)
}

func Errorf(s string, args ...any) {
	Logf(LogLevelError, s, args...)
}

// Fatalln is the equivalent of Errorln except it returns a basic error to signal the command has failed
func Fatalln(s any) error {
	Logln(LogLevelError, s)
	return errors.New("failed to run")
}

// Fatalf is the equivalent of Errorf except it returns a basic error to signal the command has failed
func Fatalf(s string, args ...any) error {
	Logf(LogLevelError, s, args...)
	return errors.New("failed to run")
}

func Debugln(s any) {
	Logln(LogLevelDebug, s)
}

func Debugf(s string, args ...any) {
	Logf(LogLevelDebug, s, args...)
}

// SafeDebugln is the same as Debugln except it will only print if progress bars
// are disabled to avoid confusing output
func SafeDebugln(s any) {
	SafeLogln(LogLevelDebug, s)
}

// SafeDebugf is the same as Debugf except it will only print if progress bars
// are disabled to avoid confusing output
func SafeDebugf(s string, args ...any) {
	SafeLogf(LogLevelDebug, s, args...)
}

// SystemInfoLn is like Infoln except it logs to stderr. This should be used
// for system messages such as update notifications
func SystemInfoln(s any) {
	loglnTo(stderr, LogLevelInfo, s)
}

// SystemInfof is like Infof except it logs to stderr. This should be used
// for system messages such as update notifications
func SystemInfof(s string, args ...any) {
	logfTo(stderr, LogLevelInfo, s, args)
}

func Logln(level LogLevel, s any) {
	loglnTo(level.getOutput(), level, s)
}

func loglnTo(output io.Writer, level LogLevel, s any) {
	if logLevel.shouldPrint(level) {
		str := fmt.Sprintln(s)
		// Capitalize first letter in string for nicer output, in case it's not already capitalized
		str = strings.ToUpper(str[:1]) + str[1:]
		str = level.getPrefix() + str
		fmt.Fprint(output, str)
	}
}

func Logf(level LogLevel, s string, args ...any) {
	logfTo(level.getOutput(), level, s, args...)
}

func logfTo(output io.Writer, level LogLevel, s string, args ...any) {
	if logLevel.shouldPrint(level) {
		// Avoid printing incomplete lines
		if !strings.HasSuffix(s, "\n") {
			s = s + "\n"
		}
		str := fmt.Sprintf(s, args...)
		// Capitalize first letter in string for nicer output, in case it's not already capitalized
		str = strings.ToUpper(str[:1]) + str[1:]
		str = level.getPrefix() + str
		fmt.Fprint(output, str)
	}
}

func SafeLogln(level LogLevel, s any) {
	if !progressEnabled {
		Logln(level, s)
	}
}

func SafeLogf(level LogLevel, s string, args ...any) {
	if !progressEnabled {
		Logf(level, s, args...)
	}
}

// ProgressLogger allows for printing info and debug lines while a progress bar
// is filling, and should be used instead of the standard output functions to prevent
// progress bars from removing log lines. Once the progress bar is done, the Wait()
// method should be called.
type ProgressLogger struct {
	output io.Writer
}

// Wait will call Wait() on the underlying mpb.Progress, if present. Otherwise,
// this is a no-op.
func (pw *ProgressLogger) Wait() {
	if progress, ok := pw.output.(*mpb.Progress); ok {
		progress.Wait()
	}
}

func (pw *ProgressLogger) Infoln(s any) {
	if logLevel.shouldPrint(LogLevelInfo) {
		loglnTo(pw.output, LogLevelInfo, s)
	}
}

func (pw *ProgressLogger) Infof(s string, args ...any) {
	if logLevel.shouldPrint(LogLevelInfo) {
		logfTo(pw.output, LogLevelInfo, s, args...)
	}
}

func (pw *ProgressLogger) Debugln(s any) {
	if logLevel.shouldPrint(LogLevelDebug) {
		loglnTo(pw.output, LogLevelDebug, s)
	}
}

func (pw *ProgressLogger) Debugf(s string, args ...any) {
	if logLevel.shouldPrint(LogLevelDebug) {
		logfTo(pw.output, LogLevelDebug, s, args...)
	}
}

func (pw *ProgressLogger) Logln(level LogLevel, s any) {
	if logLevel.shouldPrint(level) {
		loglnTo(pw.output, level, s)
	}
}

func (pw *ProgressLogger) Logf(level LogLevel, s string, args ...any) {
	if logLevel.shouldPrint(level) {
		logfTo(pw.output, level, s, args...)
	}
}
