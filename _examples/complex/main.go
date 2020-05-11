package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	doneWg := new(sync.WaitGroup)
	p := mpb.New(mpb.WithWaitGroup(doneWg))
	numBars := 4

	var bars []*mpb.Bar
	var downloadWgg []*sync.WaitGroup
	for i := 0; i < numBars; i++ {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		downloadWgg = append(downloadWgg, wg)
		task := fmt.Sprintf("Task#%02d:", i)
		job := "downloading"
		b := p.AddBar(rand.Int63n(201)+100,
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
				decor.Name(job, decor.WCSyncSpaceR),
				decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
			),
			mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
		)
		go newTask(wg, b, i+1)
		bars = append(bars, b)
	}

	for i := 0; i < numBars; i++ {
		doneWg.Add(1)
		i := i
		go func() {
			task := fmt.Sprintf("Task#%02d:", i)
			// ANSI escape sequences are not supported on Windows OS
			job := "\x1b[31;1;4mつのだ☆HIRO\x1b[0m"
			// preparing delayed bars
			b := p.AddBar(rand.Int63n(101)+100,
				mpb.BarQueueAfter(bars[i]),
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
			// waiting for download to complete, before starting install job
			downloadWgg[i].Wait()
			go newTask(doneWg, b, numBars-i)
		}()
	}

	p.Wait()
}

func newTask(wg *sync.WaitGroup, bar *mpb.Bar, incrBy int) {
	defer wg.Done()
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
