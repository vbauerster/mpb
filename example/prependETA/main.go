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

	name1 := "Bar#1:"
	bar1 := p.AddBar(50).
		PrependName(name1, 0, mpb.DwidthSync|mpb.DidentRight).
		PrependETA(4, mpb.DwidthSync|mpb.DextraSpace).
		AppendPercentage(5, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 50; i++ {
			sleep(blockSize)
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(100).
		PrependName("", 0, mpb.DwidthSync).
		PrependETA(4, mpb.DwidthSync|mpb.DextraSpace).
		AppendPercentage(5, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			sleep(blockSize)
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).
		PrependName("Bar#3:", 0, mpb.DwidthSync|mpb.DidentRight).
		PrependETA(4, mpb.DwidthSync|mpb.DextraSpace).
		AppendPercentage(5, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			sleep(blockSize)
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	wg.Wait()
	p.Stop()
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
