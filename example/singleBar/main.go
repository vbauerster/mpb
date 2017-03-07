package main

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

func main() {
	// Star mpb's rendering goroutine.
	// If you don't plan to cancel, feed with nil
	// otherwise provide context.Context, see cancel example
	p := mpb.New()
	// Set custom width for every bar, which mpb will contain
	// The default one in 70
	p.SetWidth(80)
	// Set custom format for every bar, the default one is "[=>-]"
	p.Format("╢▌▌░╟")
	// Set custom refresh rate, the default one is 100 ms
	p.RefreshRate(120 * time.Millisecond)

	// Add a bar. You're not limited to just one bar, add many if you need.
	bar := p.AddBar(100).PrependName("Single Bar:", 0).AppendPercentage()

	for i := 0; i < 100; i++ {
		bar.Incr(1) // increment progress bar
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	// Don't forget to stop mpb's rendering goroutine
	p.Stop()

	// You cannot add bars after p.Stop() has been called
	// p.AddBar(100) // will panic
}
