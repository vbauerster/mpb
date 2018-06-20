//+build go1.7

package main

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithContext(ctx),
	)
	total := 300
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name),
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 5}),
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
