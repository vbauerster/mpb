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
		bar := p.AddBarWithID(i, int64(total)).PrependName(name, len(name), 0)

		go func() {
			defer func() {
				// fmt.Printf("%s done\n", name)
				wg.Done()
			}()
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

func TestWithNilContext(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			if msg, ok := p.(string); ok && msg == "nil context" {
				return
			}
			t.Errorf("Expected nil context panic, got: %+v", p)
		}
	}()
	_ = mpb.New().WithContext(nil)
}
