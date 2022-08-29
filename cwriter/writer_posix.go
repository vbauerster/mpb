//go:build !windows

package cwriter

import (
	"golang.org/x/sys/unix"
)

// Flush flushes the underlying buffer.
// It's caller's responsibility to pass correct number of lines.
func (w *Writer) Flush(lines int) error {
	_, err := w.WriteTo(w.out)
	// some terminals interpret 'cursor up 0' as 'cursor up 1'
	if err == nil && lines > 0 {
		err = w.ew.ansiCuuAndEd(w, lines)
	}
	return err
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(ws.Col), int(ws.Row), nil
}

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return err == nil
}
