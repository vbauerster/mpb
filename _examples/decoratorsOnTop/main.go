package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func main() {
	p := mpb.New()

	total := 100
	bar := p.New(int64(total),
		mpb.NopStyle(), // make main bar style nop, so there are just decorators
		mpb.BarExtender(extended(mpb.BarStyle())), // extend wtih normal bar on the next line
		mpb.PrependDecorators(
			decor.Name("Percentage: "),
			decor.NewPercentage("%d"),
		),
		mpb.AppendDecorators(
			decor.Name("ETA: "),
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO), "done",
			),
		),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}
	// wait for our bar to complete and flush
	p.Wait()
}

func extended(builder mpb.BarFillerBuilder) mpb.BarFiller {
	filler := builder.Build()
	return mpb.BarFillerFunc(func(w io.Writer, reqWidth int, st decor.Statistics) {
		filler.Fill(w, reqWidth, st)
		w.Write([]byte("\n"))
	})
}
