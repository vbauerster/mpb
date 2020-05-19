package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func main() {
	p := mpb.New()

	total := 100
	bar := p.Add(int64(total), nil,
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
		mpb.BarExtender(nlBarFiller(mpb.NewBarFiller("╢▌▌░╟", false))),
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

func nlBarFiller(filler mpb.BarFiller) mpb.BarFiller {
	return mpb.BarFillerFunc(func(w io.Writer, reqWidth int, st decor.Statistics) {
		filler.Fill(w, reqWidth, st)
		w.Write([]byte("\n"))
	})
}
