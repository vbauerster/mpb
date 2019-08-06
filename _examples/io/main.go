package main

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func main() {
	var total int64 = 1024 * 1024 * 500
	reader := io.LimitReader(rand.Reader, total)

	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	bar := p.AddBar(total, mpb.BarStyle("[=>-|"),
		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, float64(total)/2048),
			decor.Name(" ] "),
			decor.AverageSpeed(decor.UnitKiB, "% .2f"),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(reader)
	defer proxyReader.Close()

	// copy from proxyReader, ignoring errors
	io.Copy(ioutil.Discard, proxyReader)

	p.Wait()
}
