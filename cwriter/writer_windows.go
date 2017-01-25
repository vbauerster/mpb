// +build windows

package cwriter

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"github.com/mattn/go-isatty"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
)

type (
	short int16
	word  uint16
	dword uint32

	coord struct {
		x short
		y short
	}
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
)

// FdWriter is a writer with a file descriptor.
type FdWriter interface {
	io.Writer
	Fd() uintptr
}

func (w *Writer) clearLines() {
	f, ok := w.out.(FdWriter)
	if ok && !isatty.IsTerminal(f.Fd()) {
		for i := 0; i < w.lineCount; i++ {
			fmt.Fprintf(w.out, "%c[%dA", ESC, 1) // move the cursor up
			fmt.Fprintf(w.out, "%c[2K\r", ESC)   // clear the line
		}
		return
	}
	fd := f.Fd()
	var info consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(fd, uintptr(unsafe.Pointer(&info)))

	for i := 0; i < w.lineCount; i++ {
		// move the cursor up
		info.cursorPosition.y--
		procSetConsoleCursorPosition.Call(fd, uintptr(*(*int32)(unsafe.Pointer(&info.cursorPosition))))
		// clear the line
		cursor := coord{
			x: info.window.left,
			y: info.window.top + info.cursorPosition.y,
		}
		var count, w dword
		count = dword(info.size.x)
		procFillConsoleOutputCharacter.Call(fd, uintptr(' '), uintptr(count), *(*uintptr)(unsafe.Pointer(&cursor)), uintptr(unsafe.Pointer(&w)))
	}
}

// GetTermSize returns the dimensions of the given terminal.
// the code is stolen from "golang.org/x/crypto/ssh/terminal"
func GetTermSize() (width, height int, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return 0, 0, error(e)
	}
	return int(info.size.x), int(info.size.y), nil
}
