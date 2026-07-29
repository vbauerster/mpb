package mpb_test

import (
	"bytes"
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

func TestWithContextCancel(t *testing.T) {
	var barCount int
	shutdown, depleteHeap := make(chan any), make(chan *mpb.Bar, 1)
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithDepleteHeap(depleteHeap),
		mpb.WithAutoRefresh(),
	)

	b := p.AddBar(0) // never complete bar
	cancel()

test:
	for {
		select {
		case _, ok := <-depleteHeap:
			if !ok {
				break test
			}
			barCount++
		case <-shutdown:
			shutdown = nil
		case <-time.After(timeout):
			t.Fatalf("Test timeout %v", timeout)
		}
	}
	if barCount != 1 {
		t.Errorf("Expected to receive 1 bar, got: %d", barCount)
	}
	if b.Completed() {
		t.Error("Expected bar not to be completed")
	}
	if !b.Aborted() {
		t.Error("Expected bar to be aborted")
	}
}

func TestShutdownWithOneHourRefreshRate(t *testing.T) {
	var buf bytes.Buffer
	width, shutdown := 30, make(chan any)
	p := mpb.New(
		mpb.WithRefreshRate(time.Hour),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(&buf),
		mpb.WithWidth(width),
		mpb.WithAutoRefresh(),
	)

	go func(b *mpb.Bar) {
		b.IncrBy(100)
		p.Wait()
	}(p.AddBar(100))

	select {
	case <-shutdown:
		bar := " [" + strings.Repeat("=", width-4) + "] "
		if got, _, found := strings.Cut(buf.String(), "\n"); found {
			if got != bar {
				t.Errorf("want %q, got %q", bar, got)
			}
		} else {
			t.Fatal("Expected buf to contain some ' [=..] \\n'")
		}
	case <-time.After(timeout):
		t.Fatalf("Test timeout %v", timeout)
	}
}

func TestShutdownWithManualRefreshNeverFires(t *testing.T) {
	var buf bytes.Buffer
	width, shutdown := 30, make(chan any)
	p := mpb.New(
		mpb.WithManualRefresh(make(chan any)),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithOutput(&buf),
		mpb.WithWidth(width),
	)

	go func(b *mpb.Bar) {
		b.IncrBy(100)
		p.Wait()
	}(p.AddBar(100))

	select {
	case <-shutdown:
		bar := " [" + strings.Repeat("=", width-4) + "] "
		if got, _, found := strings.Cut(buf.String(), "\n"); found {
			if got != bar {
				t.Errorf("want %q, got %q", bar, got)
			}
		} else {
			t.Fatal("Expected buf to contain some ' [=..] \\n'")
		}
	case <-time.After(timeout):
		t.Fatalf("Test timeout %v", timeout)
	}
}

func TestShutdownWithErrFiller(t *testing.T) {
	testError := errors.New("test error")
	shutdown := make(chan any)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithAutoRefresh(),
	)

	bar := p.AddBar(100,
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if st.Current > 11 {
					return testError
				}
				return base.Fill(w, st)
			})
		}),
	)

	go func() {
		for !bar.AbortedOrCompleted() {
			bar.Increment()
			time.Sleep(10 * time.Millisecond)
		}
		p.Wait()
	}()

	select {
	case <-shutdown:
		if !bar.Aborted() {
			t.Error("Expected bar to be aborted")
		}
		if !errors.Is(p.Error, testError) {
			t.Errorf("Expected err: %#v, got %#v", testError, p.Error)
		}
	case <-time.After(timeout):
		t.Fatalf("Test timeout %v", timeout)
	}
}

func TestShutdownAfterBarAbortWithDrop(t *testing.T) {
	var barCount int
	shutdown, depleteHeap := make(chan any), make(chan *mpb.Bar, 1)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithDepleteHeap(depleteHeap),
		mpb.WithAutoRefresh(),
	)

	bar := p.AddBar(100)
	go func() {
		for i := 0; !bar.AbortedOrCompleted(); i++ {
			if i > 11 {
				bar.Abort(true)
			} else {
				bar.Increment()
				time.Sleep(10 * time.Millisecond)
			}
		}
		p.Wait()
	}()

test:
	for {
		select {
		case _, ok := <-depleteHeap:
			if !ok {
				break test
			}
			barCount++
		case <-shutdown:
			shutdown = nil
		case <-time.After(timeout):
			t.Fatalf("Test timeout %v", timeout)
		}
	}
	if barCount != 0 {
		t.Errorf("Expected to receive 0 bars, got: %d", barCount)
	}
	if !bar.Aborted() {
		t.Error("Expected bar to be aborted")
	}
}

