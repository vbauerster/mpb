package mpb

import (
	"bytes"
	"io"
)

// ConsoleWriter interface.
// mpb package interacts with terminal via ConsoleWriter interface.
// Custom implementation can be provided via WithConsoleWriter ContainerOption.
type ConsoleWriter interface {
	io.Writer
	io.ReaderFrom
	IsTerminal() bool
	GetTermSize() (width, height int, err error)
	Flush(lines int) error
}

// NewLogWriter returns a ConsoleWriter that appends every rendered frame to w.
// Unlike the default terminal writer, it does not emit cursor-control sequences
// or overwrite previously rendered lines. Set the desired rendering width with
// WithWidth when constructing the Progress container.
func NewLogWriter(w io.Writer) ConsoleWriter {
	return &logWriter{dst: w}
}

type logWriter struct {
	dst io.Writer
	buf bytes.Buffer
}

func (w *logWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *logWriter) ReadFrom(r io.Reader) (int64, error) {
	return w.buf.ReadFrom(r)
}

func (*logWriter) IsTerminal() bool {
	return false
}

func (*logWriter) GetTermSize() (int, int, error) {
	return 0, 0, nil
}

func (w *logWriter) Flush(_ int) error {
	_, err := w.buf.WriteTo(w.dst)
	return err
}
