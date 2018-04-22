// +build !windows

package cwriter_test

import (
	"bytes"
	"testing"

	. "github.com/vbauerster/mpb/cwriter"
)

// TestWriterPosix by writing and flushing many times. The output buffer
// must contain the clearCursor and clearLine sequences.
func TestWriterPosix(t *testing.T) {
	out := new(bytes.Buffer)
	w := New(out)

	for _, tcase := range []struct {
		input, expectedOutput string
	}{
		{input: "foo\n", expectedOutput: "foo\n"},
		{input: "bar\n", expectedOutput: "foo\n" + ClearCursorAndLine + "bar\n"},
		{input: "fizz\n", expectedOutput: "foo\n" + ClearCursorAndLine + "bar\n" + ClearCursorAndLine + "fizz\n"},
	} {
		t.Run(tcase.input, func(t *testing.T) {
			w.Write([]byte(tcase.input))
			w.Flush()
			output := out.String()
			if output != tcase.expectedOutput {
				t.Fatalf("want %q, got %q", tcase.expectedOutput, output)
			}
		})
	}
}
