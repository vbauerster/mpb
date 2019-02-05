// +build windows

package cwriter

import (
	"fmt"
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
)

type short int16
type dword uint32
type word uint16

type coord struct {
	x short
	y short
}

type smallRect struct {
	left   short
	top    short
	right  short
	bottom short
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        word
	window            smallRect
	maximumWindowSize coord
}

func (w *Writer) clearLines() {
	if !w.isTerminal {
		fmt.Fprintf(w.out, cuuAndEd, w.lineCount)
	}
	var info consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(w.fd, uintptr(unsafe.Pointer(&info)))

	for i := 0; i < w.lineCount; i++ {
		// move the cursor up
		info.cursorPosition.y--
		procSetConsoleCursorPosition.Call(w.fd, uintptr(*(*int32)(unsafe.Pointer(&info.cursorPosition))))
		// clear the line
		cursor := coord{
			x: info.window.left,
			y: info.window.top + info.cursorPosition.y,
		}
		var count, dw dword
		count = dword(info.size.x)
		procFillConsoleOutputCharacter.Call(w.fd, uintptr(' '), uintptr(count), *(*uintptr)(unsafe.Pointer(&cursor)), uintptr(unsafe.Pointer(&dw)))
	}
}
