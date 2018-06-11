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
	totalBars = 32
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(120) + 10
		startBlock := make(chan time.Time)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name),
				decor.ETA(decor.ET_STYLE_GO, 0, startBlock, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 5}),
			),
		)

		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				startBlock <- time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}

	p.Wait()
}
