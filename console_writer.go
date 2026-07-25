package mpb

import "io"

// ConsoleWriter interface.
// mpb package interacts with terminal via ConsoleWriter interface.
// Default implementation is "github.com/vbauerster/mpb/v8/cwriter" package.
// Custom implementation can be provided via WithConsoleWriter ContainerOption.
type ConsoleWriter interface {
	io.Writer
	io.ReaderFrom
	IsTerminal() bool
	GetSafeSize() (width, height int)
	GetTermSize() (width, height int, err error)
	Flush(lines int) error
}
