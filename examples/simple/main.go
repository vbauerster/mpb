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
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// Display our static name with one space on the right
				decor.StaticName(name, len(name)+1, decor.DidentRight),
				// DwidthSync bit enables same column width synchronization
				decor.Percentage(0, decor.DwidthSync),
			),
			mpb.AppendDecorators(
				// replace our ETA decorator with "done!", on bar completion event
				decor.OnComplete(decor.ETA(3, 0), "done!", 0, 0),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}
	// first wait for provided wg, then
	// wait for all bars to complete and flush
	p.Wait()
}
