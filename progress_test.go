package mpb_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestBarCount(t *testing.T) {
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	check := make(chan struct{})
	b := p.AddBar(100)
	go func() {
		for i := 0; i < 100; i++ {
			if i == 10 {
				close(check)
			}
			b.Increment()
			time.Sleep((time.Duration(rand.Intn(10)+1) * (10 * time.Millisecond)) / 2)
		}
	}()

	<-check
	count := p.BarCount()
	if count != 1 {
		t.Errorf("BarCount want: %q, got: %q\n", 1, count)
	}

	b.Abort(false)
	p.Wait()
}

func TestBarAbort(t *testing.T) {
	n := 2
	p := mpb.New(mpb.WithOutput(ioutil.Discard))
	bars := make([]*mpb.Bar, n)
	for i := 0; i < n; i++ {
		b := p.AddBar(100)
		switch i {
		case n - 1:
			var abortCalledTimes int
			for j := 0; !b.Completed(); j++ {
				if j >= 33 {
					b.Abort(true)
					abortCalledTimes++
				} else {
					b.Increment()
					time.Sleep((time.Duration(rand.Intn(10)+1) * (10 * time.Millisecond)) / 2)
				}
			}
			if abortCalledTimes != 1 {
				t.Errorf("Expected abortCalledTimes: %d, got: %d\n", 1, abortCalledTimes)
			}
			count := p.BarCount()
			if count != 1 {
				t.Errorf("BarCount want: %d, got: %d\n", 1, count)
			}
		default:
			go func() {
				for !b.Completed() {
					b.Increment()
					time.Sleep((time.Duration(rand.Intn(10)+1) * (10 * time.Millisecond)) / 2)
				}
			}()
		}
		bars[i] = b
	}

	bars[0].Abort(false)
	p.Wait()
}

func TestWithContext(t *testing.T) {
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx, mpb.WithShutdownNotifier(shutdown), mpb.WithOutput(ioutil.Discard))

	start := make(chan struct{})
	done := make(chan struct{})
	fail := make(chan struct{})
	bar := p.AddBar(0) // never complete bar
	go func() {
		close(start)
		for !bar.Completed() {
			bar.Increment()
			time.Sleep(randomDuration(100 * time.Millisecond))
		}
		close(done)
	}()

	go func() {
		select {
		case <-done:
			p.Wait()
		case <-time.After(150 * time.Millisecond):
			close(fail)
		}
	}()

	<-start
	cancel()
	select {
	case <-shutdown:
	case <-fail:
		t.Error("Progress didn't shutdown")
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

	total := 80
	numBars := 6
	p := mpb.New(mpb.WithOutput(ioutil.Discard))
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(total),
			mpb.BarOptional(mpb.BarRemoveOnComplete(), i == 0),
			mpb.PrependDecorators(
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncSpace),
			),
		)
		go func() {
			<-ready
			for i := 0; i < total; i++ {
				start := time.Now()
				if id := bar.ID(); id > 1 && i >= 42 {
					if id&1 == 1 {
						bar.Abort(true)
					} else {
						bar.Abort(false)
					}
				}
				time.Sleep((time.Duration(rand.Intn(10)+1) * (50 * time.Millisecond)) / 2)
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

func getLastLine(bb []byte) []byte {
	split := bytes.Split(bb, []byte("\n"))
	return split[len(split)-2]
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
