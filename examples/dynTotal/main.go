package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	p := mpb.New(mpb.WithWidth(64))

	bar := p.AddBar(0, // by setting total to 0, we indicate that it's dynamic
		mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% .1f / % .1f")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	var written int64
	maxSleep := 100 * time.Millisecond
	read := makeStream(200)
	for {
		n, err := read()
		written += int64(n)
		time.Sleep(time.Duration(rand.Intn(10)+1) * maxSleep / 10)
		bar.IncrBy(n)
		if err == io.EOF {
			break
		}
		bar.SetTotal(written+1024, false)
	}

	// final set total, final=true
	bar.SetTotal(written, true)
	// need to increment once, to shutdown the bar
	bar.IncrBy(0)

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
