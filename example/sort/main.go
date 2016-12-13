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

	p := mpb.New().RefreshRate(80 * time.Millisecond).Sort(mpb.SortTop).SetWidth(60)

	bar1 := p.AddBar(100).AppendETA().PrependFunc(getDecor("Bar#1"))
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(60).AppendETA().PrependFunc(getDecor("Bar#2"))
	go func() {
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 60; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).AppendETA().PrependFunc(getDecor("Bar#3"))
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

func getDecor(name string) mpb.DecoratorFunc {
	return func(s *mpb.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Completed, s.Total)
		return fmt.Sprintf("%s %-7s", name, str)
	}
}
