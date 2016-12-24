package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

func main() {

	name := "Single bar:"
	p := mpb.New(nil)
	bar := p.AddBar(100).PrependName(name, 0).AppendPercentage()

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()
	fmt.Println("stop")
}
