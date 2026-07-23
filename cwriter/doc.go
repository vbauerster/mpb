// Package cwriter is a console writer abstraction for the underlying OS.
//
// # Constructing a Writer
//
// New takes the destination writer, a terminal width hint, and forceTTY:
//
//	w := cwriter.New(os.Stdout, 80, false)
//
// Width is used when the destination is not a TTY (or size cannot be probed).
// Pass 0 to fall back to the package default width (80). forceTTY forces TTY
// behavior even when the destination is not detected as a terminal.
//
// To preserve the previous single-argument construction, probe the terminal
// size first when available:
//
//	width := 0
//	if f, ok := out.(*os.File); ok {
//		if w, _, err := cwriter.GetSize(int(f.Fd())); err == nil {
//			width = w
//		}
//	}
//	cw := cwriter.New(out, width, false)
package cwriter
