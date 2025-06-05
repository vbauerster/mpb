package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	var wg sync.WaitGroup
	// passed wg will be accounted at p.Wait() call
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(64))
	total, numBars := 100, 3
	wg.Add(numBars)

	fileToDownload := []string{
		"One Hundred Years of Solitude by Gabriel Garcia Marquez",
		"The Brothers Karamazov by Fyodor Dostoyevsky",
		"Alice's Adventures in Wonderland by Lewis Carroll",
	}

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				decor.OnComplete(
					// Marquee decorator with default style
					Marquee(fileToDownload[i], 20, decor.WCSyncSpace), "done",
				),
			),
			mpb.AppendDecorators(
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncWidth),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}
	// wait for passed wg and for all bars to complete and flush
	p.Wait()
}
