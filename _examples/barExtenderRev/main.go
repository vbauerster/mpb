package main

import (
	"fmt"
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var curTask uint32
var doneTasks uint32

type task struct {
	id    uint32
	total int64
	bar   *mpb.Bar
}

func main() {
	numTasks := 4

	var total int64
	var filler mpb.BarFiller
	tasks := make([]*task, numTasks)

	for i := 0; i < numTasks; i++ {
		task := &task{
			id:    uint32(i),
			total: rand.Int63n(666) + 100,
		}
		total += task.total
		filler = middleware(filler, task.id)
		tasks[i] = task
	}

	filler = newLineMiddleware(filler)

	p := mpb.New()

	for i := 0; i < numTasks; i++ {
		bar := p.AddBar(tasks[i].total,
			mpb.BarExtender(filler, true),
			mpb.BarFuncOptional(func() mpb.BarOption {
				return mpb.BarQueueAfter(tasks[i-1].bar, false)
			}, i != 0),
			mpb.PrependDecorators(
				decor.Name("current:", decor.WCSyncWidthR),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WCSyncWidth),
			),
		)
		tasks[i].bar = bar
	}

	tb := p.AddBar(0,
		mpb.PrependDecorators(
			decor.Any(func(st decor.Statistics) string {
				return fmt.Sprintf("TOTAL(%d/%d)", atomic.LoadUint32(&doneTasks), len(tasks))
			}, decor.WCSyncWidthR),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncWidth),
		),
	)

	tb.SetTotal(total, false)

	for _, t := range tasks {
		atomic.StoreUint32(&curTask, t.id)
		complete(tb, t)
		atomic.AddUint32(&doneTasks, 1)
	}

	tb.EnableTriggerComplete()

	p.Wait()
}

func middleware(base mpb.BarFiller, id uint32) mpb.BarFiller {
	var done bool
	fn := func(w io.Writer, st decor.Statistics) {
		if !done {
			cur := atomic.LoadUint32(&curTask) == id
			if !cur {
				fmt.Fprintf(w, "   Taksk %02d\n", id)
				return
			}
			if !st.Completed {
				fmt.Fprintf(w, "=> Taksk %02d\n", id)
				return
			}
			done = cur
		}
		fmt.Fprintf(w, "   Taksk %02d: Done!\n", id)
	}
	if base == nil {
		return mpb.BarFillerFunc(fn)
	}
	return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) {
		fn(w, st)
		base.Fill(w, st)
	})
}

func newLineMiddleware(base mpb.BarFiller) mpb.BarFiller {
	return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) {
		fmt.Fprintln(w)
		base.Fill(w, st)
	})
}

func complete(tb *mpb.Bar, t *task) {
	bar := t.bar
	max := 100 * time.Millisecond
	for !bar.Completed() {
		n := rand.Int63n(10) + 1
		bar.IncrInt64(n)
		go tb.IncrInt64(n)
		time.Sleep(time.Duration(n) * max / 10)
	}
	bar.Wait()
}
