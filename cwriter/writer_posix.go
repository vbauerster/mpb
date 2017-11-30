// +build !windows

package cwriter

import (
	"io"
	"strings"
	"syscall"
	"unsafe"
)

func (w *Writer) clearLines() error {
	_, err := io.WriteString(w.out, strings.Repeat(clearCursorAndLine, w.lineCount))
	return err
}

// TermSize returns the dimensions of the given terminal.
// the code is stolen from "golang.org/x/crypto/ssh/terminal"
func TermSize() (width, height int, err error) {
	var dimensions [4]uint16

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdout), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}
