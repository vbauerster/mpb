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
	p := mpb.New() // Star mpb's rendering goroutine
	for i := 0; i < 3; i++ {
		wg.Add(1) // add wg delta
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name)).AppendPercentage()
		go func() {
			defer wg.Done()
			// you can p.AddBar() here, but ordering will be non deterministic
			for i := 0; i < 100; i++ {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}
	wg.Wait() // Wait for goroutines to finish
	p.Stop()  // Stop mpb's rendering goroutine
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("finish")
}
