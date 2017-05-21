//+build go1.7

package main

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p := mpb.New().WithContext(ctx)

	var wg sync.WaitGroup
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBarWithID(i, int64(total)).
			PrependName(name, 0, mpb.DwidthSync|mpb.DidentRight).
			PrependETA(4, mpb.DwidthSync|mpb.DextraSpace).
			AppendPercentage(5, 0)
		go func() {
			defer func() {
				// fmt.Printf("%s done\n", name)
				wg.Done()
			}()
			blockSize := rand.Intn(maxBlockSize) + 1
			for i := 0; i < total; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}
				sleep(blockSize)
				bar.Incr(1)
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
