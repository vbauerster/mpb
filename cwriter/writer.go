package cwriter

import (
	"bytes"
	"io"
	"os"
	"strconv"
)

// https://github.com/dylanaraps/pure-sh-bible#cursor-movement
const (
	escOpen  = "\x1b["
	cuuAndEd = "A\x1b[J"
)

// New returns a new Writer with defaults.
func New(out io.Writer, defaultWidth int, forceTTY bool) *Writer {
	w := &Writer{
		Buffer:   new(bytes.Buffer),
		out:      out,
		forceTTY: forceTTY,
		termSize: func(_ int) (int, int, error) {
			height := defaultWidth*3/2 + 1
			return defaultWidth, height, nil
		},
	}
	if f, ok := out.(*os.File); ok {
		if fd := int(f.Fd()); IsTerminal(fd) {
			w.fd = fd
			w.terminal = true
			w.termSize = GetSize
		}
	}
	bb := make([]byte, 16)
	w.ew = escWriter(bb[:copy(bb, []byte(escOpen))])
	return w
}

// IsTerminal tells whether underlying io.Writer is terminal aka TTY.
func (w *Writer) IsTerminal() bool {
	return w.terminal
}

// GetTermSize returns WxH of underlying terminal.
func (w *Writer) GetTermSize() (width, height int, err error) {
	return w.termSize(w.fd)
}

type escWriter []byte

func (b escWriter) ansiCuuAndEd(out io.Writer, n int) error {
	b = strconv.AppendInt(b, int64(n), 10)
	_, err := out.Write(append(b, []byte(cuuAndEd)...))
	return err
}
