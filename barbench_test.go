package mpb

import (
	"io/ioutil"
	"testing"

	"github.com/vbauerster/mpb/decor"
)

func BenchmarkIncrSingleBar(b *testing.B) {
	p := New(WithOutput(ioutil.Discard))
	bar := p.AddBar(int64(b.N))
	for i := 0; i < b.N; i++ {
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

func BenchmarkIncrSingleBarWithNameDecorator(b *testing.B) {
	p := New(WithOutput(ioutil.Discard))
	bar := p.AddBar(int64(b.N), PrependDecorators(decor.Name("test")))
	for i := 0; i < b.N; i++ {
		bar.Increment()
	}
}
