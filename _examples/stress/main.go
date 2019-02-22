package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

const (
	totalBars = 32
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithRefreshRate(50*time.Millisecond),
	)
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(320) + 10
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name),
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.Percentage(decor.WC{W: 5}), "done",
				),
			),
		)

		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for !bar.Completed() {
				start := time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				// ewma based decorators require work duration measurement
				bar.IncrBy(1, time.Since(start))
			}
		}()
	}

	p.Wait()
}
