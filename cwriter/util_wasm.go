//+build wasm js

package cwriter

// There is no ioctl on wasm, so we just use a dummy value.
const ioctlReadTermios = 0
