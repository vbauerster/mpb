package main

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func main() {
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
	startBlock := make(chan time.Time)
	// adding a single bar
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(
			// Display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// Replace ETA decorator with message, OnComplete event
			decor.OnComplete(
				// ETA decorator with default eta age, and width reservation of 4
				decor.ETA(decor.ET_STYLE_GO, 0, startBlock, decor.WC{W: 4}),
				"done",
			),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		// update start block time, required for ETA calculation
		startBlock <- time.Now()
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// Increment by 1 (there is bar.IncrBy(int) method, if needed)
		bar.Increment()
	}
	// wait for our bar to complete and flush
	p.Wait()
}
