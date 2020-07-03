// +build windows

package cwriter

import (
	"unsafe"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
)

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")

var (
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
)

func (w *Writer) clearLines() error {
	if !w.isTerminal && isatty.IsCygwinTerminal(w.fd) {
		return w.ansiCuuAndEd()
	}

	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(w.fd), &info); err != nil {
		return err
	}

	info.CursorPosition.Y -= int16(w.lineCount)
	if info.CursorPosition.Y < 0 {
		info.CursorPosition.Y = 0
	}
	_, _, _ = procSetConsoleCursorPosition.Call(
		w.fd,
		uintptr(uint32(uint16(info.CursorPosition.Y))<<16|uint32(uint16(info.CursorPosition.X))),
	)

	// clear the lines
	cursor := &windows.Coord{
		X: info.Window.Left,
		Y: info.CursorPosition.Y,
	}
	count := uint32(info.Size.X) * uint32(w.lineCount)
	_, _, _ = procFillConsoleOutputCharacter.Call(
		w.fd,
		uintptr(' '),
		uintptr(count),
		*(*uintptr)(unsafe.Pointer(cursor)),
		uintptr(unsafe.Pointer(new(uint32))),
	)
	return nil
}

// GetSize returns the visible dimensions of the given terminal.
//
// These dimensions don't include any scrollback buffer height.
func GetSize(fd uintptr) (width, height int, err error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info); err != nil {
		return 0, 0, err
	}
	// terminal.GetSize from crypto/ssh returns following line with both "+ 1",
	// but looks like this causing issue #66. Removing both "+ 1" fixed the issue.
	// return int(info.Window.Right - info.Window.Left + 1), int(info.Window.Bottom - info.Window.Top + 1), nil
	return int(info.Window.Right - info.Window.Left), int(info.Window.Bottom - info.Window.Top), nil
}
