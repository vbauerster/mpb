package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

var quietMode bool

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.BoolVar(&quietMode, "q", false, "quiet mode")
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	// pass &wg (optional), so p will wait for it eventually
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.ContainerOptOnCond(
			// setting to nil will:
			// set output to ioutil.Discard and disable internal refresh rate
			// cycling, in order to not consume much CPU, hovewer a single refresh
			// still will be triggered on bar complete event, per each bar.
			mpb.WithOutput(nil),
			func() bool { return quietMode },
		),
	)
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with ewma age of 60
					decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
				),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				// since ewma decorator is used, we need to pass time.Since(start)
				bar.Increment(time.Since(start))
			}
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()
	fmt.Println("done")
}
