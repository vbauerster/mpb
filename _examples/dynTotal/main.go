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

	// Start with total = 0 because we don't know the final size yet.
	// A bar created with total = 0 will NOT auto-complete when current
	// reaches total, so we must call SetTotal(final, true) at the end.
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
		// SetTotal(written+1024, false) does two things:
		//   1. Updates the denominator so Percentage() and Counters show
		//      meaningful progress instead of "0 / 0".
		//   2. The `false` argument means "do NOT trigger completion" —
		//      we're still streaming and this is just a running estimate.
		bar.SetTotal(written+1024, false)
		bar.IncrBy(n)
		time.Sleep(time.Duration(rand.Intn(10)+1) * maxSleep / 10)
	}

	// Now that we know the exact total, set it with `true` to trigger
	// the completion event and let p.Wait() return.
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
