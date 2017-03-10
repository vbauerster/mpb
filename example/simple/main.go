package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gopkg.in/vbauerster/mpb.v1"
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New(nil)
	for i := 0; i < 3; i++ {
		wg.Add(1) // add wg delta
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name)).AppendPercentage()
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
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("finish")
}
