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
			mpb.BarExtender(filler, true), // all bars share same extender filler
			mpb.BarFuncOptional(func() mpb.BarOption {
				return mpb.BarQueueAfter(tasks[i-1].bar)
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
		complete(t.bar, tb)
		atomic.AddUint32(&doneTasks, 1)
	}

	tb.EnableTriggerComplete()

	p.Wait()
}

func middleware(base mpb.BarFiller, id uint32) mpb.BarFiller {
	var done bool
	fn := func(w io.Writer, st decor.Statistics) error {
		if !done {
			if atomic.LoadUint32(&curTask) != id {
				_, err := fmt.Fprintf(w, "   Taksk %02d\n", id)
				return err
			}
			if !st.Completed {
				_, err := fmt.Fprintf(w, "=> Taksk %02d\n", id)
				return err
			}
			done = true
		}
		_, err := fmt.Fprintf(w, "   Taksk %02d: Done!\n", id)
		return err
	}
	if base == nil {
		return mpb.BarFillerFunc(fn)
	}
	return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
		err := fn(w, st)
		if err != nil {
			return err
		}
		return base.Fill(w, st)
	})
}

func newLineMiddleware(base mpb.BarFiller) mpb.BarFiller {
	return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
		_, err := fmt.Fprintln(w)
		if err != nil {
			return err
		}
		return base.Fill(w, st)
	})
}

func complete(bar, totalBar *mpb.Bar) {
	max := 100 * time.Millisecond
	for !bar.Completed() {
		n := rand.Int63n(10) + 1
		incrementBars(n, bar, totalBar)
		time.Sleep(time.Duration(n) * max / 10)
	}
	bar.Wait()
}

func incrementBars(n int64, bb ...*mpb.Bar) {
	for _, b := range bb {
		b.IncrInt64(n)
	}
}
