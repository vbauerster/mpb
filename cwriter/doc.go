// Package cwriter is a console writer abstraction for the underlying OS.
//
// # Constructing a Writer
//
// New takes the destination writer, a terminal width hint, and forceTTY:
//
//	w := cwriter.New(os.Stdout, 0, false)
//
// Width is used when the destination is not a TTY (or size cannot be probed).
// Pass 0 to fall back to the package default width (80). forceTTY forces TTY
// behavior even when the destination is not detected as a terminal.
//
// To preserve the previous single-argument construction, simply pass 0 and
// false — GetSize is called internally only when the terminal probe succeeds:
//
//	cw := cwriter.New(out, 0, false)
package cwriter
