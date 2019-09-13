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
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		b := p.AddBar(int64(total), mpb.BarID(i),
			mpb.BarOptOnCond(mpb.BarRemoveOnComplete(), func() bool { return i == 0 }),
			mpb.PrependDecorators(
				decor.Name(name),
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(decor.Percentage()),
		)
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; !b.Completed(); i++ {
				start := time.Now()
				if b.ID() == 2 && i >= 42 {
					// aborting and removing while bar is running
					b.Abort(true)
				}
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				// since ewma decorator is used, we need to pass time.Since(start)
				b.Increment(time.Since(start))
			}
		}()
	}

	p.Wait()
}
