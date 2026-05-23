//go:build js || wasm

package cwriter

// We can not use writer in wasm, so we just use a dummy value.
type Writer struct {
}

func (w *Writer) Flush(lines int) error {
	return nil
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	return 0, 0, nil
}

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	return true
}
