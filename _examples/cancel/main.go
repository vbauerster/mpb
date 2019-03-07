package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	p := mpb.NewWithContext(ctx, mpb.WithWaitGroup(&wg))
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
				// note that OnComplete will not be fired, because of cancel
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
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
