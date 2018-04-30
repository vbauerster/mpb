package main

import (
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

	// initialize bar with dynamic total and initial total guess = 80
	bar := p.AddBar(80,
		// indicate that total is dynamic
		mpb.BarDynamicTotal(),
		// trigger total auto increment by 1, when 18 % remains till bar completion
		mpb.BarAutoIncrTotal(18, 1),
		mpb.PrependDecorators(
			decor.CountersNoUnit("%d / %d", 12, 0),
		),
		mpb.AppendDecorators(
			decor.Percentage(4, 0),
		),
	)

	totalUpd := make(chan int64)
	go func() {
		for {
			total, ok := <-totalUpd
			bar.SetTotal(total, !ok)
			if !ok {
				break
			}
		}
	}()

	max := 100 * time.Millisecond
	for i := 0; !bar.Completed(); i++ {
		if i == 140 {
			totalUpd <- 190
		}
		if i == 250 {
			totalUpd <- 300
			// final upd, so closing channel
			close(totalUpd)
		}
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}

	p.Wait()
}
