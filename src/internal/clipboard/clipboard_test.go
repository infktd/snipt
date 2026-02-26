package clipboard

import (
	"runtime"
	"testing"
)

func TestAvailable(t *testing.T) {
	// On macOS, pbcopy should always be available
	if runtime.GOOS == "darwin" {
		if !Available() {
			t.Error("expected clipboard to be available on macOS")
		}
	}
}

func TestRoundTrip(t *testing.T) {
	if !Available() {
		t.Skip("no clipboard tool available")
	}

	text := "snipt-test-clipboard-roundtrip"
	if err := Write(text); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got != text {
		t.Errorf("clipboard round-trip failed: got %q, want %q", got, text)
	}
}
