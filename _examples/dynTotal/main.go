package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	p := mpb.New(mpb.WithWidth(64))

	// new bar with 'trigger complete event' disabled, because total is zero
	bar := p.AddBar(0,
		mpb.PrependDecorators(decor.Counters(decor.SizeB1024(0), "% .1f / % .1f")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	maxSleep := 100 * time.Millisecond
	read := makeStream(200)
	for {
		n, err := read()
		if err == io.EOF {
			// triggering complete event now
			bar.SetTotal(-1, true)
			break
		}
		// increment methods won't trigger complete event because bar was constructed with total = 0
		bar.IncrBy(n)
		// following call is not required, it's called to show some progress instead of an empty bar
		bar.SetTotal(bar.Current()+2048, false)
		time.Sleep(time.Duration(rand.Intn(10)+1) * maxSleep / 10)
	}

	p.Wait()
}

func makeStream(limit int) func() (int, error) {
	return func() (int, error) {
		if limit <= 0 {
			return 0, io.EOF
		}
		limit--
		return rand.Intn(1024) + 1, nil
	}
}
