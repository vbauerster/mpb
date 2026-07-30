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

	var written int64
	maxSleep := 100 * time.Millisecond
	read := makeStream(200)
	for {
		n, err := read()
		if err == io.EOF {
			break
		}
		written += int64(n)
		// following call is not required, it's called to show some progress instead of an empty bar
		bar.SetTotal(written+1024, false)
		// increment method won't trigger completion because bar was constructed with total = 0
		bar.IncrBy(n)
		time.Sleep(time.Duration(rand.Intn(10)+1) * maxSleep / 10)
	}

	bar.SetTotal(written, true)

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
