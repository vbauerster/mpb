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
	bar := p.AddBar(totalItem).PrependFunc(decor).AppendETA(-6)

	blockSize := rand.Intn(maxBlockSize) + 1
	// Fallowing will hang, to prevent
	// use bar.InProgress() bool method
	// for i := 0; bar.InProgress(); i += blockSize {
	for i := 0; i < totalItem; i += blockSize {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Incr(blockSize)
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	p.Stop()
	fmt.Println("stop")
}
