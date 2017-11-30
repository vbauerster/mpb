// +build !windows

package cwriter_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/vbauerster/mpb/cwriter"
)

var (
	cursorUp           = fmt.Sprintf("%c[%dA", cwriter.ESC, 1)
	clearLine          = fmt.Sprintf("%c[2K\r", cwriter.ESC)
	clearCursorAndLine = cursorUp + clearLine
)

// TestWriterPosix by writing and flushing many times. The output buffer
// must contain the clearCursor and clearLine sequences.
func TestWriterPosix(t *testing.T) {
	out := new(bytes.Buffer)
	w := cwriter.New(out)

	for _, tcase := range []struct {
		input, expectedOutput string
	}{
		{input: "foo\n", expectedOutput: "foo\n"},
		{input: "bar\n", expectedOutput: "foo\n" + clearCursorAndLine + "bar\n"},
		{input: "fizz\n", expectedOutput: "foo\n" + clearCursorAndLine + "bar\n" + clearCursorAndLine + "fizz\n"},
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
