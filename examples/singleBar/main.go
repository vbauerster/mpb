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
	// adding a single bar
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(
			// Display our static name with one space on the right
			decor.StaticName(name, len(name)+1, decor.DidentRight),
			// ETA decorator with width reservation of 3 runes
			decor.ETA(3, 0),
		),
		mpb.AppendDecorators(
			// Percentage decorator with width reservation of 5 runes
			decor.Percentage(5, 0),
		),
	)

	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// Increment by 1 (there is bar.IncrBy(int) method, if needed)
		bar.Increment()
	}
	// wait for our bar to complete and flush
	p.Wait()
}
