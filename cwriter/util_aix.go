//go:build aix

package cwriter

import "golang.org/x/sys/unix"

// util_linux.go is filename-constrained to GOOS=linux, so AIX needs its own
// file even though it uses the same TCGETS constant (issue #150).
const ioctlReadTermios = unix.TCGETS
