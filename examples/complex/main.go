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
	p := mpb.New(mpb.WithWidth(64))
	total := 200
	numBars := 3

	var bars []*mpb.Bar
	var downloadGroup []*sync.WaitGroup
	for i := 0; i < numBars; i++ {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		downloadGroup = append(downloadGroup, wg)
		task := fmt.Sprintf("Task#%02d:", i)
		job := "downloading"
		b := p.AddBar(int64(total), mpb.BarRemoveOnComplete(),
			mpb.PrependDecorators(
				decor.StaticName(task, len(task)+1, decor.DidentRight),
				decor.StaticName(job, 0, decor.DSyncSpaceR),
				decor.CountersNoUnit("%d / %d", 0, decor.DwidthSync),
			),
			mpb.AppendDecorators(decor.Percentage(5, 0)),
		)
		go newTask(wg, b, i+1)
		bars = append(bars, b)
	}

	var installGroup []*sync.WaitGroup
	for i := 0; i < numBars; i++ {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		installGroup = append(installGroup, wg)
		i := i
		go func() {
			task := fmt.Sprintf("Task#%02d:", i)
			job := "installing"
			// preparing delayed bars
			b := p.AddBar(int64(total), mpb.BarReplaceOnComplete(bars[i]), mpb.BarClearOnComplete(),
				mpb.PrependDecorators(
					decor.StaticName(task, len(task)+1, decor.DidentRight),
					decor.OnComplete(decor.StaticName(job, 0, decor.DSyncSpaceR), "done!", 0, decor.DSyncSpaceR),
					decor.OnComplete(decor.ETA(0, decor.DwidthSync|decor.DslowMotion), "", 0, decor.DwidthSync),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(5, 0), "", 0, 0),
				),
			)
			// waiting for download to complete, before starting install job
			downloadGroup[i].Wait()
			go newTask(wg, b, numBars-i)
		}()
	}

	// this wait loop may be skipped, but not recommended
	for _, wg := range installGroup {
		wg.Wait()
	}

	p.Wait()
}

func newTask(wg *sync.WaitGroup, b *mpb.Bar, incrBy int) {
	defer wg.Done()
	max := 100 * time.Millisecond
	for !b.Completed() {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		b.IncrBy(incrBy)
	}
}
