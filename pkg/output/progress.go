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
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/term"
	"oras.land/oras-go/v2"
)

func shouldPrintProgress() bool {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	switch progressStyle {
	case "none", "false":
		return false
	default:
		return true
	}
}

func barStyle() mpb.BarStyleComposer {
	switch progressStyle {
	case "plain":
		return mpb.BarStyle().Lbound("|").Filler("=").Tip(">").Padding("-").Rbound("|")
	case "fancy":
		return mpb.BarStyle().Lbound("|").Filler("░").Tip("░").Padding("·").Rbound("|")
	case "cherry":
		return mpb.BarStyle().Lbound("| ").Filler("· ").Tip("(<", "(-").Padding("• ").Rbound("🍒  ")
	default:
		return mpb.BarStyle().Lbound("|").Filler("=").Tip(">").Padding("-").Rbound("|")
	}
}

func GenericProgressBar(name, doneMsg string, total int64) *ProgressBar {
	if !progressEnabled {
		return &ProgressBar{}
	}
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithAutoRefresh(),
	)
	return &ProgressBar{
		bar: p.New(total,
			barStyle(),
			mpb.PrependDecorators(
				decor.OnComplete(decor.Name(name, decor.WC{C: decor.DindentRight | decor.DextraSpace}), doneMsg),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
			),
			mpb.BarFillerClearOnComplete(),
		),
		progress: p,
	}
}

type ProgressBar struct {
	bar      *mpb.Bar
	progress *mpb.Progress
}

func (b *ProgressBar) Increment() {
	if b.bar != nil {
		b.bar.Increment()
	}
}

func (b *ProgressBar) Done() {
	if b.progress != nil {
		b.progress.Wait()
	}
}

// wrappedRepo wraps oras.Target to show a progress bar on Push() operations.
type wrappedRepo struct {
	oras.Target
	progress *mpb.Progress
}

func (w *wrappedRepo) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	shortDigest := expected.Digest.Encoded()[0:8]
	bar := w.progress.New(expected.Size,
		barStyle(),
		mpb.PrependDecorators(
			decor.Name("Copying "+shortDigest),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .1f / % .1f"), fmt.Sprintf("%-9s", FormatBytes(expected.Size))),
			decor.OnComplete(decor.Name(" | "), " | "),
			decor.OnComplete(decor.AverageSpeed(decor.SizeB1024(0), "% .2f"), "done"),
		),
		mpb.BarFillerOnComplete("|"),
	)
	proxyReader, err := bar.ProxyReader(content)
	if err != nil {
		// Should never happen; ProxyReader only returns error if bar is closed or aborted
		return err
	}
	defer proxyReader.Close()

	return w.Target.Push(ctx, expected, proxyReader)
}

// WrapTarget wraps an oras.Target so that calls to Push print a progress bar.
// If output is configured to not print progress bars, this is a no-op.
func WrapTarget(wrap oras.Target) (oras.Target, *ProgressLogger) {
	if !progressEnabled {
		return wrap, &ProgressLogger{stdout}
	}
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(150*time.Millisecond),
	)
	return &wrappedRepo{
		Target:   wrap,
		progress: p,
	}, &ProgressLogger{p}
}

type ProgressTar struct {
	tw  *tar.Writer
	pw  io.WriteCloser
	bar *mpb.Bar
}

func (t *ProgressTar) Write(b []byte) (int, error) {
	if t.pw != nil {
		return t.pw.Write(b)
	}
	return t.tw.Write(b)
}

func (t *ProgressTar) WriteHeader(hdr *tar.Header) error {
	return t.tw.WriteHeader(hdr)
}

func (t *ProgressTar) Close() error {
	// Note: at the time of implementation, the proxied io.WriteCloser returned by ProxyWriter is simply the underlying
	// io.Writer, wrapped in a no-op Close() if it is not already a WriteCloser. Closing both pw and tw represents a double
	// close on the underlying WriteCloser and should be avoided. This is something that could change in future versions
	// of the dependency.
	if t.pw != nil {
		return t.pw.Close()
	} else {
		return t.tw.Close()
	}
}

// TarProgress is used to get an io.Writer and logger that print progress bars when written to (if progress bars
// are enabled). Closing the returned io.WriteCloser will close the input tar.Writer as well; the input tar.Writer
// *should not* be closed independently.
func TarProgress(total int64, tw *tar.Writer) (*ProgressTar, *ProgressLogger) {
	if !progressEnabled || total == 0 {
		return &ProgressTar{tw: tw}, &ProgressLogger{stdout}
	}

	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(150*time.Millisecond),
	)
	bar := p.New(total,
		barStyle(),
		mpb.PrependDecorators(
			decor.Name("Packing"),
		),
		mpb.AppendDecorators(
			decor.Counters(decor.SizeB1024(0), "% .1f / % .1f"),
			decor.Name(" | "),
			decor.AverageSpeed(decor.SizeB1024(0), "% .2f"),
		),
		mpb.BarRemoveOnComplete(),
	)

	// Note: ProxyWriter returns a WriteCloser, but the Close() implementation simply forwards to the underlying writer's
	// Close() implmentation if it exists. This leads to potential accidental double closes so we have to be careful
	pw, err := bar.ProxyWriter(tw)
	if err != nil {
		// Should not happen; ProxyWriter returns error only if the bar is already completed or aborted, but we just created it.
		SafeLogf(LogLevelWarn, "Error creating progress bar: %s", err)
	}
	return &ProgressTar{tw: tw, pw: pw, bar: bar}, &ProgressLogger{p}
}

