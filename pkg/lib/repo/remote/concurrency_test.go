package remote

import (
	"os"
	"runtime"
	"testing"
)

func TestDefaultUploadConcurrency(t *testing.T) {
	os.Unsetenv("KITOPS_UPLOAD_CONCURRENCY")
	got := defaultUploadConcurrency()
	want := int64(runtime.NumCPU() * 4)
	if want < 4 {
		want = 4
	}
	if want > 64 {
		want = 64
	}
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestDefaultUploadConcurrencyOverride(t *testing.T) {
	os.Setenv("KITOPS_UPLOAD_CONCURRENCY", "42")
	defer os.Unsetenv("KITOPS_UPLOAD_CONCURRENCY")
	got := defaultUploadConcurrency()
	if got != 42 {
		t.Fatalf("expected override 42, got %d", got)
	}
}
