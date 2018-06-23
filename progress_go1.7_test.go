//+build go1.7

package mpb_test

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
)

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	p := mpb.New(
		mpb.WithOutput(ioutil.Discard),
		mpb.WithContext(ctx),
		mpb.WithShutdownNotifier(shutdown),
	)

	total := 1000
	numBars := 3
	bars := make([]*mpb.Bar, 0, numBars)
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(total))
		bars = append(bars, bar)
		go func() {
			for !bar.Completed() {
				time.Sleep(randomDuration(40 * time.Millisecond))
				bar.Increment()
			}
		}()
	}

	time.AfterFunc(100*time.Millisecond, cancel)

	p.Wait()
	for _, bar := range bars {
		if bar.Current() >= int64(total) {
			t.Errorf("bar %d: total = %d, current = %d\n", bar.ID(), total, bar.Current())
		}
	}
	select {
	case <-shutdown:
	case <-time.After(100 * time.Millisecond):
		t.Error("Progress didn't stop")
	}
}
