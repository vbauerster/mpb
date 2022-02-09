package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	numBars := 4
	bars := make([]*mpb.Bar, numBars)
	p := mpb.New()

	for i := 0; i < numBars; i++ {
		task := fmt.Sprintf("Task#%02d:", i)
		bars[i] = p.AddBar(rand.Int63n(201)+100,
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
				decor.Name("downloading", decor.WCSyncSpaceR),
				decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
			),
			mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
		)
		go complete(bars[i], i+1)
	}

	for i := 0; i < numBars; i++ {
		afterBar := bars[i]
		task := fmt.Sprintf("Task#%02d:", i)
		incrBy := numBars - i
		go func() {
			job := "\x1b[31minstalling\x1b[0m"
			// preparing queued bars
			b := p.AddBar(rand.Int63n(101)+100,
				mpb.BarQueueAfter(afterBar),
				mpb.BarFillerClearOnComplete(),
				mpb.PrependDecorators(
					decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
					decor.OnComplete(decor.Name(job, decor.WCSyncSpaceR), "done!"),
					decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
				),
			)
			complete(b, incrBy) // blocks until afterBar completes
		}()
	}

	p.Wait()
}

func complete(bar *mpb.Bar, incrBy int) {
	max := 100 * time.Millisecond
	for !bar.Completed() {
		// start variable is solely for EWMA calculation
		// EWMA's unit of measure is an iteration's duration
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.IncrBy(incrBy)
		// we need to call DecoratorEwmaUpdate to fulfill ewma decorator's contract
		bar.DecoratorEwmaUpdate(time.Since(start))
	}
}
