package main

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func main() {
	var wg sync.WaitGroup
	// pass &wg (optional), so p will wait for it eventually
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(60))
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		var pdecorators mpb.BarOption
		if i == 0 {
			pdecorators = mpb.PrependDecorators(
				// Merge to sync width with decorators on lines 37 and 38
				decor.Merge(
					// decor.OnComplete(newVariadicSpinner(decor.WCSyncSpace), "done"),
					newVariadicSpinner(decor.WCSyncSpace),
					decor.WCSyncSpace, // Placeholder
				),
			)
		} else {
			pdecorators = mpb.PrependDecorators(
				decor.CountersNoUnit("% .1d / % .1d", decor.WCSyncSpace),
				decor.OnComplete(decor.Spinner(nil, decor.WCSyncSpace), "done"),
			)
		}
		bar := p.AddBar(int64(total),
			pdecorators,
			mpb.AppendDecorators(
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), "done"),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				// since ewma decorator is used, we need to pass time.Since(start)
				bar.Increment(time.Since(start))
			}
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()
}

func newVariadicSpinner(wc decor.WC) decor.Decorator {
	wc.Init()
	d := &variadicSpinner{
		WC: wc,
		d:  decor.Spinner(nil),
	}
	return d
}

type variadicSpinner struct {
	decor.WC
	d        decor.Decorator
	complete *string
}

func (d *variadicSpinner) Decor(st *decor.Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(*d.complete)
	}
	msg := d.d.Decor(st)
	msg = strings.Repeat(msg, int(st.Current/3))
	return d.FormatMsg(msg)
}

func (d *variadicSpinner) OnCompleteMessage(msg string) {
	d.complete = &msg
}
