package main

import (
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(80))

	total := 100
	name := "Complex Filler:"
	bs := mpb.BarStyle()
	bs.Lbound("[\u001b[36;1m")
	bs.Filler("_")
	bs.Tip("\u001b[0m⛵\u001b[36;1m")
	bs.Padding("_")
	bs.Rbound("\u001b[0m]")
	bar := p.New(int64(total), bs,
		mpb.PrependDecorators(decor.Name(name)),
		mpb.AppendDecorators(decor.Percentage()),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}
	// wait for our bar to complete
	p.Wait()
}
