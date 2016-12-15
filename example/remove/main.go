package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	maxBlockSize = 12
)

func main() {

	p := mpb.New().SetWidth(64)
	// p := mpb.New().RefreshRate(100 * time.Millisecond).SetWidth(64)

	name1 := "Bar#1:"
	bar1 := p.AddBar(50).AppendETA().PrependPercentage(3).PrependName(name1, len(name1))
	p.Wg.Add(1)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 50; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(100).AppendETA().PrependPercentage(3).PrependName("", 0-len(name1))
	p.Wg.Add(1)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).AppendETA().PrependPercentage(3).PrependName("Bar#3:", 0)
	p.Wg.Add(1)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	time.Sleep(3 * time.Second)
	p.RemoveBar(bar2)

	p.WaitAndStop()
	fmt.Println("stop")
	// p.AddBar(1) // panic: send on closed channnel
}
