//go:build windows

package cwriter

import (
	"bytes"
	"io"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                       = windows.NewLazySystemDLL("kernel32.dll")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterA")
)

// Writer is a buffered terminal writer, which moves cursor N lines up
// on each flush except the first one, where N is a number of lines of
// a previous flush.
type Writer struct {
	*bytes.Buffer
	out      io.Writer
	ew       escWriter
	fd       int
	lines    int
	terminal bool
	forceTTY bool
	termSize func(int) (int, int, error)
}

// Flush flushes the underlying buffer.
// It's caller's responsibility to pass correct number of lines.
func (w *Writer) Flush(lines int) error {
	if w.terminal {
		err := w.clearLines()
		if err != nil {
			return err
		}
	}

	_, err := w.WriteTo(w.out)
	if err != nil {
		return err
	}

	if !w.terminal && w.forceTTY {
		return w.ew.ansiCuuAndEd(w, lines)
	}

	w.lines = lines
	return nil
}

func (w *Writer) clearLines() error {
	if w.lines <= 0 {
		return nil
	}

	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(w.fd), &info)
	if err != nil {
		return err
	}

	newPosition := info.CursorPosition
	newPosition.Y -= int16(w.lines)
	if newPosition.Y < 0 {
		newPosition.Y = 0
	}

	// clear lines by writing space character n times starting at newPosition
	var written uint32
	n := uint32(info.Size.X) * uint32(w.lines)
	_, _, err = procFillConsoleOutputCharacter.Call(
		uintptr(w.fd),
		uintptr(' '),
		uintptr(n),
		*(*uintptr)(unsafe.Pointer(&newPosition)),
		uintptr(unsafe.Pointer(&written)),
	)
	if written != n {
		return err
	}

	// move cursor to newPosition for the next write
	_, _, _ = procSetConsoleCursorPosition.Call(
		uintptr(w.fd),
		uintptr(uint32(uint16(newPosition.Y))<<16|uint32(uint16(newPosition.X))),
	)

	return nil
}

// GetSize returns the visible dimensions of the given terminal.
// These dimensions don't include any scrollback buffer height.
func GetSize(fd int) (width, height int, err error) {
	var info windows.ConsoleScreenBufferInfo
	err = windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info)
	if err != nil {
		return 0, 0, err
	}
	// terminal.GetSize from crypto/ssh adds "+ 1" to both width and height:
	// https://go.googlesource.com/crypto/+/refs/heads/release-branch.go1.14/ssh/terminal/util_windows.go#75
	// but looks like this is a root cause of issue #66, so removing both "+ 1" have fixed it.
	return int(info.Window.Right - info.Window.Left), int(info.Window.Bottom - info.Window.Top), nil
}

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(fd), &st)
	return err == nil
}
