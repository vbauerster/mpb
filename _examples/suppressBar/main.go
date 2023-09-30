package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	p := mpb.New()

	total, numBars := 100, 3
	err := new(errorWrapper)
	timer := time.AfterFunc(2*time.Second, func() {
		err.set(errors.New("timeout"), rand.Intn(numBars))
	})
	defer timer.Stop()

	for i := 0; i < numBars; i++ {
		msgCh := make(chan string, 1)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(newTitleDecorator(fmt.Sprintf("Bar#%d:", i), msgCh, 16)),
			mpb.AppendDecorators(decor.Percentage(decor.WCSyncWidth)),
		)
		// simulating some work
		barID := i
		go func() {
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				if err.check(barID) {
					msgCh <- fmt.Sprintf("%s at %d, retrying...", err.Error(), i)
					err.reset()
					i--
					bar.SetRefill(int64(i))
					continue
				}
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}

	p.Wait()
}

type errorWrapper struct {
	sync.RWMutex
	err   error
	barID int
}

func (ew *errorWrapper) Error() string {
	ew.RLock()
	defer ew.RUnlock()
	return ew.err.Error()
}

func (ew *errorWrapper) check(barID int) bool {
	ew.RLock()
	defer ew.RUnlock()
	return ew.err != nil && ew.barID == barID
}

func (ew *errorWrapper) set(err error, barID int) {
	ew.Lock()
	ew.err = err
	ew.barID = barID
	ew.Unlock()
}

func (ew *errorWrapper) reset() {
	ew.Lock()
	ew.err = nil
	ew.Unlock()
}

type title struct {
	decor.Decorator
	name  string
	msgCh <-chan string
	msg   string
	count int
	limit int
}

func (d *title) Decor(stat decor.Statistics) (string, int) {
	if d.count == 0 {
		select {
		case msg := <-d.msgCh:
			d.count = d.limit
			d.msg = msg
		default:
			return d.Decorator.Decor(stat)
		}
	}
	d.count--
	_, _ = d.Format("")
	return fmt.Sprintf("%s %s", d.name, d.msg), math.MaxInt
}

func newTitleDecorator(name string, msgCh <-chan string, limit int) decor.Decorator {
	return &title{
		Decorator: decor.Name(name),
		name:      name,
		msgCh:     msgCh,
		limit:     limit,
	}
}