func TestShutdownAfterBarAbortWithNoDrop(t *testing.T) {
	var barCount int
	shutdown, depleteHeap := make(chan any), make(chan *mpb.Bar, 1)
	p := mpb.New(
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithDepleteHeap(depleteHeap),
		mpb.WithAutoRefresh(),
	)

	bar := p.AddBar(100)
	go func() {
		for i := 0; !bar.AbortedOrCompleted(); i++ {
			if i > 11 {
				bar.Abort(false)
			} else {
				bar.Increment()
				time.Sleep(10 * time.Millisecond)
			}
		}
		p.Wait()
	}()

test:
	for {
		select {
		case _, ok := <-depleteHeap:
			if !ok {
				break test
			}
			barCount++
		case <-shutdown:
			shutdown = nil
		case <-time.After(timeout):
			t.Fatalf("Test timeout %v", timeout)
		}
	}
	if barCount != 1 {
		t.Errorf("Expected to receive 1 bar, got: %d", barCount)
	}
	if !bar.Aborted() {
		t.Error("Expected bar to be aborted")
	}
}

func TestBarPristinePopOrder(t *testing.T) {
	var received []*mpb.Bar
	shutdown, depleteHeap := make(chan any), make(chan *mpb.Bar, 1)
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithDepleteHeap(depleteHeap),
		mpb.WithAutoRefresh(),
	)
	a := p.AddBar(100, mpb.BarPriority(1), mpb.BarID(1))
	b := p.AddBar(100, mpb.BarPriority(2), mpb.BarID(2))
	c := p.AddBar(100, mpb.BarPriority(3), mpb.BarID(3))
	pristineOrder := []*mpb.Bar{c, b, a}

	cancel()

test:
	for {
		select {
		case b, ok := <-depleteHeap:
			if !ok {
				break test
			}
			received = append(received, b)
		case <-shutdown:
			shutdown = nil
		case <-time.After(timeout):
			t.Fatalf("Test timeout %v", timeout)
		}
	}
	if len(received) != len(pristineOrder) {
		t.Fatalf("Expected to receive %d bars, got: %d", len(pristineOrder), len(received))
	}
	for i, b := range received {
		if bar := pristineOrder[i]; bar.ID() != b.ID() {
			t.Errorf("Expected bar id: %d, got bar id: %d", bar.ID(), b.ID())
		}
	}
}

func TestUpdateBarPriority(t *testing.T) {
	testCases := []struct {
		name    string
		refresh bool
		lazy    bool
	}{
		{"refresh=n,lazy=n", false, false},
		{"refresh=y,lazy=n", true, false},
		{"refresh=n,lazy=y", false, true},
		{"refresh=y,lazy=y", true, true},
	}

	for _, test := range testCases {
		t.Run(test.name, makeUpdateBarPriorityTest(test.refresh, test.lazy))
	}
}

func TestAddAfterDone(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	p.AddBar(100).IncrBy(100)
	p.Wait()

	_, err := p.Add(100, nil)

	if err != mpb.ErrDone {
		t.Errorf("Expected %q, got: %q\n", mpb.ErrDone, err)
	}
}

func makeUpdateBarPriorityTest(refresh, lazy bool) func(*testing.T) {
	return func(t *testing.T) {
		var received []*mpb.Bar
		shutdown, handOverBarCh := make(chan any), make(chan *mpb.Bar, 1)
		refreshCh := make(chan any)
		ctx, cancel := context.WithCancel(context.Background())
		p := mpb.NewWithContext(ctx,
			mpb.WithOutput(io.Discard),
			mpb.WithManualRefresh(refreshCh),
			mpb.WithShutdownNotifier(shutdown),
			mpb.WithDepleteHeap(handOverBarCh),
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

		cancel()

	test:
		for {
			select {
			case b, ok := <-handOverBarCh:
				if !ok {
					break test
				}
				received = append(received, b)
			case <-shutdown:
				shutdown = nil
			case <-time.After(timeout):
				t.Fatalf("Test timeout %v", timeout)
			}
		}
		if len(received) != len(checkOrder) {
			t.Fatalf("Expected to receive %d bars, got: %d", len(checkOrder), len(received))
		}
		for i, b := range received {
			if bar := checkOrder[i]; bar.ID() != b.ID() {
				t.Errorf("Expected bar id: %d, got bar id: %d", bar.ID(), b.ID())
			}
		}
	}
}
