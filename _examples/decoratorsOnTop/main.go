package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	p := mpb.New()

	total := 100
	bar := p.New(int64(total),
		mpb.NopStyle(), // make main bar style nop, so there are just decorators
		mpb.BarExtender(extended(mpb.BarStyle().Build()), false), // extend wtih normal bar on the next line
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

func extended(base mpb.BarFiller) mpb.BarFiller {
	return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
		err := base.Fill(w, st)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\n")
		return err
	})
}
