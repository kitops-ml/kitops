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

package output

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

// DataSink is a writer for command-output payloads (e.g. SLSA predicates,
// generated Kitfiles) that must not corrupt the destination on failure. For a
// file target, writes go to a temp file in the same directory; Commit renames
// the temp file to the user-specified path atomically. For stdout, writes go
// directly to os.Stdout and Commit is a no-op.
//
// Lifecycle: open with OpenDataSink, defer Close, Write payload, then call
// Commit on success. Close removes the temp file if Commit was not called, so
// the original target stays untouched on the failure path.
type DataSink interface {
	io.Writer
	Commit() error
	Close() error
}

// sinkMu guards the swap of the package-level stdout writer when a command
// directs structured data to stdout via OpenDataSink("-").
var sinkMu sync.Mutex

// stdoutSink writes data payloads to os.Stdout and reroutes the package info
// writer to stderr while it is open, so subsequent Infof/Infoln output does
// not corrupt the data stream. Close restores the original info writer.
type stdoutSink struct {
	prevStdout io.Writer
	restored   bool
}

func (s *stdoutSink) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (s *stdoutSink) Commit() error               { return nil }
func (s *stdoutSink) Close() error {
	if s.restored {
		return nil
	}
	sinkMu.Lock()
	stdout = s.prevStdout
	sinkMu.Unlock()
	s.restored = true
	return nil
}

// fileSink stages writes in a temp file alongside the target so the rename in
// Close removes the temp file when Commit was not called
type fileSink struct {
	target    string
	tmp       *os.File
	committed bool
}

func (s *fileSink) Write(p []byte) (int, error) { return s.tmp.Write(p) }

func (s *fileSink) Commit() error {
	if s.committed {
		return nil
	}
	tmpName := s.tmp.Name()
	if err := s.tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.target); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	s.committed = true
	return nil
}

func (s *fileSink) Close() error {
	if s.committed {
		return nil
	}
	tmpName := s.tmp.Name()
	_ = s.tmp.Close()
	return os.Remove(tmpName)
}

// OpenDataSink prepares a sink for the given path.
//
//   - "-": writes go to os.Stdout. The package info writer is rerouted to
//     stderr so subsequent Infof/Infoln output does not corrupt the data
//     stream. The original writer is restored on Close.
//   - any other value: a temp file is created in the same directory as path.
//     Writes go to that file; Commit renames it onto path. The user's
//     existing file at path is not touched until Commit succeeds, so a
//     failure between OpenDataSink and Commit leaves the original intact.
//
// Callers must ensure path is non-empty.
func OpenDataSink(path string) (DataSink, error) {
	if path == "-" {
		sinkMu.Lock()
		defer sinkMu.Unlock()
		s := &stdoutSink{prevStdout: stdout}
		stdout = stderr
		return s, nil
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return nil, err
	}
	// CreateTemp opens with 0600; preserve the historical 0644 for the
	// final file once renamed.
	if err := tmp.Chmod(0644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, err
	}
	return &fileSink{target: path, tmp: tmp}, nil
}

// InfoWriter returns the writer currently used for info-level output. Callers
// that bypass the standard logging helpers (for example, attaching a child
// process's stdout) should use this so they share the same routing decisions
// as Infof/Infoln.
func InfoWriter() io.Writer {
	sinkMu.Lock()
	defer sinkMu.Unlock()
	return stdout
}
