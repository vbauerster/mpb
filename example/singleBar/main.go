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

	name := "Single:"
	p := mpb.New()
	bar := p.AddBar(totalItem).
		PrependName(name, 0).
		AppendPercentage().
		TrimRightSpace()

	blockSize := rand.Intn(maxBlockSize) + 1
	for i := 0; i < totalItem; i++ {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Incr(1)
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	p.Stop()
	fmt.Println("stop")
}
