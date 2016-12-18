// +build !windows

package cwriter

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var tty *os.File

func init() {
	var err error
	tty, err = os.Open("/dev/tty")
	if err != nil {
		tty = os.Stdin
	}
}

func (w *Writer) clearLines() {
	for i := 0; i < w.lineCount; i++ {
		fmt.Fprintf(w.out, "%c[%dA", ESC, 0) // move the cursor up
		fmt.Fprintf(w.out, "%c[2K\r", ESC)   // clear the line
	}
}

// TerminalWidth returns width of the terminal.
func TerminalWidth() (int, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		tty.Fd(),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if errno != 0 {
		return 0, errno
	}
	return int(w.Col), nil
}

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
