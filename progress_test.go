package mpb_test

import (
	"bytes"
	"container/heap"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	timeout = 300 * time.Millisecond
)

func TestWithContext(t *testing.T) {
	shutdown := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
	)
	_ = p.AddBar(0) // never complete bar
	_ = p.AddBar(0) // never complete bar
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	p.Wait()

	select {
	case v := <-shutdown:
		if l := len(v.([]*mpb.Bar)); l != 2 {
			t.Errorf("Expected len of bars: %d, got: %d", 2, l)
		}
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownsWithErrFiller(t *testing.T) {
	var debug bytes.Buffer
	shutdown := make(chan interface{})
	p := mpb.New(
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(io.Discard),
		mpb.WithDebugOutput(&debug),
		mpb.WithAutoRefresh(),
	)

	var errReturnCount int
	testError := errors.New("test error")
	bar := p.AddBar(100,
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if st.Current >= 22 {
					errReturnCount++
					return testError
				}
				return base.Fill(w, st)
			})
		}),
	)

	go func() {
		for bar.IsRunning() {
			bar.Increment()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	p.Wait()

	if errReturnCount != 1 {
		t.Errorf("Expected errReturnCount: %d, got: %d\n", 1, errReturnCount)
	}

	select {
	case v := <-shutdown:
		if l := len(v.([]*mpb.Bar)); l != 0 {
			t.Errorf("Expected len of bars: %d, got: %d\n", 0, l)
		}
		if err := strings.TrimSpace(debug.String()); err != testError.Error() {
			t.Errorf("Expected err: %q, got %q\n", testError.Error(), err)
		}
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownAfterBarAbortWithDrop(t *testing.T) {
	shutdown := make(chan interface{})
	p := mpb.New(
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(io.Discard),
		mpb.WithAutoRefresh(),
	)
	b := p.AddBar(100)

	var count int
	for i := 0; !b.Aborted(); i++ {
		if i >= 10 {
			count++
			b.Abort(true)
		} else {
			b.Increment()
			time.Sleep(10 * time.Millisecond)
		}
	}

	p.Wait()

	if count != 1 {
		t.Errorf("Expected count: %d, got: %d", 1, count)
	}

	select {
	case v := <-shutdown:
		if l := len(v.([]*mpb.Bar)); l != 0 {
			t.Errorf("Expected len of bars: %d, got: %d", 0, l)
		}
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownAfterBarAbortWithNoDrop(t *testing.T) {
	shutdown := make(chan interface{})
	p := mpb.New(
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(io.Discard),
		mpb.WithAutoRefresh(),
	)
	b := p.AddBar(100)

	var count int
	for i := 0; !b.Aborted(); i++ {
		if i >= 10 {
			count++
			b.Abort(false)
		} else {
			b.Increment()
			time.Sleep(10 * time.Millisecond)
		}
	}

	p.Wait()

	if count != 1 {
		t.Errorf("Expected count: %d, got: %d", 1, count)
	}

	select {
	case v := <-shutdown:
		if l := len(v.([]*mpb.Bar)); l != 1 {
			t.Errorf("Expected len of bars: %d, got: %d", 1, l)
		}
	case <-time.After(timeout):
		t.Errorf("Progress didn't shutdown after %v", timeout)
	}
}

func TestBarPristinePopOrder(t *testing.T) {
	shutdown := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard), // auto refresh is disabled
		mpb.WithShutdownNotifier(shutdown),
	)
	a := p.AddBar(100, mpb.BarPriority(1), mpb.BarID(1))
	b := p.AddBar(100, mpb.BarPriority(2), mpb.BarID(2))
	c := p.AddBar(100, mpb.BarPriority(3), mpb.BarID(3))
	pristineOrder := []*mpb.Bar{c, b, a}

	go cancel()

	bars := (<-shutdown).([]*mpb.Bar)
	if l := len(bars); l != 3 {
		t.Fatalf("Expected len of bars: %d, got: %d", 3, l)
	}

	p.Wait()
	pq := mpb.PriorityQueue(bars)

	for _, b := range pristineOrder {
		// higher priority pops first
		if bar := heap.Pop(&pq).(*mpb.Bar); bar.ID() != b.ID() {
			t.Errorf("Expected bar id: %d, got bar id: %d", b.ID(), bar.ID())
		}
	}
}

func makeUpdateBarPriorityTest(refresh, lazy bool) func(*testing.T) {
	return func(t *testing.T) {
		shutdown := make(chan interface{})
		refreshCh := make(chan interface{})
		ctx, cancel := context.WithCancel(context.Background())
		p := mpb.NewWithContext(ctx,
			mpb.WithOutput(io.Discard),
			mpb.WithManualRefresh(refreshCh),
			mpb.WithShutdownNotifier(shutdown),
		)
		a := p.AddBar(100, mpb.BarPriority(1), mpb.BarID(1))
		b := p.AddBar(100, mpb.BarPriority(2), mpb.BarID(2))
		c := p.AddBar(100, mpb.BarPriority(3), mpb.BarID(3))

		p.UpdateBarPriority(c, 2, lazy)
		p.UpdateBarPriority(b, 3, lazy)
		checkOrder := []*mpb.Bar{b, c, a} // updated order

		if refresh {
			refreshCh <- time.Now()
		} else if lazy {
			checkOrder = []*mpb.Bar{c, b, a} // pristine order
		}

		go cancel()

		bars := (<-shutdown).([]*mpb.Bar)
		if l := len(bars); l != 3 {
			t.Fatalf("Expected len of bars: %d, got: %d", 3, l)
		}

		p.Wait()
		pq := mpb.PriorityQueue(bars)

		for _, b := range checkOrder {
			// higher priority pops first
			if bar := heap.Pop(&pq).(*mpb.Bar); bar.ID() != b.ID() {
				t.Errorf("Expected bar id: %d, got bar id: %d", b.ID(), bar.ID())
			}
		}
	}
}

func TestUpdateBarPriority(t *testing.T) {
	makeUpdateBarPriorityTest(false, false)(t)
	makeUpdateBarPriorityTest(true, false)(t)
}

func TestUpdateBarPriorityLazy(t *testing.T) {
	makeUpdateBarPriorityTest(false, true)(t)
	makeUpdateBarPriorityTest(true, true)(t)
}

func TestNoOutput(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(mpb.WithOutput(&buf))
	bar := p.AddBar(100)

	go func() {
		for !bar.Completed() {
			bar.Increment()
		}
	}()

	p.Wait()

	if buf.Len() != 0 {
		t.Errorf("Expected buf.Len == 0, got: %d\n", buf.Len())
	}
}

func TestAddAfterDone(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.AddBar(100)
	bar.IncrBy(100)

	p.Wait()

	_, err := p.Add(100, nil)

	if err != mpb.ErrDone {
		t.Errorf("Expected %q, got: %q\n", mpb.ErrDone, err)
	}
}
