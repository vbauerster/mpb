package mpb_test

import (
	crand "crypto/rand"
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func Example() {
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(64))

	total := 100
	name := "Single Bar:"
	// create a single bar, which will inherit container's width
	bar := p.New(int64(total),
		// BarFillerBuilder with custom style
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{C: decor.DindentRight | decor.DextraSpace}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done"),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
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
	// import crand "crypto/rand"

	var total int64 = 1024 * 1024 * 500
	reader := io.LimitReader(crand.Reader, total)

	p := mpb.New()
	bar := p.AddBar(total,
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(reader)
	defer proxyReader.Close()

	// and copy from reader, ignoring errors
	_, _ = io.Copy(io.Discard, proxyReader)

	p.Wait()
}
