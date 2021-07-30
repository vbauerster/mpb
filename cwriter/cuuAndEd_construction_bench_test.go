package cwriter

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"
)

func BenchmarkWithFprintf(b *testing.B) {
	cuuAndEd := "\x1b[%dA\x1b[J"
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, cuuAndEd, 4)
	}
}

func BenchmarkWithJoin(b *testing.B) {
	bCuuAndEd := [][]byte{[]byte("\x1b["), []byte("A\x1b[J")}
	for i := 0; i < b.N; i++ {
		ioutil.Discard.Write(bytes.Join(bCuuAndEd, []byte(strconv.Itoa(4))))
	}
}

func BenchmarkWithAppend(b *testing.B) {
	escOpen := []byte("\x1b[")
	cuuAndEd := []byte("A\x1b[J")
	for i := 0; i < b.N; i++ {
		ioutil.Discard.Write(append(strconv.AppendInt(escOpen, 4, 10), cuuAndEd...))
	}
}

func BenchmarkWithCopy(b *testing.B) {
	w := New(ioutil.Discard)
	w.lines = 4
	for i := 0; i < b.N; i++ {
		w.ansiCuuAndEd()
	}
}
