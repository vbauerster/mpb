package main

import (
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
	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithWidth(13),
	)
	total, numBars := 101, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		var bar *mpb.Bar
		if i == 0 {
			bar = p.AddBar(int64(total),
				// set custom bar style, default one is "[=>-]"
				mpb.BarStyle("╢▌▌░╟"),
				mpb.PrependDecorators(
					// simple name decorator
					decor.Name(name),
				),
				mpb.AppendDecorators(
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60
						decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
					),
				),
			)
		} else {
			bar = p.AddSpinner(int64(total), mpb.SpinnerOnMiddle,
				// set custom spinner style, default one is {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				mpb.SpinnerStyle([]string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"}),
				mpb.PrependDecorators(
					// simple name decorator
					decor.Name(name),
				),
				mpb.AppendDecorators(
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60
						decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
					),
				),
			)
		}

		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				// ewma based decorators require work duration measurement
				bar.IncrBy(1, time.Since(start))
			}
		}()
	}
	// wait for all bars to complete and flush
	p.Wait()
}
