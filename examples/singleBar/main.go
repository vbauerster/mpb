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
		mpb.WithWidth(100),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	total := 100
	name := "Single Bar:"
	// Add a bar
	// You're not limited to just a single bar, add as many as you need
	bar := p.AddBar(int64(total),
		// Prepending decorators
		mpb.PrependDecorators(
			// StaticName decorator with one extra space on right
			decor.StaticName(name, len(name)+1, decor.DidentRight),
			// ETA decorator with width reservation of 3 runes
			decor.ETA(3, 0),
		),
		// Appending decorators
		mpb.AppendDecorators(
			// Percentage decorator with width reservation of 5 runes
			decor.Percentage(5, 0),
		),
	)

	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}
	// Wait for all bars to complete
	p.Wait()
}
