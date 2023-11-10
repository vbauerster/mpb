// +build wasm js

package cwriter

// Don't overwrite clearLines()

// GetSize do nothing
func GetSize(fd int) (width, height int, err error) {
	return 0, 0, nil
}

// IsTerminal do nothing
func IsTerminal(fd int) bool {
	return true
}
