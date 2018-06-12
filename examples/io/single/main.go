package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func main() {
	url := "https://github.com/onivim/oni/releases/download/v0.3.4/Oni-0.3.4-osx.dmg"

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

	p := mpb.New(mpb.WithWidth(64))

	startBlock := make(chan time.Time)
	bar := p.AddBar(size,
		mpb.PrependDecorators(
			decor.CountersKibiByte("% 6.1f / % 6.1f", decor.WC{W: 18}),
		),
		mpb.AppendDecorators(
			decor.ETA(decor.ET_STYLE_HHMMSS, 60, startBlock),
			decor.SpeedKibiByte("% 6.1f", decor.WC{W: 14}),
		),
	)

	// create proxy reader
	reader := bar.ProxyReader(resp.Body, startBlock)

	// and copy from reader, ignoring errors
	io.Copy(dest, reader)

	p.Wait()
}
