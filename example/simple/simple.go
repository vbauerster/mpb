package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

const (
	totalItem    = 100
	maxBlockSize = 10
)

func main() {
	decor := func(s *mpb.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Current, s.Total)
		return fmt.Sprintf("%-7s", str)
	}

	p := mpb.New()
	bar := p.AddBar(totalItem).AppendETA().PrependFunc(decor)
	// if you omit the following line, bar rendering goroutine may not have a
	// chance to coplete, thus better always use.
	p.Wg.Add(1)

	blockSize := rand.Intn(maxBlockSize) + 1
	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Incr(1)
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	p.WaitAndStop()
	fmt.Println("stop")
}
