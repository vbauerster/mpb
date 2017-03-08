// +build !windows

package cwriter

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	cursorUp           = fmt.Sprintf("%c[%dA", ESC, 1)
	clearLine          = fmt.Sprintf("%c[2K\r", ESC)
	clearCursorAndLine = cursorUp + clearLine
)

func (w *Writer) clearLines() {
	for i := 0; i < w.lineCount; i++ {
		fmt.Fprint(w.out, clearCursorAndLine)
	}
}

// GetTermSize returns the dimensions of the given terminal.
// the code is stolen from "golang.org/x/crypto/ssh/terminal"
func GetTermSize() (width, height int, err error) {
	var dimensions [4]uint16

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdout), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}
