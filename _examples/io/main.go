package main

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	var total int64 = 64 * 1024 * 1024

	r, w := io.Pipe()

	go func() {
		for i := 0; i < 1024; i++ {
			_, _ = io.Copy(w, io.LimitReader(rand.Reader, 64*1024))
			time.Sleep(time.Second / 10)
		}
		w.Close()
	}()

	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	bar := p.New(total,
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 30),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(r)
	defer proxyReader.Close()

	// copy from proxyReader, ignoring errors
	_, _ = io.Copy(ioutil.Discard, proxyReader)

	p.Wait()
}
