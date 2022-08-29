package mpb_test

import (
	"context"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

const (
	timeout = 200 * time.Millisecond
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestBarCount(t *testing.T) {
	shutdown := make(chan struct{})
	p := mpb.New(mpb.WithShutdownNotifier(shutdown), mpb.WithOutput(io.Discard))

	b := p.AddBar(0, mpb.BarRemoveOnComplete())

	if count := p.BarCount(); count != 1 {
		t.Errorf("BarCount want: %d, got: %d\n", 1, count)
	}

	b.SetTotal(100, true)

	b.Wait()

	if count := p.BarCount(); count != 0 {
		t.Errorf("BarCount want: %d, got: %d\n", 0, count)
	}

	go p.Wait()

	select {
	case <-shutdown:
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestBarAbort(t *testing.T) {
	shutdown := make(chan struct{})
	p := mpb.New(mpb.WithShutdownNotifier(shutdown), mpb.WithOutput(io.Discard))
	n := 2
	bars := make([]*mpb.Bar, n)
	for i := 0; i < n; i++ {
		b := p.AddBar(100)
		switch i {
		case n - 1:
			var abortCalledTimes int
			for j := 0; !b.Aborted(); j++ {
				if j >= 10 {
					b.Abort(true)
					abortCalledTimes++
				} else {
					b.Increment()
				}
			}
			if abortCalledTimes != 1 {
				t.Errorf("Expected abortCalledTimes: %d, got: %d\n", 1, abortCalledTimes)
			}
			b.Wait()
			count := p.BarCount()
			if count != 1 {
				t.Errorf("BarCount want: %d, got: %d\n", 1, count)
			}
		default:
			go func() {
				for !b.Completed() {
					b.Increment()
					time.Sleep(randomDuration(100 * time.Millisecond))
				}
			}()
		}
		bars[i] = b
	}

	bars[0].Abort(false)

	go p.Wait()

	select {
	case <-shutdown:
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestWithContext(t *testing.T) {
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx, mpb.WithShutdownNotifier(shutdown), mpb.WithOutput(io.Discard))

	done := make(chan struct{})
	bar := p.AddBar(0) // never complete bar
	go func() {
		for !bar.Aborted() {
			time.Sleep(randomDuration(100 * time.Millisecond))
			cancel()
		}
		close(done)
	}()

	go func() {
		<-done
		p.Wait()
	}()

	select {
	case <-shutdown:
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

// MaxWidthDistributor shouldn't stuck in the middle while removing or aborting a bar
func TestMaxWidthDistributor(t *testing.T) {

	makeWrapper := func(f func([]chan int), start, end chan struct{}) func([]chan int) {
		return func(column []chan int) {
			start <- struct{}{}
			f(column)
			<-end
		}
	}

	ready := make(chan struct{})
	start := make(chan struct{})
	end := make(chan struct{})
	mpb.MaxWidthDistributor = makeWrapper(mpb.MaxWidthDistributor, start, end)

	total := 100
	numBars := 6
	p := mpb.New(mpb.WithOutput(io.Discard))
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(total),
			mpb.BarOptional(mpb.BarRemoveOnComplete(), i == 0),
			mpb.PrependDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncSpace)),
		)
		go func() {
			<-ready
			for i := 0; i < total; i++ {
				start := time.Now()
				if id := bar.ID(); id > 1 && i >= 32 {
					if id&1 == 1 {
						bar.Abort(true)
					} else {
						bar.Abort(false)
					}
				}
				time.Sleep(randomDuration(100 * time.Millisecond))
				bar.IncrInt64(rand.Int63n(5) + 1)
				bar.DecoratorEwmaUpdate(time.Since(start))
			}
		}()
	}

	go func() {
		<-ready
		p.Wait()
		close(start)
	}()

	res := t.Run("maxWidthDistributor", func(t *testing.T) {
		close(ready)
		for v := range start {
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case end <- v:
				timer.Stop()
			case <-timer.C:
				t.FailNow()
			}
		}
	})

	if !res {
		t.Error("maxWidthDistributor stuck in the middle")
	}
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
