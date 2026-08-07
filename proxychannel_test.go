package mpb_test

import (
	"io"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// countItem is a plain value with no size information; each counts as 1.
type countItem struct{ id int }

// sizedItem implements mpb.Sizer.
type sizedItem struct{ n int64 }

func (s sizedItem) Size() int64 { return s.n }

func sendAndClose[T any](items []T) <-chan T {
	ch := make(chan T, len(items))
	for _, v := range items {
		ch <- v
	}
	close(ch)
	return ch
}

func drainAll(ch <-chan any) int {
	var n int
	for range ch {
		n++
	}
	return n
}

// TestProxyChannel verifies that non-Sizer values each count as 1.
func TestProxyChannel(t *testing.T) {
	const total = 5
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(total), mpb.NopStyle())

	items := make([]countItem, total)
	for i := range items {
		items[i] = countItem{i}
	}

	out := mpb.ProxyChannel(bar, sendAndClose(items), 0)
	n := drainAll(out)
	p.Wait()

	if n != total {
		t.Errorf("received %d values, want %d", n, total)
	}
	if got := bar.Current(); got != int64(total) {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestProxyChannelSizer verifies that Sizer values contribute their byte size.
func TestProxyChannelSizer(t *testing.T) {
	items := []sizedItem{{5}, {6}, {11}}
	var total int64
	for _, item := range items {
		total += item.Size()
	}

	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(total, mpb.NopStyle())

	out := mpb.ProxyChannel(bar, sendAndClose(items), 0)
	drainAll(out)
	p.Wait()

	if got := bar.Current(); got != total {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestEwmaProxyChannel verifies EWMA decorator path increments correctly.
func TestEwmaProxyChannel(t *testing.T) {
	const total = 4
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(total),
		mpb.NopStyle(),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 30)),
	)

	items := make([]countItem, total)
	out := mpb.ProxyChannel(bar, sendAndClose(items), 0)
	drainAll(out)
	p.Wait()

	if got := bar.Current(); got != int64(total) {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestEwmaProxyChannelSizer verifies EWMA + Sizer path increments by byte size.
func TestEwmaProxyChannelSizer(t *testing.T) {
	items := []sizedItem{{3}, {7}}
	var total int64
	for _, item := range items {
		total += item.Size()
	}

	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(total,
		mpb.NopStyle(),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 30)),
	)

	out := mpb.ProxyChannel(bar, sendAndClose(items), 0)
	drainAll(out)
	p.Wait()

	if got := bar.Current(); got != total {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestProxyChannelOrdering verifies that values arrive in the same order they
// were sent.
func TestProxyChannelOrdering(t *testing.T) {
	const total = 10
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(total), mpb.NopStyle())

	items := make([]countItem, total)
	for i := range items {
		items[i] = countItem{i}
	}

	out := mpb.ProxyChannel(bar, sendAndClose(items), 0)
	var idx int
	for v := range out {
		item := v.(countItem)
		if item.id != idx {
			t.Errorf("index %d: got id %d, want %d", idx, item.id, idx)
		}
		idx++
	}
	p.Wait()

	if idx != total {
		t.Errorf("received %d values, want %d", idx, total)
	}
}

// TestProxyChannelUpdateInterval verifies that a large updateInterval still
// produces the correct final count after ch is closed (the end-of-stream flush).
func TestProxyChannelUpdateInterval(t *testing.T) {
	const total = 6
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(total), mpb.NopStyle())

	items := make([]countItem, total)
	// huge interval so every value is a fast-pass — only the close flush counts
	out := mpb.ProxyChannel(bar, sendAndClose(items), time.Hour)
	drainAll(out)
	p.Wait()

	if got := bar.Current(); got != int64(total) {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestProxyChannelUpdateIntervalSizer same as above but with Sizer values.
func TestProxyChannelUpdateIntervalSizer(t *testing.T) {
	items := []sizedItem{{4}, {8}, {16}}
	var total int64
	for _, item := range items {
		total += item.Size()
	}

	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(total, mpb.NopStyle())

	out := mpb.ProxyChannel(bar, sendAndClose(items), time.Hour)
	drainAll(out)
	p.Wait()

	if got := bar.Current(); got != total {
		t.Errorf("bar.Current() = %d, want %d", got, total)
	}
}

// TestProxyChannelOutputClosed verifies that the output channel is closed when
// the input channel is closed.
func TestProxyChannelOutputClosed(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(3, mpb.NopStyle())

	in := make(chan countItem, 3)
	in <- countItem{0}
	in <- countItem{1}
	in <- countItem{2}
	close(in)

	out := mpb.ProxyChannel(bar, in, 0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		drainAll(out)
	}()

	select {
	case <-done:
		// output channel was closed as expected
	case <-time.After(timeout):
		t.Fatal("output channel was not closed within timeout")
	}
	p.Wait()
}
