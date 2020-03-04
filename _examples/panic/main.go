package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithDebugOutput(os.Stderr))

	wantPanic := "Some really long panic panic panic panic panic panic panic, really it is very long"
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("b#%02d:", i)
		bar := p.AddBar(100, mpb.BarID(i), mpb.PrependDecorators(panicDecorator(name, wantPanic)))

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				time.Sleep(50 * time.Millisecond)
				bar.Increment()
			}
		}()
	}

	p.Wait()
}

func panicDecorator(name, panicMsg string) decor.Decorator {
	return decor.Any(func(s *decor.Statistics) string {
		if s.ID == 1 && s.Current >= 42 {
			panic(panicMsg)
		}
		return name
	})
}
