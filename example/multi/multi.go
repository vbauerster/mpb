package main

import (
	"sync"
	"time"

	"github.com/vbauerster/uiprogress"
)

func main() {
	waitTime := time.Millisecond * 100
	// p := uiprogress.New().RefreshInterval(100 * time.Millisecond)
	p := uiprogress.New()

	var wg sync.WaitGroup
	bar1 := p.AddBar(20).AppendCompleted().PrependElapsed()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bar1.Incr() {
			time.Sleep(waitTime)
		}
	}()

	bar2 := p.AddBar(40).AppendCompleted().PrependElapsed()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bar2.Incr() {
			time.Sleep(waitTime)
		}
	}()

	time.Sleep(time.Second)
	bar3 := p.AddBar(80).PrependElapsed().AppendCompleted()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bar3.Incr() {
			time.Sleep(waitTime)
		}
	}()

	wg.Wait()
	p.Stop()
	// p.AddBar(1) // panic: send on closed channnel
}
