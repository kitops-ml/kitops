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

// formatAndWrite formats a log message and writes it to w.
// Used by defaultLogger and ProgressLogger.
func formatAndWrite(w io.Writer, level LogLevel, format string, args ...any) {
	if !logLevel.shouldPrint(level) {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	str := fmt.Sprintf(format, args...)
	str = strings.ToUpper(str[:1]) + str[1:]
	str = level.getPrefix() + str
	fmt.Fprint(w, str)
}

type defaultLogger struct{}

func (defaultLogger) Log(level LogLevel, format string, args ...any) {
	formatAndWrite(level.getOutput(), level, format, args...)
}

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

// SystemInfoln is like Infoln except it logs to stderr. This should be used
// for system messages such as update notifications
func SystemInfoln(s any) {
	currentLogger.Log(LogLevelSystem, "%s\n", fmt.Sprint(s))
}

// SystemInfof is like Infof except it logs to stderr. This should be used
// for system messages such as update notifications
func SystemInfof(s string, args ...any) {
	currentLogger.Log(LogLevelSystem, s, args...)
}

func Logln(level LogLevel, s any) {
	currentLogger.Log(level, "%s\n", fmt.Sprint(s))
}

func Logf(level LogLevel, s string, args ...any) {
	currentLogger.Log(level, s, args...)
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
//
// Note: ProgressLogger writes directly to its underlying writer (typically an
// mpb.Progress instance) and does not route through a custom Logger set via
// SetLogger. This is necessary to coordinate output with active progress bars.
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
	formatAndWrite(pw.output, LogLevelInfo, "%s\n", fmt.Sprint(s))
}

func (pw *ProgressLogger) Infof(s string, args ...any) {
	formatAndWrite(pw.output, LogLevelInfo, s, args...)
}

func (pw *ProgressLogger) Debugln(s any) {
	formatAndWrite(pw.output, LogLevelDebug, "%s\n", fmt.Sprint(s))
}

func (pw *ProgressLogger) Debugf(s string, args ...any) {
	formatAndWrite(pw.output, LogLevelDebug, s, args...)
}

func (pw *ProgressLogger) Logln(level LogLevel, s any) {
	formatAndWrite(pw.output, level, "%s\n", fmt.Sprint(s))
}

func (pw *ProgressLogger) Logf(level LogLevel, s string, args ...any) {
	formatAndWrite(pw.output, level, s, args...)
}
