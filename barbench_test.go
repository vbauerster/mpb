package mpb

import (
	"sync"
	"testing"
)

const total = 1000

func BenchmarkNopStyleOneBar(b *testing.B) {
	bench(b, NopStyle(), 1)
}

func BenchmarkNopStyleTwoBars(b *testing.B) {
	bench(b, NopStyle(), 2)
}

func BenchmarkNopStyleThreeBars(b *testing.B) {
	bench(b, NopStyle(), 3)
}

func BenchmarkBarStyleOneBar(b *testing.B) {
	bench(b, BarStyle(), 1)
}

func BenchmarkBarStyleTwoBars(b *testing.B) {
	bench(b, BarStyle(), 2)
}

func BenchmarkBarStyleThreeBars(b *testing.B) {
	bench(b, BarStyle(), 3)
}

func bench(b *testing.B, builder BarFillerBuilder, n int) {
	var wg sync.WaitGroup
	p := New(WithOutput(nil), WithWidth(80))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			bar := p.New(total, builder)
			switch j {
			case n - 1:
				complete(b, bar)
			default:
				wg.Add(1)
				go func() {
					complete(b, bar)
					wg.Done()
				}()
			}
		}
		wg.Wait()
	}
	p.Wait()
}

func complete(b *testing.B, bar *Bar) {
	for i := 0; i < total; i++ {
		bar.Increment()
	}
	if !bar.Completed() {
		b.Fail()
	}
	bar.Wait()
}
