package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/pkg/profile"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	totalBars = 32
)

var proftype = flag.String("prof", "", "profile type (cpu, mem)")

func main() {
	flag.Parse()
	switch *proftype {
	case "cpu":
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}
	var wg sync.WaitGroup
	// passed wg will be accounted at p.Wait() call
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithDebugOutput(os.Stderr))
	wg.Add(totalBars)

	for i := 0; i < totalBars; i++ {
		name := fmt.Sprintf("Bar#%02d: ", i)
		total := rand.Intn(320) + 10
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name, decor.WCSyncWidthR),
				decor.Elapsed(decor.ET_STYLE_GO, decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.Percentage(decor.WC{W: 5}), "done",
				),
			),
		)

		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for !bar.Completed() {
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}
	// wait for passed wg and for all bars to complete and flush
	p.Wait()
}
