package local

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"github.com/kitops-ml/kitops/pkg/output"
)

type slowReadSeekCloser struct {
	*bytes.Reader
	delay time.Duration
}

func (s *slowReadSeekCloser) Read(p []byte) (int, error) {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.Reader.Read(p)
}
func (s *slowReadSeekCloser) Close() error { return nil }

// stubTarget implements oras.ReadOnlyTarget for testing getNetworkAdjustedConfig
// It returns a reader that simulates network speed based on delay per read.
type stubTarget struct {
	data  []byte
	delay time.Duration
}

var _ oras.ReadOnlyTarget = (*stubTarget)(nil)

func (s *stubTarget) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return &slowReadSeekCloser{Reader: bytes.NewReader(s.data), delay: s.delay}, nil
}

func (s *stubTarget) Exists(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	return true, nil
}
func (s *stubTarget) Resolve(ctx context.Context, reference string) (ocispec.Descriptor, error) {
	return ocispec.Descriptor{}, nil
}

func TestGetNetworkAdjustedConfigScales(t *testing.T) {
	ctx := context.Background()
	progress := output.NewPullProgress(ctx)
	initial := downloadConfig{
		chunkConcurrency:    4,
		layerConcurrency:    2,
		chunkSize:           8 * 1024 * 1024,
		largeLayerThreshold: 1,
	}
	desc := ocispec.Descriptor{Size: 50 * 1024 * 1024}
	fastTarget := &stubTarget{data: make([]byte, 10*1024*1024), delay: 0}
	l := &localRepo{}
	fast := l.getNetworkAdjustedConfig(ctx, fastTarget, initial, desc, progress)
	if fast.chunkConcurrency <= initial.chunkConcurrency {
		t.Fatalf("expected increased concurrency on fast network")
	}

	slowTarget := &stubTarget{data: make([]byte, 10*1024*1024), delay: 3 * time.Millisecond}
	slow := l.getNetworkAdjustedConfig(ctx, slowTarget, initial, desc, progress)
	if slow.chunkConcurrency >= initial.chunkConcurrency {
		t.Fatalf("expected decreased concurrency on slow network")
	}
}
