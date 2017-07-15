package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const (
	maxBlockSize = 12
)

type barSlice []*mpb.Bar

func (bs barSlice) Len() int { return len(bs) }

func (bs barSlice) Less(i, j int) bool {
	ip := decor.CalcPercentage(bs[i].Total(), bs[i].Current(), 100)
	jp := decor.CalcPercentage(bs[j].Total(), bs[j].Current(), 100)
	return ip < jp
}

func (bs barSlice) Swap(i, j int) { bs[i], bs[j] = bs[j], bs[i] }

func sortByProgressFunc() mpb.BeforeRender {
	return func(bars []*mpb.Bar) {
		sort.Sort(sort.Reverse(barSlice(bars)))
	}
}

func main() {

	p := mpb.New(mpb.WithWidth(64), mpb.WithBeforeRenderFunc(sortByProgressFunc()))

	total := 100
	numBars := 3
	var wg sync.WaitGroup
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		var name string
		if i != 1 {
			name = fmt.Sprintf("Bar#%d:", i)
		}
		b := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.StaticName(name, 0, decor.DwidthSync),
				decor.Counters("%3s/%3s", 0, 10, decor.DSyncSpace),
			),
			mpb.AppendDecorators(
				decor.ETA(3, 0),
			),
		)
		go func() {
			defer wg.Done()
			blockSize := rand.Intn(maxBlockSize) + 1
			for i := 0; i < total; i++ {
				sleep(blockSize)
				b.Incr(1)
				blockSize = rand.Intn(maxBlockSize) + 1
			}
		}()
	}

	wg.Wait()
	p.Stop()
	fmt.Println("stop")
}

func sleep(blockSize int) {
	time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
}
