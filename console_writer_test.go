package mpb

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestLogWriterFlush(t *testing.T) {
	var dst bytes.Buffer
	cw := NewLogWriter(&dst)

	if cw.IsTerminal() {
		t.Fatal("log writer must not report itself as a terminal")
	}

	if _, err := io.WriteString(cw, "first\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := cw.ReadFrom(strings.NewReader("second\n")); err != nil {
		t.Fatal(err)
	}
	if err := cw.Flush(2); err != nil {
		t.Fatal(err)
	}

	if got, want := dst.String(), "first\nsecond\n"; got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}

	if err := cw.Flush(0); err != nil {
		t.Fatal(err)
	}
	if got, want := dst.String(), "first\nsecond\n"; got != want {
		t.Fatalf("empty flush duplicated output: got %q, want %q", got, want)
	}
}
