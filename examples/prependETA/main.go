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
		sbEta := make(chan time.Time)
		b := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name, decor.WCSyncWidth),
				decor.OnComplete(
					decor.ETA(decor.ET_STYLE_MMSS, 60, sbEta, decor.WC{W: 6}),
					"Done",
					decor.WCSyncSpace,
				),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 5}),
			),
		)
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				sbEta <- time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				b.Increment()
			}
		}()
	}

	p.Wait()
}
