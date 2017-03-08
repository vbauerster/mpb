package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vbauerster/mpb"
)

func main() {
	url := "https://homebrew.bintray.com/bottles/libtiff-4.0.7.sierra.bottle.tar.gz"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server return non-200 status: %s\n", resp.Status)
		return
	}

	size := resp.ContentLength

	// create dest
	destName := filepath.Base(url)
	dest, err := os.Create(destName)
	if err != nil {
		fmt.Printf("Can't create %s: %v\n", destName, err)
		return
	}
	defer dest.Close()

	p := mpb.New().SetWidth(64)

	bar := p.AddBar(size).
		PrependCounters("%3s / %3s", mpb.UnitBytes, 18, 0).
		AppendETA(3, 0)

	// create proxy reader
	reader := bar.ProxyReader(resp.Body)

	// and copy from reader, ignoring errors
	io.Copy(dest, reader)

	p.Stop() // if you omit this line, rendering bars goroutine will quit early
	fmt.Println("Finished")
}
