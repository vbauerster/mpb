package mpb_test

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Example() {
	// Star mpb's rendering goroutine.
	p := mpb.New(
		// override default (80) width
		mpb.WithWidth(100),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 100ms refresh rate
		mpb.WithRefreshRate(120*time.Millisecond),
	)

	total := 100
	barName := "Single Bar:"
	// Add a bar. You're not limited to just one bar, add many if you need.
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(barName, 0, decor.DwidthSync|decor.DidentRight),
			decor.ETA(4, decor.DSyncSpace),
		),
		mpb.AppendDecorators(decor.Percentage(5, 0)),
	)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		bar.Incr(1) // increment progress bar
	}

	// Don't forget to stop mpb's rendering goroutine
	p.Stop()
}

func ExampleBar_InProgress() {
	p := mpb.New()
	bar := p.AddBar(100, mpb.AppendDecorators(decor.Percentage(5, 0)))

	for bar.InProgress() {
		time.Sleep(time.Millisecond * 20)
		bar.Incr(1)
	}
}
