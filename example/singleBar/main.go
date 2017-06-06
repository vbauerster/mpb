package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func main() {
	p := mpb.New(
		// Override default (80) width
		mpb.WithWidth(100),
		// Override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// Override default 100ms refresh rate
		mpb.WithRefreshRate(120*time.Millisecond),
	)

	// Add a bar. You're not limited to just one bar, add many if you need.
	bar := p.AddBar(100,
		mpb.PrependDecorators(decor.Name("Single Bar:", 0, 0)),
		mpb.AppendDecorators(decor.Percentage(5, 0)),
	)

	for i := 0; i < 100; i++ {
		bar.Incr(1) // increment progress bar
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	// Don't forget to stop mpb's rendering goroutine
	p.Stop()
	fmt.Println("Stop")
}