type PullProgress struct {
	progress *mpb.Progress
	ProgressLogger
}

func (p *PullProgress) ProxyWriter(w io.Writer, digest string, size, offset int64) io.Writer {
	if !progressEnabled || p.progress == nil {
		return w
	}
	shortDigest := digest[0:8]

	bar := p.progress.New(size,
		barStyle(),
		mpb.PrependDecorators(
			decor.Name("Copying "+shortDigest),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Counters(decor.SizeB1024(0), "% .1f / % .1f"), fmt.Sprintf("%-9s", FormatBytes(size))),
			decor.OnComplete(decor.Name(" | "), " | "),
			decor.OnComplete(decor.AverageSpeed(decor.SizeB1024(0), "% .2f"), "done"),
		),
		mpb.BarFillerOnComplete("|"),
	)
	bar.IncrInt64(offset)
	pw, err := bar.ProxyWriter(w)
	if err != nil {
		// Should not happen; ProxyWriter returns error only if the bar is already completed or aborted, but we just created it.
		SafeLogf(LogLevelWarn, "Error creating progress bar: %s", err)
	}
	return pw
}

func (p *PullProgress) Done() {
	if p.progress != nil {
		p.progress.Wait()
	}
}

func NewPullProgress(ctx context.Context) *PullProgress {
	if !progressEnabled {
		return &PullProgress{
			ProgressLogger: ProgressLogger{stdout},
		}
	}
	p := mpb.NewWithContext(ctx,
		mpb.WithWidth(60),
		mpb.WithRefreshRate(150*time.Millisecond),
	)
	return &PullProgress{
		progress:       p,
		ProgressLogger: ProgressLogger{p},
	}
}

type DownloadProgressBar struct {
	progress *mpb.Progress
}

func NewDownloadProgress() (*DownloadProgressBar, *ProgressLogger) {
	if !progressEnabled {
		return &DownloadProgressBar{}, &ProgressLogger{stdout}
	}
	p := mpb.New(
		mpb.WithWidth(30),
		mpb.WithRefreshRate(150*time.Millisecond),
	)
	return &DownloadProgressBar{
		progress: p,
	}, &ProgressLogger{p}
}

func (pb *DownloadProgressBar) TrackDownload(rc io.ReadCloser, name string, totalSize int64) io.ReadCloser {
	if pb.progress == nil {
		return rc
	}
	bar := pb.progress.New(totalSize,
		barStyle(),
		mpb.PrependDecorators(
			decor.Name("Downloading"),
		),
		mpb.AppendDecorators(
			decor.Counters(decor.SizeB1024(0), "% .1f / % .1f"),
			decor.Name(" | "),
			decor.AverageSpeed(decor.SizeB1024(0), "% .2f"),
			decor.Name(" | "),
			decor.Name(name),
		),
		mpb.BarRemoveOnComplete(),
	)
	barRC, err := bar.ProxyReader(rc)
	if err != nil {
		// Should never happen; ProxyReader only returns error if bar is closed or aborted
		SafeLogf(LogLevelWarn, "Error creating progress bar: %s", err)
	}
	return barRC
}

func (pb *DownloadProgressBar) Done() {
	if pb.progress != nil {
		pb.progress.Wait()
	}
}

// WrapReadCloser wraps a io.ReadCloser so that a progress bar is printed when it is read from.
// The returned io.ReadCloser will forward Close() calls to the underlying io.ReadCloser, so only
// Close() should be called _only_ for the returned io.ReadCloser to avoid double-closing.
//
// Once all read operations are complete, the returned *ProgressLogger's Wait() method should be called to
// ensure any progress bars are marked as completed.
func WrapReadCloser(prependText string, size int64, rc io.ReadCloser) (io.ReadCloser, *ProgressLogger) {
	if !progressEnabled {
		return rc, &ProgressLogger{stdout}
	}
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(150*time.Millisecond),
	)
	bar := p.New(size,
		barStyle(),
		mpb.PrependDecorators(
			decor.Name(prependText),
		),
		mpb.AppendDecorators(
			decor.Counters(decor.SizeB1024(0), "% .1f / % .1f"),
			decor.Name(" | "),
			decor.AverageSpeed(decor.SizeB1024(0), "% .2f"),
		),
		mpb.BarRemoveOnComplete(),
	)

	// Note: bar.ProxyReader takes an io.Reader only, suggesting an independent Close() implementation, but it simply
	// forward Close() calls to the underlying ReadCloser via interface assertion. This means closing both the ReadCloser
	// provided *and* the proxied ReadCloser is a double-close. In case the implementation of ProxyReader() changes
	// in the future, *only* the proxied reader should be closed.
	pr, err := bar.ProxyReader(rc)
	if err != nil {
		// Should never happen; ProxyReader only returns error if bar is closed or aborted
		SafeLogf(LogLevelWarn, "Error creating progress bar: %s", err)
	}
	return pr, &ProgressLogger{p}
}
