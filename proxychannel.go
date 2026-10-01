package mpb

import "time"

// Sizer is implemented by values passed through [ProxyChannel] that can
// report their size in bytes. When T implements Sizer, the bar is incremented
// by T.Size() bytes per value; values that do not implement Sizer count as 1.
type Sizer interface {
	Size() int64
}

// ProxyChannel wraps ch with bar progress tracking. A goroutine is started that
// reads from ch, increments b, and forwards each value to the returned channel.
//
// Bar updates are batched: the accumulated count is flushed to b only after
// updateInterval has elapsed since the last flush. Values are always forwarded
// to the output channel immediately so the consumer is never held back by the
// update throttle. Any remaining accumulated count is flushed when ch is closed.
//
// T values that implement [Sizer] contribute their byte size to the bar;
// all other values count as 1.
//
// The output channel has the same capacity as ch and is closed when ch is
// closed or b's context is cancelled. Returns nil if b is already done.
func ProxyChannel[T any](b *Bar, ch <-chan T, updateInterval time.Duration) <-chan any {
	result := make(chan bool, 1)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.ewmaDecorators) != 0 }:
		hasEwma := <-result
		out := make(chan any, cap(ch))
		go runProxyChannel(b, ch, out, updateInterval, hasEwma)
		return out
	case <-b.ctx.Done():
		return nil
	}
}

func runProxyChannel[T any](b *Bar, in <-chan T, out chan<- any, updateInterval time.Duration, hasEwma bool) {
	defer close(out)

	var pending int64
	now := time.Now()
	lastFlush := now
	batchStart := now

	flush := func() {
		if pending == 0 {
			return
		}
		if hasEwma {
			b.EwmaIncrInt64(pending, time.Since(batchStart))
		} else {
			b.IncrInt64(pending)
		}
		pending = 0
		now := time.Now()
		lastFlush = now
		batchStart = now
	}

	for {
		select {
		case v, ok := <-in:
			if !ok {
				// ch closed: flush remainder then close out
				flush()
				return
			}
			if s, isSizer := any(v).(Sizer); isSizer {
				pending += s.Size()
			} else {
				pending++
			}
			if time.Since(lastFlush) >= updateInterval {
				flush()
			}
			// fast-pass: forward immediately regardless of bar update
			select {
			case out <- v:
			case <-b.ctx.Done():
				return
			}
		case <-b.ctx.Done():
			return
		}
	}
}
