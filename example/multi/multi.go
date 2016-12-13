package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	maxBlockSize = 12
)

func main() {
	decor := func(s *mpb.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Completed, s.Total)
		return fmt.Sprintf("%-7s", str)
	}

	p := mpb.New().RefreshRate(80 * time.Millisecond)

	bar1 := p.AddBar(50).AppendETA().PrependFunc(decor)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 50; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(100).AppendETA().PrependFunc(decor)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).AppendETA().PrependFunc(decor)
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	// time.Sleep(time.Second)
	// p.RemoveBar(bar2)

	p.WaitAndStop()
	bar2.Incr(2)
	fmt.Println("stop")
	// p.AddBar(1) // panic: send on closed channnel
}
