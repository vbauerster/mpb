package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	p := mpb.New(mpb.PopCompletedMode())
	total, numBars := 100, 4
	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.BarFillerOnComplete(fmt.Sprintf("%s has been completed", name)),
			mpb.BarFillerTrim(),
			mpb.PrependDecorators(
				decor.OnComplete(decor.Name(name), ""),
				decor.OnComplete(decor.NewPercentage(" % d "), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Name(" "), ""),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), ""),
			),
		)
		// simulating some work
		max := 100 * time.Millisecond
		for i := 0; i < total; i++ {
			// start variable is solely for EWMA calculation
			// EWMA's unit of measure is an iteration's duration
			start := time.Now()
			time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
			// we need to call EwmaIncrement to fulfill ewma decorator's contract
			bar.EwmaIncrement(time.Since(start))
		}
	}

	p.Wait()
}
