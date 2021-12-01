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
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.BarID(i),
			mpb.BarOptional(mpb.BarRemoveOnComplete(), i == 0),
			mpb.PrependDecorators(
				decor.Name(name),
			),
			mpb.AppendDecorators(
				decor.Any(func(s decor.Statistics) string {
					return fmt.Sprintf("completed: %v", s.Completed)
				}, decor.WCSyncSpaceR),
				decor.Any(func(s decor.Statistics) string {
					return fmt.Sprintf("aborted: %v", s.Aborted)
				}, decor.WCSyncSpaceR),
				decor.OnComplete(decor.NewPercentage("%d", decor.WCSyncSpace), "done"),
				decor.OnAbort(decor.NewPercentage("%d", decor.WCSyncSpace), "ohno"),
			),
		)
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; !bar.Completed(); i++ {
				if bar.ID() == 2 && i >= 42 {
					bar.Abort(false)
				}
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}

	p.Wait()
}
