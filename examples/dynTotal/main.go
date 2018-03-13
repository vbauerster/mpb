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
	p := mpb.New()

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
			decor.Percentage(5, 0),
		),
	)

	totalUpd1 := make(chan struct{})
	totalUpd2 := make(chan struct{})
	go func() {
		<-totalUpd1
		// intermediate not final total update
		bar.SetTotal(200, false)
		<-totalUpd2
		// final total update
		bar.SetTotal(300, true)
	}()

	for i := 0; i < 300; i++ {
		if i == 140 {
			close(totalUpd1)
		}
		if i == 250 {
			close(totalUpd2)
		}
		time.Sleep(time.Duration(rand.Intn(10)+1) * (200 * time.Millisecond) / 10)
		bar.Increment()
	}

	p.Stop()
}
