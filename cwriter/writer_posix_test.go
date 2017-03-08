// +build !windows

package cwriter_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/vbauerster/mpb/cwriter"
)

var clearSequence = fmt.Sprintf("%c[%dA%c[2K\r", 27, 1, 27)

// TestWriterPosix by writing and flushing many times. The output buffer
// must contain the clearCursor and clearLine sequences.
func TestWriterPosix(t *testing.T) {
	out := new(bytes.Buffer)
	w := cwriter.New(out)

	testCases := []struct {
		input, expectedOutput string
	}{
		{input: "foo\n", expectedOutput: "foo\n"},
		{input: "bar\n", expectedOutput: "foo\n" + clearSequence + "bar\n"},
		{input: "fizz\n", expectedOutput: "foo\n" + clearSequence + "bar\n" + clearSequence + "fizz\n"},
	}
	for _, testCase := range testCases {
		w.Write([]byte(testCase.input))
		w.Flush()
		output := out.String()
		if output != testCase.expectedOutput {
			t.Fatalf("want %q, got %q", testCase.expectedOutput, output)
		}
	}
}
