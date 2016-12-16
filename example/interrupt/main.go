package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	totalItem    = 100
	maxBlockSize = 14
)

func main() {
	decor := func(s *mpb.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Current, s.Total)
		return fmt.Sprintf("%-7s", str)
	}

	p := mpb.New()
	bar := p.AddBar(totalItem).AppendETA().PrependFunc(decor)

	blockSize := rand.Intn(maxBlockSize) + 1
	for i := 0; bar.InProgress(); i++ {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Incr(blockSize)
		if bar.Current() > 42 {
			p.RemoveBar(bar)
		}
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	p.Stop()
	fmt.Println("stop")
}
