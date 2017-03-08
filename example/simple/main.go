package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New()
	wg.Add(3) // add wg delta
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).
			PrependName(name, len(name), 0).
			// Prepend Percentage decorator and sync width
			PrependPercentage(3, mpb.DwidthSync|mpb.DextraSpace).
			// Append ETA and don't sync width
			AppendETA(2, 0)
		go func() {
			defer wg.Done()
			// you can p.AddBar() here, but ordering will be non deterministic
			// if you still need p.AddBar() here and maintain ordering, use
			// (*mpb.Progress).BeforeRenderFunc(f mpb.BeforeRender)
			for i := 0; i < 100; i++ {
				bar.Incr(1)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}()
	}
	wg.Wait() // Wait for goroutines to finish
	p.Stop()  // Stop mpb's rendering goroutine
}
