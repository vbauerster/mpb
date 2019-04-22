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
	// pass &wg (optional), so p will wait for it eventually
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		var pdecorators mpb.BarOption
		if i == 0 {
			pdecorators = mpb.PrependDecorators(decor.Name(name),
				// Merge to sync width with CountersNoUnit and Percentage decorators
				decor.Merge(
					decor.OnComplete(variadicName(decor.WCSyncSpace), "done"),
					decor.WCSyncSpace, // Placeholder
				),
			)
		} else {
			pdecorators = mpb.PrependDecorators(decor.Name(name),
				decor.CountersNoUnit("% .1d / % .1d", decor.WCSyncSpace),
				decor.Percentage(decor.WCSyncSpace),
			)
		}
		bar := p.AddBar(int64(total),
			pdecorators,
			mpb.AppendDecorators(
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
				// ewma based decorators require work duration measurement
				bar.IncrBy(1, time.Since(start))
			}
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()
}

func variadicName(wc decor.WC) decor.Decorator {
	wc.Init()
	d := &varName{
		WC: wc,
	}
	return d
}

type varName struct {
	decor.WC
	complete *string
}

func (d *varName) Decor(st *decor.Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(*d.complete)
	}
	if st.Current < 30 {
		return d.FormatMsg("low low low")
	} else if st.Current < 70 {
		return d.FormatMsg("medium medium medium")
	} else {
		return d.FormatMsg("high high high high high high")
	}
}

func (d *varName) OnCompleteMessage(msg string) {
	d.complete = &msg
}
