package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithWidth(14),
	)
	total, numBars := 101, 3
	wg.Add(numBars)

	spinnerStyle := []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"}

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		var bar *mpb.Bar
		if i == 0 {
			bar = p.Add(int64(total),
				mpb.NewBarFiller(mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟")),
				mpb.PrependDecorators(
					// simple name decorator
					decor.Name(name),
				),
				mpb.AppendDecorators(
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60
						decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
					),
				),
			)
		} else {
			bar = p.Add(int64(total),
				mpb.NewBarFiller(mpb.SpinnerStyle(spinnerStyle...)),
				mpb.PrependDecorators(
					// simple name decorator
					decor.Name(name),
				),
				mpb.AppendDecorators(
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60
						decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
					),
				),
			)
		}

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
	// wait for all bars to complete and flush
	p.Wait()
}
