package mpb_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	timeout = 200 * time.Millisecond
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestBarCount(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	b := p.AddBar(0, mpb.BarRemoveOnComplete())

	if count := p.BarCount(); count != 1 {
		t.Errorf("BarCount want: %d, got: %d\n", 1, count)
	}

	b.SetTotal(100, true)

	b.Wait()

	if count := p.BarCount(); count != 0 {
		t.Errorf("BarCount want: %d, got: %d\n", 0, count)
	}

	p.Wait()
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

	go p.Wait()

	bars[0].Abort(false)

	select {
	case <-shutdown:
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestWithContext(t *testing.T) {
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(io.Discard),
	)
	_ = p.AddBar(0) // never complete bar
	_ = p.AddBar(0) // never complete bar
	go func() {
		time.Sleep(randomDuration(100 * time.Millisecond))
		cancel()
		p.Wait()
	}()

	select {
	case <-shutdown:
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

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
	// mpb.MaxWidthDistributor shouldn't stuck in the middle while removing or aborting a bar
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
				bar.EwmaIncrInt64(rand.Int63n(5)+1, time.Since(start))
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

func TestProgressShutdownsWithErrFiller(t *testing.T) {
	var debug bytes.Buffer
	shutdown := make(chan struct{})
	p := mpb.New(
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(io.Discard),
		mpb.WithDebugOutput(&debug),
	)

	testError := errors.New("test error")
	bar := p.AddBar(100,
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if st.Current >= 42 {
					return testError
				}
				return base.Fill(w, st)
			})
		}),
	)

	go func() {
		for bar.IsRunning() {
			bar.Increment()
		}
	}()

	select {
	case <-shutdown:
		if err := strings.TrimSpace(debug.String()); err != testError.Error() {
			t.Errorf("Expected err: %q, got %q\n", testError.Error(), err)
		}
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
	p.Wait()
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
