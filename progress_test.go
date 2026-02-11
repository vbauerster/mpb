package mpb_test

import (
	"bytes"
	"container/heap"
	"context"
	"errors"
	"io"
	"slices"
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
	handOverBarHeap := make(chan []*mpb.Bar, 1)
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithHandOverBarHeap(handOverBarHeap),
	)
	_ = p.AddBar(0) // never complete bar

	cancel()

	select {
	case <-shutdown:
		p.Wait()
		select {
		case bars := <-handOverBarHeap:
			if l := len(bars); l != 1 {
				t.Errorf("Expected len of bars: %d, got: %d", 1, l)
			}
		default:
			t.Fatal("<-handOverBarHeap failure")
		}
	case <-time.After(timeout):
		t.Fatalf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownsWithErrFiller(t *testing.T) {
	var debug bytes.Buffer
	shutdown := make(chan interface{})
	handOverBarHeap := make(chan []*mpb.Bar, 1)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithDebugOutput(&debug),
		mpb.WithAutoRefresh(),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithHandOverBarHeap(handOverBarHeap),
	)

	testError := errors.New("test error")
	bar := p.AddBar(100,
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if st.Current > 21 {
					return testError
				}
				return base.Fill(w, st)
			})
		}),
	)

	go func() {
		for !bar.AbortedOrCompleted() {
			bar.Increment()
		}
		p.Wait()
	}()

	select {
	case <-shutdown:
		p.Wait()
		select {
		case bars := <-handOverBarHeap:
			if l := len(bars); l != 1 {
				t.Errorf("Expected len of bars: %d, got: %d", 1, l)
			}
			if !slices.Contains(bars, bar) {
				t.Errorf("Expected []*mpb.Bar to contain: %#v", bar)
			}
			if err := strings.TrimSpace(debug.String()); err != testError.Error() {
				t.Errorf("Expected err: %q, got %q", testError.Error(), err)
			}
		default:
			t.Fatal("<-handOverBarHeap failure")
		}
	case <-time.After(timeout):
		t.Fatalf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownAfterBarAbortWithDrop(t *testing.T) {
	shutdown := make(chan interface{})
	handOverBarHeap := make(chan []*mpb.Bar, 1)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithAutoRefresh(),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithHandOverBarHeap(handOverBarHeap),
	)
	b := p.AddBar(100)

	for i := 0; !b.Aborted(); i++ {
		if i > 0 {
			b.Abort(true)
		} else {
			b.Increment()
		}
	}

	go p.Wait()

	select {
	case <-shutdown:
		p.Wait()
		select {
		case bars := <-handOverBarHeap:
			if l := len(bars); l != 0 {
				t.Errorf("Expected len of bars: %d, got: %d", 0, l)
			}
		default:
			t.Fatal("<-handOverBarHeap failure")
		}
	case <-time.After(timeout):
		t.Fatalf("Progress didn't shutdown after %v", timeout)
	}
}

func TestShutdownAfterBarAbortWithNoDrop(t *testing.T) {
	shutdown := make(chan interface{})
	handOverBarHeap := make(chan []*mpb.Bar, 1)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithAutoRefresh(),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithHandOverBarHeap(handOverBarHeap),
	)
	b := p.AddBar(100)

	for i := 0; !b.Aborted(); i++ {
		if i > 0 {
			b.Abort(false)
		} else {
			b.Increment()
		}
	}

	go p.Wait()

	select {
	case <-shutdown:
		p.Wait()
		select {
		case bars := <-handOverBarHeap:
			if l := len(bars); l != 1 {
				t.Errorf("Expected len of bars: %d, got: %d", 1, l)
			}
		default:
			t.Fatal("<-handOverBarHeap failure")
		}
	case <-time.After(timeout):
		t.Fatalf("Progress didn't shutdown after %v", timeout)
	}
}

func TestBarPristinePopOrder(t *testing.T) {
	shutdown := make(chan interface{})
	handOverBarHeap := make(chan []*mpb.Bar, 1)
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard), // auto refresh is disabled
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithHandOverBarHeap(handOverBarHeap),
	)
	a := p.AddBar(100, mpb.BarPriority(1), mpb.BarID(1))
	b := p.AddBar(100, mpb.BarPriority(2), mpb.BarID(2))
	c := p.AddBar(100, mpb.BarPriority(3), mpb.BarID(3))
	pristineOrder := []*mpb.Bar{c, b, a}

	cancel()

	select {
	case <-shutdown:
		p.Wait()
		select {
		case bars := <-handOverBarHeap:
			if len(bars) != len(pristineOrder) {
				t.Fatalf("Expected len of bars: %d, got: %d", len(pristineOrder), len(bars))
			}
			for i, b := range bars {
				if bar := pristineOrder[i]; bar.ID() != b.ID() {
					t.Errorf("Expected bar id: %d, got bar id: %d", bar.ID(), b.ID())
				}
			}
		default:
			t.Fatal("<-handOverBarHeap failure")
		}
	case <-time.After(timeout):
		t.Fatalf("Progress didn't shutdown after %v", timeout)
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
		pq := mpb.BarHeap(bars)

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
