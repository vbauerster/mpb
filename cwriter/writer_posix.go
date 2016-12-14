// +build !windows

package cwriter

import (
	"fmt"
)

func (w *Writer) clearLines() {
	for i := 0; i < w.lineCount; i++ {
		fmt.Fprintf(w.out, "%c[%dA", ESC, 0) // move the cursor up
		fmt.Fprintf(w.out, "%c[2K\r", ESC)   // clear the line
	}
}
