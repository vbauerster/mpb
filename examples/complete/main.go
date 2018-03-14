package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
			for i := 0; i < total; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+1) * (200 * time.Millisecond) / 10)
				b.Increment()
				if b.ID() == 1 && i == 42 {
					b.Complete()
					return
				}
			}
		}()
	}

	p.Stop()
	fmt.Println("stop")
}
