package cwriter

import (
	"bytes"
	"cmp"
	"io"
	"os"
	"strconv"
)

const (
	defaultWidth = 80

	// https://github.com/dylanaraps/pure-sh-bible#cursor-movement
	escOpen  = "\x1b["
	cuuAndEd = "A\x1b[J"
)

// New constructs a new Writer which abstracts writing multi
// lines to a fixed position within a terminal. If terminal
// probe fails provided width will be used as a sane default.
func New(out io.Writer, width int, forceTTY bool) *Writer {
	w := &Writer{
		Buffer:   new(bytes.Buffer),
		out:      out,
		width:    cmp.Or(width, defaultWidth),
		forceTTY: forceTTY,
	}
	if f, ok := out.(*os.File); ok {
		if fd := int(f.Fd()); IsTerminal(fd) {
			w.fd = fd
			w.terminal = true
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

// GetSafeSize returns sane default width and height.
// Should be called only if (*Writer).IsTerminal returns false.
func (w *Writer) GetSafeSize() (width, height int) {
	return w.width, w.width*3/2 + 1
}

// GetTermSize returns width and height of underlying terminal.
// Should be called only if (*Writer).IsTerminal returns true.
func (w *Writer) GetTermSize() (width, height int, err error) {
	return GetSize(w.fd)
}

type escWriter []byte

func (b escWriter) ansiCuuAndEd(out io.Writer, n int) error {
	// some terminals interpret 'cursor up 0' as 'cursor up 1'
	// therefore ignore n <= 0 case
	if n <= 0 {
		return nil
	}
	b = strconv.AppendInt(b, int64(n), 10)
	_, err := out.Write(append(b, []byte(cuuAndEd)...))
	return err
}
