package mpb_test

import (
	"io"
	"sync"
	"testing"

	"github.com/vbauerster/mpb/v8"
)

const total = 1000

func BenchmarkNopStyle1Bar(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 1)
}

func BenchmarkNopStyle1BarWithAutoRefresh(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 1)
}

func BenchmarkNopStyle2Bars(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 2)
}

func BenchmarkNopStyle2BarsWithAutoRefresh(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 2)
}

func BenchmarkNopStyle3Bars(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 3)
}

func BenchmarkNopStyle3BarsWithAutoRefresh(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 3)
}

func BenchmarkBarStyle1Bar(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 1)
}

func BenchmarkBarStyle1BarWithAutoRefresh(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 1)
}

func BenchmarkBarStyle2Bars(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 2)
}

func BenchmarkBarStyle2BarsWithAutoRefresh(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 2)
}

func BenchmarkBarStyle3Bars(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 3)
}

func BenchmarkBarStyle3BarsWithAutoRefresh(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 3)
}

func bench(b *testing.B, builder mpb.BarFillerBuilder, autoRefresh bool, n int) {
	var wg sync.WaitGroup
	p := mpb.New(
		mpb.WithWidth(100),
		mpb.WithOutput(io.Discard),
		mpb.ContainerOptional(mpb.WithAutoRefresh(), autoRefresh),
	)
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

func complete(b *testing.B, bar *mpb.Bar) {
	for i := 0; i < total; i++ {
		bar.Increment()
	}
	bar.Wait()
}
