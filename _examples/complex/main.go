package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	numBars := 4
	// to support color in Windows following both options are required
	p := mpb.New(
		mpb.WithOutput(color.Output),
		mpb.WithAutoRefresh(),
	)

	red, green := color.New(color.FgRed), color.New(color.FgGreen)

	for i := 0; i < numBars; i++ {
		task := fmt.Sprintf("Task#%02d:", i)
		b1 := p.AddBar(rand.Int63n(101)+100,
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{C: decor.DindentRight | decor.DextraSpace}),
				decor.Name("downloading", decor.WCSyncSpaceR),
				decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
			),
		)
		go complete(b1)

		// b2 is queued and will run after b1 is done replacing its place
		b2 := p.AddBar(rand.Int63n(101)+100,
			mpb.BarQueueAfter(b1),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{C: decor.DindentRight | decor.DextraSpace}),
				decor.OnCompleteMeta(
					decor.OnComplete(
						decor.Meta(decor.Name("installing", decor.WCSyncSpaceR), toMetaFunc(red)),
						"done!",
					),
					toMetaFunc(green),
				),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
			),
		)
		go complete(b2)
	}

	p.Wait()
}

func complete(bar *mpb.Bar) {
	max := 100 * time.Millisecond
	for !bar.Completed() {
		// start variable is solely for EWMA calculation
		// EWMA's unit of measure is an iteration's duration
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// we need to call EwmaIncrement to fulfill ewma decorator's contract
		bar.EwmaIncrInt64(rand.Int63n(5)+1, time.Since(start))
	}
}

func toMetaFunc(c *color.Color) func(string) string {
	return func(s string) string {
		return c.Sprint(s)
	}
}
