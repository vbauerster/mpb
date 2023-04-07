package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// passed wg will be accounted at p.Wait() call
	p := mpb.NewWithContext(ctx, mpb.WithWaitGroup(&wg))
	total := 300
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name, decor.WCSyncWidthR),
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				// note that OnComplete will not be fired, because of cancel
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
			),
		)

		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for bar.IsRunning() {
				// start variable is solely for EWMA calculation
				// EWMA's unit of measure is an iteration's duration
				start := time.Now()
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				// we need to call EwmaIncrement to fulfill ewma decorator's contract
				bar.EwmaIncrement(time.Since(start))
			}
		}()
	}
	// wait for passed wg and for all bars to complete and flush
	p.Wait()
}
