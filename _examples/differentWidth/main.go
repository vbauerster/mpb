package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		// container's width.
		mpb.WithWidth(60),
	)
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			// set BarWidth 40 for bar 1 and 2
			mpb.BarOptOnCond(mpb.BarWidth(40), func() bool { return i > 0 }),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with ewma age of 60
					decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
				),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				// since ewma decorator is used, we need to pass time.Since(start)
				bar.Increment(time.Since(start))
			}
		}()
	}
	// wait for all bars to complete and flush
	p.Wait()
}
