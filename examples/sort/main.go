package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/curser100500/mpb"
	"github.com/curser100500/mpb/decor"
)

const (
	maxBlockSize = 12
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		var name string
		if i != 1 {
			name = fmt.Sprintf("Bar#%d:", i)
		}
		b := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.StaticName(name, 0, decor.DwidthSync),
				decor.CountersNoUnit("%d / %d", 10, decor.DSyncSpace),
			),
			mpb.AppendDecorators(
				decor.ETA(3, 0),
			),
		)
		go func() {
			defer wg.Done()
			for blockSize, i := 0, 0; i < total; i++ {
				blockSize = rand.Intn(maxBlockSize) + 1
				if i&1 == 1 {
					priority := total - int(b.Current())
					p.UpdateBarPriority(b, priority)
				}
				b.Increment()
				sleep(blockSize)
			}
		}()
	}

	p.Stop()
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
