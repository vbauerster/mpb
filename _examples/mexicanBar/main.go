package main

import (
	"fmt"
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
	bs.LboundMeta(func(s string) string {
		return fmt.Sprint("\033[34m", s, "\033[0m") // blue
	})
	bs.Filler("_").FillerMeta(func(s string) string {
		return fmt.Sprint("\033[36m", s, "\033[0m") // cyan
	})
	bs.Tip("⛵").TipMeta(func(s string) string {
		return fmt.Sprint("\033[31m", s, "\033[0m") // red
	})
	bs.TipOnComplete() // leave tip on complete
	bs.Padding("_").PaddingMeta(func(s string) string {
		return fmt.Sprint("\033[36m", s, "\033[0m") // cyan
	})
	bs.RboundMeta(func(s string) string {
		return fmt.Sprint("\033[34m", s, "\033[0m") // blue
	})
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
