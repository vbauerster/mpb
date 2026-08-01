package main

import (
	"fmt"
	"io"
	"math/rand"
	"slices"
	"sync/atomic"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const numTasks = 4

var curTask, doneTasks atomic.Uint32

type task struct {
	id    uint32
	total int64
	bar   *mpb.Bar
}

func main() {

	var tasks []*task

	for i := range numTasks {
		task := &task{
			id:    uint32(i),
			total: rand.Int63n(666) + 100,
		}
		tasks = append(tasks, task)
	}

	p := mpb.New()

	fillers := toBarFillers(tasks)

	for i := range numTasks {
		bar := p.AddBar(tasks[i].total,
			mpb.BarTopExtender(fillers...),
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

	tb := p.AddBar(numTasks,
		mpb.PrependDecorators(
			decor.Any(func(st decor.Statistics) string {
				return fmt.Sprintf("TOTAL(%d/%d)", doneTasks.Load(), len(tasks))
			}, decor.WCSyncWidthR),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncWidth),
		),
	)

	max := 100 * time.Millisecond
	for _, t := range tasks {
		curTask.Store(t.id)
		for !t.bar.Completed() {
			n := rand.Int63n(10) + 1
			t.bar.IncrInt64(n)
			time.Sleep(time.Duration(n) * max / 10)
		}
		doneTasks.Add(1)
		tb.IncrBy(1)
		t.bar.Wait()
	}

	p.Wait()
}

func toBarFillers(tasks []*task) []mpb.BarFiller {
	var fillers []mpb.BarFiller
	filler := mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) (err error) {
		_, err = fmt.Fprintln(w)
		return
	})
	fillers = append(fillers, filler)
	for _, task := range slices.Backward(tasks) {
		if task == nil {
			continue
		}
		var done bool
		filler = mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) (err error) {
			if task.id == curTask.Load() {
				if !st.Completed {
					_, err = fmt.Fprintf(w, "=> Taksk %02d\n", task.id)
					return
				}
				done = true
			}
			if done {
				_, err = fmt.Fprintf(w, "   Taksk %02d: Done!\n", task.id)
				return
			}
			_, err = fmt.Fprintf(w, "   Taksk %02d\n", task.id)
			return
		})
		fillers = append(fillers, filler)
	}
	return fillers
}
