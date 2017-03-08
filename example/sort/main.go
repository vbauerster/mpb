package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	maxBlockSize = 12
)

type barSlice []*mpb.Bar

func (bs barSlice) Len() int { return len(bs) }

func (bs barSlice) Less(i, j int) bool {
	is := bs[i].GetStatistics()
	js := bs[j].GetStatistics()
	ip := percentage(is.Total, is.Current, 100)
	jp := percentage(js.Total, js.Current, 100)
	return ip < jp
}

func (bs barSlice) Swap(i, j int) { bs[i], bs[j] = bs[j], bs[i] }

func sortByProgressFunc() mpb.BeforeRender {
	return func(bars []*mpb.Bar) {
		sort.Sort(sort.Reverse(barSlice(bars)))
	}
}

func percentage(total, current int64, ratio int) int {
	if total <= 0 {
		return 0
	}
	return int(float64(ratio) * float64(current) / float64(total))
}

func main() {

	var wg sync.WaitGroup
	p := mpb.New().SetWidth(60).BeforeRenderFunc(sortByProgressFunc())

	name1 := "Bar#1:"
	bar1 := p.AddBar(100).
		PrependName(name1, 0, mpb.DwidthSync).
		PrependCounters("%3s/%3s", 0, 10, mpb.DwidthSync|mpb.DextraSpace).
		AppendETA(3, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 100; i++ {
			sleep(blockSize)
			bar1.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar2 := p.AddBar(60).
		PrependName("", 0, mpb.DwidthSync).
		PrependCounters("%3s/%3s", 0, 10, mpb.DwidthSync|mpb.DextraSpace).
		AppendETA(3, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 60; i++ {
			sleep(blockSize)
			bar2.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	bar3 := p.AddBar(80).
		PrependName("Bar#3:", 0, mpb.DwidthSync).
		PrependCounters("%3s/%3s", 0, 10, mpb.DwidthSync|mpb.DextraSpace).
		AppendETA(3, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockSize := rand.Intn(maxBlockSize) + 1
		for i := 0; i < 80; i++ {
			sleep(blockSize)
			bar3.Incr(1)
			blockSize = rand.Intn(maxBlockSize) + 1
		}
	}()

	wg.Wait()
	p.Stop()
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
