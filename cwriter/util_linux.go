//go:build linux

package cwriter

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TCGETS
