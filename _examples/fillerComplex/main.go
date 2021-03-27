package main

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/complexfiller"
	"github.com/vbauerster/mpb/v6/decor"
)

func main() {
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(64))

	total := 100
	name := "Complex Filler:"
	bar := p.Add(int64(total),
		// NewBarComplexFiller allows for multi-rune strings
		complexfiller.NewBarFiller("[\u001b[36;1m", "_", "\u001b[0m⛵\u001b[36;1m", "_", "\u001b[0m]", "\u001b[0m⛵\u001b[36;1m"),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name(name),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}
	// wait for our bar to complete
	p.Wait()
}
