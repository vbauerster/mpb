package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/uiprogress"
)

const (
	totalItem    = 1000
	maxBlockSize = 20
)

func main() {
	p := uiprogress.New()
	bar := p.AddBar(totalItem) // Add a new bar

	// optionally, append and prepend completion and elapsed time
	// bar.AppendCompleted()
	// bar.PrependElapsed()

	blockSize := rand.Intn(maxBlockSize) + 1
	for i := 0; i < totalItem; i += blockSize {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Update(i)
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	fmt.Println("stop")

	p.Stop()
}
