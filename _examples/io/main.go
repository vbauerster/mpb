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
	var total int64 = 1024 * 1024 * 500
	reader := io.LimitReader(rand.Reader, total)

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
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(reader)
	defer proxyReader.Close()

	// copy from proxyReader, ignoring errors
	_, _ = io.Copy(ioutil.Discard, proxyReader)

	p.Wait()
}
