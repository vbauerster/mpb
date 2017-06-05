//+build go1.7

package mpb_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	p := mpb.New(mpb.WithContext(ctx), mpb.WithShutdownNotifier(shutdown))

	var wg sync.WaitGroup
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total), mpb.BarID(i),
			mpb.PrependDecorators(decor.Name(name, len(name), 0)))

		go func() {
			defer wg.Done()
			for i := 0; i < total; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}

	time.AfterFunc(300*time.Millisecond, cancel)

	wg.Wait()
	p.Stop()

	select {
	case <-shutdown:
	case <-time.After(500 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}
