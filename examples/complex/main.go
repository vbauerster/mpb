package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	doneWg := new(sync.WaitGroup)
	p := mpb.New(mpb.WithWidth(64), mpb.WithWaitGroup(doneWg))
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
			mpb.BarRemoveOnComplete(),
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
			job := "installing"
			// preparing delayed bars
			b := p.AddBar(rand.Int63n(101)+100,
				mpb.BarReplaceOnComplete(bars[i]),
				mpb.BarClearOnComplete(),
				mpb.PrependDecorators(
					decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
					decor.OnComplete(decor.Name(job, decor.WCSyncSpaceR), "done!", decor.WCSyncSpaceR),
					decor.OnComplete(
						decor.EwmaETA(decor.ET_STYLE_MMSS, 60, decor.WCSyncWidth), "", decor.WCSyncSpace,
					),
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

func newTask(wg *sync.WaitGroup, b *mpb.Bar, incrBy int) {
	defer wg.Done()
	max := 100 * time.Millisecond
	for !b.Completed() {
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// ewma based decorators require work duration measurement
		b.IncrBy(incrBy, time.Since(start))
	}
}
