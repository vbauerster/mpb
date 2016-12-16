package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	maxBlockSize = 12
)

func main() {

	var wg sync.WaitGroup
	p := mpb.New().SetWidth(64)
	// p := mpb.New().RefreshRate(100 * time.Millisecond).SetWidth(64)

	name1 := "Bar#1:"
	bar1 := p.AddBar(50).AppendETA().PrependPercentage(3).PrependName(name1, len(name1))
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 50; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(100).AppendETA().PrependPercentage(3).PrependName("", 0-len(name1))
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).AppendETA().PrependPercentage(3).PrependName("Bar#3:", 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	time.Sleep(3 * time.Second)
	// After removing the bar, it is good practice to ask underlying goroutine
	// (2nd one in our example) to stop, so its wg.Done() will execute in time
	p.RemoveBar(bar2)

	wg.Wait()
	p.Stop()
	fmt.Println("stop")
	// p.AddBar(1) // panic: send on closed channnel
}
