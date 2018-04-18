package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const (
	totalBars = 32
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(120) + 10
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.StaticName(name, len(name), 0),
				decor.ETA(4, decor.DSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Percentage(5, 0),
			),
		)

		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}

	p.Wait()
	fmt.Println("done")
}
