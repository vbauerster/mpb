package mpb_test

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

func Example() {
	// Star mpb's rendering goroutine.
	// If you don't plan to cancel, feed with nil
	// otherwise provide context.Context, see cancel example
	p := mpb.New(nil)
	// Set custom format, the default one is "[=>-]"
	p.Format("╢▌▌░╟")

	bar := p.AddBar(100).PrependName("Single Bar:", 0).AppendPercentage()

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		bar.Incr(1)
	}

	// Don't forget to stop mpb's rendering goroutine
	p.Stop()
}
