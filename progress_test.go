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

func TestBarPriorityPopOrder(t *testing.T) {
	shutdown := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
	)
	a := p.AddBar(100, mpb.BarPriority(1))
	b := p.AddBar(100, mpb.BarPriority(2))
	c := p.AddBar(100, mpb.BarPriority(3))

	identity := map[*mpb.Bar]string{
		a: "a",
		b: "b",
		c: "c",
	}

	cancel()

	bars := (<-shutdown).([]*mpb.Bar)
	if l := len(bars); l != 3 {
		t.Errorf("Expected len of bars: %d, got: %d", 3, l)
	}

	p.Wait()
	pq := mpb.PriorityQueue(bars)

	if bar := heap.Pop(&pq).(*mpb.Bar); bar != c {
		t.Errorf("Expected bar c, got: %s", identity[bar])
	}
	if bar := heap.Pop(&pq).(*mpb.Bar); bar != b {
		t.Errorf("Expected bar b, got: %s", identity[bar])
	}
	if bar := heap.Pop(&pq).(*mpb.Bar); bar != a {
		t.Errorf("Expected bar a, got: %s", identity[bar])
	}
}

func TestUpdateBarPriority(t *testing.T) {
	shutdown := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
	)
	a := p.AddBar(100, mpb.BarPriority(1))
	b := p.AddBar(100, mpb.BarPriority(2))
	c := p.AddBar(100, mpb.BarPriority(3))

	identity := map[*mpb.Bar]string{
		a: "a",
		b: "b",
		c: "c",
	}

	p.UpdateBarPriority(c, 2)
	p.UpdateBarPriority(b, 3)

	cancel()

	bars := (<-shutdown).([]*mpb.Bar)
	if l := len(bars); l != 3 {
		t.Errorf("Expected len of bars: %d, got: %d", 3, l)
	}

	p.Wait()
	pq := mpb.PriorityQueue(bars)

	if bar := heap.Pop(&pq).(*mpb.Bar); bar != b {
		t.Errorf("Expected bar b, got: %s", identity[bar])
	}
	if bar := heap.Pop(&pq).(*mpb.Bar); bar != c {
		t.Errorf("Expected bar c, got: %s", identity[bar])
	}
	if bar := heap.Pop(&pq).(*mpb.Bar); bar != a {
		t.Errorf("Expected bar a, got: %s", identity[bar])
	}
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
