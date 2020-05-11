package main

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func main() {
	var wg sync.WaitGroup
	// pass &wg (optional), so p will wait for it eventually
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(60))
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		var pdecorators mpb.BarOption
		if i == 0 {
			pdecorators = mpb.PrependDecorators(
				decor.Merge(
					decor.OnComplete(
						newVariadicSpinner(decor.WCSyncSpace),
						"done",
					),
					decor.WCSyncSpace, // Placeholder
					decor.WCSyncSpace, // Placeholder
				),
			)
		} else {
			pdecorators = mpb.PrependDecorators(
				decor.CountersNoUnit("% .1d / % .1d", decor.WCSyncSpace),
				decor.OnComplete(decor.Spinner(nil, decor.WCSyncSpace), "done"),
				decor.OnComplete(decor.Spinner(nil, decor.WCSyncSpace), "done"),
			)
		}
		bar := p.AddBar(int64(total),
			pdecorators,
			mpb.AppendDecorators(
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), "done"),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				// start variable is solely for EWMA calculation
				// EWMA's unit of measure is an iteration's duration
				start := time.Now()
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				bar.Increment()
				// we need to call DecoratorEwmaUpdate to fulfill ewma decorator's contract
				bar.DecoratorEwmaUpdate(time.Since(start))
			}
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()
}

func newVariadicSpinner(wc decor.WC) decor.Decorator {
	spinner := decor.Spinner(nil)
	fn := func(s decor.Statistics) string {
		return strings.Repeat(spinner.Decor(s), int(s.Current/3))
	}
	return decor.Any(fn, wc)
}
