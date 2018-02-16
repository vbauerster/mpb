package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const (
	totalBars    = 32
	maxBlockSize = 8
)

func main() {

	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(120) + 10
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.StaticName(name, len(name), 0),
				decor.ETA(4, decor.DSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Percentage(5, 0),
			),
		)

		go func() {
			defer wg.Done()
			blockSize := rand.Intn(maxBlockSize) + 1
			for i := 0; i < total; i++ {
				sleep(blockSize)
				bar.Increment()
				blockSize = rand.Intn(maxBlockSize) + 1
			}
		}()
	}

	p.Stop()
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
