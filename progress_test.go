package mpb_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	b := p.AddBar(100)
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 100; i++ {
			if i == 33 {
				wg.Done()
			}
			b.Increment()
			time.Sleep((time.Duration(rng.Intn(10)+1) * (10 * time.Millisecond)) / 2)
		}
	}()

	wg.Wait()
	count := p.BarCount()
	if count != 1 {
		t.Errorf("BarCount want: %q, got: %q\n", 1, count)
	}

	b.Abort(true)
	p.Wait()
}

func TestBarAbort(t *testing.T) {
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	var wg sync.WaitGroup
	wg.Add(1)
	bars := make([]*mpb.Bar, 3)
	for i := 0; i < 3; i++ {
		b := p.AddBar(100)
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		go func(n int) {
			for i := 0; !b.Completed(); i++ {
				if n == 0 && i >= 33 {
					b.Abort(true)
					wg.Done()
				}
				b.Increment()
				time.Sleep((time.Duration(rng.Intn(10)+1) * (10 * time.Millisecond)) / 2)
			}
		}(i)
		bars[i] = b
	}

	wg.Wait()
	count := p.BarCount()
	if count != 2 {
		t.Errorf("BarCount want: %d, got: %d\n", 2, count)
	}
	bars[1].Abort(true)
	bars[2].Abort(true)
	p.Wait()
}

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(ioutil.Discard),
		mpb.WithRefreshRate(50*time.Millisecond),
		mpb.WithShutdownNotifier(shutdown),
	)

	total := 10000
	numBars := 3
	bars := make([]*mpb.Bar, 0, numBars)
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(total))
		bars = append(bars, bar)
		go func() {
			for !bar.Completed() {
				bar.Increment()
				time.Sleep(randomDuration(100 * time.Millisecond))
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	p.Wait()
	select {
	case <-shutdown:
	case <-time.After(100 * time.Millisecond):
		t.Error("Progress didn't stop")
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
	*mpb.MaxWidthDistributor = makeWrapper(*mpb.MaxWidthDistributor, start, end)

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
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < total; i++ {
				start := time.Now()
				if id := bar.ID(); id > 1 && i >= 42 {
					if id&1 == 1 {
						bar.Abort(true)
					} else {
						bar.Abort(false)
					}
				}
				time.Sleep((time.Duration(rng.Intn(10)+1) * (50 * time.Millisecond)) / 2)
				bar.Increment()
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
