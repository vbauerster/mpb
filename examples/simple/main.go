package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		startBlock := make(chan time.Time)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// display our name with one space on the right
				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
				// decor.DSyncWidth bit enables same column width synchronization
				decor.Percentage(decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with default eta age, and width reservation of 3
					decor.ETA(decor.ET_STYLE_GO, 0, startBlock, decor.WC{W: 3}), "done",
				),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				startBlock <- time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}
	// wait for all bars to complete and flush
	p.Wait()
}
