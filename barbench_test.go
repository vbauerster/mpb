package mpb

import (
	"sync"
	"testing"
)

const total = 1000

func BenchmarkIncrementOneBar(b *testing.B) {
	benchBody(1, b)
}

func BenchmarkIncrementTwoBars(b *testing.B) {
	benchBody(2, b)
}

func BenchmarkIncrementThreeBars(b *testing.B) {
	benchBody(3, b)
}

func BenchmarkIncrementFourBars(b *testing.B) {
	benchBody(4, b)
}

func benchBody(n int, b *testing.B) {
	p := New(WithOutput(nil), WithWidth(80))
	wg := new(sync.WaitGroup)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			switch j {
			case n - 1:
				bar := p.AddBar(total)
				for c := 0; c < total; c++ {
					bar.Increment()
				}
				if !bar.Completed() {
					b.Fail()
				}
			default:
				wg.Add(1)
				go func() {
					bar := p.AddBar(total)
					for c := 0; c < total; c++ {
						bar.Increment()
					}
					if !bar.Completed() {
						b.Fail()
					}
					wg.Done()
				}()
			}
		}
		wg.Wait()
	}
	p.Wait()
}
