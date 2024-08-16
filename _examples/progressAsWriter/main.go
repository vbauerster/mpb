package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	total, numBars := 100, 2
	var bwg sync.WaitGroup
	bwg.Add(numBars)
	done := make(chan interface{})
	p := mpb.New(mpb.WithWidth(64), mpb.WithShutdownNotifier(done))

	log.SetOutput(p)
	nopBar := p.MustAdd(0, nil)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name),
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
				),
			),
		)
		// simulating some work
		go func() {
			defer bwg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				// start variable is solely for EWMA calculation
				// EWMA's unit of measure is an iteration's duration
				start := time.Now()
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
				// we need to call EwmaIncrement to fulfill ewma decorator's contract
				bar.EwmaIncrement(time.Since(start))
			}
			log.Println(name, "done")
		}()
	}

	var qwg sync.WaitGroup
	qwg.Add(1)
	go func() {
		defer qwg.Done()
		for {
			select {
			case <-done:
				// after done, underlying io.Writer returns mpb.DoneError
				// so following isn't printed
				log.Println("all done")
				return
			default:
				log.Println("waiting for done")
				time.Sleep(150 * time.Millisecond)
			}
		}
	}()

	bwg.Wait()
	log.Println("completing nop bar")
	nopBar.EnableTriggerComplete()

	p.Wait()
	qwg.Wait()
}
