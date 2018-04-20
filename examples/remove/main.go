package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		var name string
		name = fmt.Sprintf("Bar#%d:", i)

		var bOption mpb.BarOption
		if i == 0 {
			bOption = mpb.BarRemoveOnComplete()
		}

		b := p.AddBar(int64(total), mpb.BarID(i),
			mpb.PrependDecorators(
				decor.StaticName(name, 0, decor.DwidthSync|decor.DidentRight),
				decor.ETA(4, decor.DSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Percentage(5, 0),
			),
			bOption,
		)
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				if b.ID() == 2 && i == 42 {
					p.Abort(b)
					return
				}
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				b.Increment()
			}
		}()
	}

	p.Wait()
	fmt.Println("done")
}
