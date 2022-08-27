package cwriter

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
)

// ErrNotTTY not a TeleTYpewriter error.
var ErrNotTTY = errors.New("not a terminal")

// https://github.com/dylanaraps/pure-sh-bible#cursor-movement
const (
	escOpen  = "\x1b["
	cuuAndEd = "A\x1b[J"
)

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

func ansiCuuAndEd(out io.Writer, n int) error {
	buf := make([]byte, 8)
	buf = strconv.AppendInt(buf[:copy(buf, escOpen)], int64(n), 10)
	_, err := out.Write(append(buf, cuuAndEd...))
	return err
}
