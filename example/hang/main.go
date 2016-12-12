package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/vbauerster/uiprogress"
)

const (
	totalItem    = 100
	maxBlockSize = 10
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	decor := func(s *uiprogress.Statistics) string {
		str := fmt.Sprintf("%d/%d", s.Completed, s.Total)
		return fmt.Sprintf("%-7s", str)
	}

	p := uiprogress.New()
	bar := p.AddBar(totalItem).AppendETA().PrependFunc(decor)

	blockSize := rand.Intn(maxBlockSize) + 1
	// Fallowing will hang, in order not to hang
	// use !bar.IsCompleted in loop condition
	// for i := 0; !bar.IsCompleted(); i += blockSize {
	for i := 0; i < totalItem; i += blockSize {
		time.Sleep(time.Duration(blockSize) * (50*time.Millisecond + time.Duration(rand.Intn(5*int(time.Millisecond)))))
		bar.Incr(blockSize)
		blockSize = rand.Intn(maxBlockSize) + 1
	}

	p.WaitAndStop()
	fmt.Println("stop")
}
