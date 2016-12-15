package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

func main() {
	// No need to initialize sync.WaitGroup, as it is initialized implicitly
	p := mpb.New() // Star mpb container
	for i := 0; i < 3; i++ {
		p.Wg.Add(1) // add wg counter
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name)).AppendPercentage()
		go func() {
			// you can p.AddBar() here, but ordering will be non deterministic
			for i := 0; i < 100; i++ {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}
	p.WaitAndStop() // Wait for goroutines to finish
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("finish")
}
