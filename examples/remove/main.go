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
		b := p.AddBar(int64(total), mpb.BarID(i),
			mpb.PrependDecorators(
				decor.StaticName(name, 0, decor.DwidthSync|decor.DidentRight),
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
				b.Incr(1)
				blockSize = rand.Intn(maxBlockSize) + 1
				if b.ID() == 1 && i >= 42 {
					p.RemoveBar(b)
					return
				}
			}
		}()
	}

	p.Stop()
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
