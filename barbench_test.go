package mpb_test

import (
	"io"
	"testing"

	"github.com/vbauerster/mpb/v8"
)

const total = 1000

func BenchmarkNopStyleB1(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 1)
}

func BenchmarkNopStyleWithAutoRefreshB1(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 1)
}

func BenchmarkNopStylesB2(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 2)
}

func BenchmarkNopStylesWithAutoRefreshB2(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 2)
}

func BenchmarkNopStylesB3(b *testing.B) {
	bench(b, mpb.NopStyle(), false, 3)
}

func BenchmarkNopStylesWithAutoRefreshB3(b *testing.B) {
	bench(b, mpb.NopStyle(), true, 3)
}

func BenchmarkBarStyleB1(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 1)
}

func BenchmarkBarStyleWithAutoRefreshB1(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 1)
}

func BenchmarkBarStylesB2(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 2)
}

func BenchmarkBarStylesWithAutoRefreshB2(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 2)
}

func BenchmarkBarStylesB3(b *testing.B) {
	bench(b, mpb.BarStyle(), false, 3)
}

func BenchmarkBarStylesWithAutoRefreshB3(b *testing.B) {
	bench(b, mpb.BarStyle(), true, 3)
}

func bench(b *testing.B, builder mpb.BarFillerBuilder, autoRefresh bool, n int) {
	p := mpb.New(
		mpb.WithWidth(100),
		mpb.WithOutput(io.Discard),
		mpb.ContainerOptional(mpb.WithAutoRefresh(), autoRefresh),
	)
	defer p.Wait()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var bars []*mpb.Bar
		for j := 0; j < n; j++ {
			bars = append(bars, p.New(total, builder))
			switch j {
			case n - 1:
				complete(bars[j])
			default:
				go complete(bars[j])
			}
		}
		for _, bar := range bars {
			bar.Wait()
		}
	}
}

func complete(bar *mpb.Bar) {
	for i := 0; i < total; i++ {
		bar.Increment()
	}
}
