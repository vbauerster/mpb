package main

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const size = 32

func main() {
	var total int64 = size * 1024 * 1024

	r, w := io.Pipe()

	go func() {
		for range 1024 {
			_, err := io.Copy(w, io.LimitReader(rand.Reader, size*1024))
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second / 10)
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
			decor.EwmaETA(decor.ET_STYLE_GO, 30),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(r)
	defer func() {
		_ = proxyReader.Close()
	}()

	// copy from proxyReader, ignoring errors
	_, err := io.Copy(io.Discard, proxyReader)
	if err != nil {
		panic(err)
	}

	p.Wait()
}
