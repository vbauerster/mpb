package cwriter

import (
	"bytes"
	"io"
)

// ESC is the ASCII code for escape character
const ESC = 27

// Writer is a buffered the writer that updates the terminal.
// The contents of writer will be flushed when Flush is called.
type Writer struct {
	out io.Writer

	buf       bytes.Buffer
	lineCount int
}

// New returns a new Writer with defaults
func New(w io.Writer) *Writer {
	return &Writer{
		out: w,
	}
}

// Flush flushes the underlying buffer
func (w *Writer) Flush() error {
	// do nothing if buffer is empty
	if len(w.buf.Bytes()) == 0 {
		return nil
	}
	w.clearLines()

	lines := 0
	for _, b := range w.buf.Bytes() {
		if b == '\n' {
			lines++
		}
	}
	w.lineCount = lines
	_, err := w.out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

// Write save the contents of b to its buffers. The only errors returned are ones encountered while writing to the underlying buffer.
func (w *Writer) Write(b []byte) (n int, err error) {
	return w.buf.Write(b)
}
