package main

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const size = 16
const ewmaAge = 30

func main() {
	var total int64 = size * 1024 * 1024

	r, w := io.Pipe()

	go func() {
		for range 1024 {
			_, err := io.CopyN(w, rand.Reader, size*1024)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second / 11)
		}
		_ = w.Close()
	}()

	p := mpb.New(mpb.WithWidth(60))

	bar := p.New(total,
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, ewmaAge),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", ewmaAge),
		),
	)

	// create proxy reader
	proxyReader, err := bar.ProxyReader(r)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = proxyReader.Close()
	}()

	// copy from proxyReader, ignoring errors
	_, err = io.Copy(io.Discard, proxyReader)
	if err != nil {
		panic(err)
	}

	p.Wait()
}
