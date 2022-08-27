package cwriter

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
)

// https://github.com/dylanaraps/pure-sh-bible#cursor-movement
const (
	escOpen  = "\x1b["
	cuuAndEd = "A\x1b[J"
)

// used by ansiCuuAndEd func
var escBuf = make([]byte, 8)

// ErrNotTTY not a TeleTYpewriter error.
var ErrNotTTY = errors.New("not a terminal")

// Writer is a buffered the writer that updates the terminal. The
// contents of writer will be flushed when Flush is called.
type Writer struct {
	*bytes.Buffer
	out      io.Writer
	fd       int
	terminal bool
	termSize func(int) (int, int, error)
}

// New returns a new Writer with defaults.
func New(out io.Writer) *Writer {
	w := &Writer{
		Buffer: new(bytes.Buffer),
		out:    out,
		termSize: func(_ int) (int, int, error) {
			return -1, -1, ErrNotTTY
		},
	}
	if f, ok := out.(*os.File); ok {
		w.fd = int(f.Fd())
		if IsTerminal(w.fd) {
			w.terminal = true
			w.termSize = func(fd int) (int, int, error) {
				return GetSize(fd)
			}
		}
	}
	return w
}

// GetTermSize returns WxH of underlying terminal.
func (w *Writer) GetTermSize() (width, height int, err error) {
	return w.termSize(w.fd)
}

// if n > 99 it will allocate because of len(escBuf) = 8
func ansiCuuAndEd(out io.Writer, n int) error {
	escBuf = strconv.AppendInt(escBuf[:copy(escBuf, escOpen)], int64(n), 10)
	_, err := out.Write(append(escBuf, cuuAndEd...))
	return err
}
