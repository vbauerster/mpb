//+build go1.7

package mpb_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
)

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	p := mpb.New().WithContext(ctx).ShutdownNotify(shutdown)
	numBars := 3

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name), 0)

		go func() {
			for i := 0; i < 10000; i++ {
				bar.Incr(1)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}()
	}

	cancel()

	select {
	case <-shutdown:
	case <-time.After(500 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}
