package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	maxBlockSize = 12
)

func main() {

	var wg sync.WaitGroup
	p := mpb.New().SetWidth(60).WithSort(mpb.SortTop)

	name1 := "Bar#1: "
	bar1 := p.AddBar(100).
		PrependName(name1, len(name1)).PrependFunc(getDecor()).
		AppendETA()
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(60).
		PrependName("", 0-len(name1)).PrependFunc(getDecor()).
		AppendETA()
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 60; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).
		PrependName("Bar#3: ", 0).PrependFunc(getDecor()).
		AppendETA()
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	wg.Wait()
	p.Stop()
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("stop")
}

func getDecor() mpb.DecoratorFunc {
	return func(s *mpb.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Current, s.Total)
		return fmt.Sprintf("%-7s", str)
	}
}
