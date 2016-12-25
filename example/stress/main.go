package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	totalBars    = 32
	maxBlockSize = 12
)

func main() {

	p := mpb.New(nil)
	var wg sync.WaitGroup
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(120) + 10
		bar := p.AddBar(int64(total)).
			PrependName(name, len(name)).PrependETA(4).
			AppendPercentage().TrimRightSpace()

		go func() {
			defer wg.Done()
			blockSize := rand.Intn(maxBlockSize) + 1
			for i := 0; i < total; i++ {
				sleep(blockSize)
				bar.Incr(1)
				blockSize = rand.Intn(maxBlockSize) + 1
			}
		}()
	}

	wg.Wait()
	p.Stop()
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
