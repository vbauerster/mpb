package mpb

import (
	"io/ioutil"
	"testing"
)

func BenchmarkIncrSingleBar(b *testing.B) {
	p := New(WithOutput(ioutil.Discard))
	bar := p.AddBar(int64(b.N))
	for i := 0; i < b.N; i++ {
		bar.Increment()
	}
}

func BenchmarkIncrSingleBarStartBlock(b *testing.B) {
	p := New(WithOutput(ioutil.Discard))
	bar := p.AddBar(int64(b.N))
	for i := 0; i < b.N; i++ {
		bar.StartBlock()
		bar.Increment()
	}
}

func BenchmarkIncrSingleBarWhileIsNotCompleted(b *testing.B) {
	p := New(WithOutput(ioutil.Discard))
	bar := p.AddBar(int64(b.N))
	for !bar.Completed() {
		bar.Increment()
	}
}
