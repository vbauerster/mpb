package mpb_test

import (
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Example() {
	p := mpb.New(
		// override default (80) width
		mpb.WithWidth(64),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	total := 100
	name := "Single Bar:"
	// adding a single bar
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 60, and width reservation of 4
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// ewma based decorators require work duration measurement
		bar.IncrBy(1, time.Since(start))
	}
	// wait for our bar to complete and flush
	p.Wait()
}

func ExampleBar_Completed() {
	p := mpb.New()
	bar := p.AddBar(100)

	max := 100 * time.Millisecond
	for !bar.Completed() {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}

	p.Wait()
}

func ExampleBar_ProxyReader() {
	p := mpb.New()
	// make http get request, ignoring errors
	resp, _ := http.Get("https://homebrew.bintray.com/bottles/libtiff-4.0.7.sierra.bottle.tar.gz")
	defer resp.Body.Close()

	// Assuming ContentLength > 0
	bar := p.AddBar(resp.ContentLength,
		mpb.AppendDecorators(
			decor.CountersKibiByte("%6.1f / %6.1f"),
		),
	)

	// create proxy reader
	reader := bar.ProxyReader(resp.Body)

	// and copy from reader, ignoring errors
	io.Copy(ioutil.Discard, reader)

	p.Wait()
}
