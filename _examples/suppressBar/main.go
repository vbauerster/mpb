package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	p := mpb.New()

	total := 100
	msgCh := make(chan string)
	resumeCh := make(chan struct{})
	nextCh := make(chan struct{}, 1)
	ew := &errorWrapper{}
	timer := time.AfterFunc(2*time.Second, func() {
		ew.reset(errors.New("timeout"))
	})
	defer timer.Stop()
	bar := p.AddBar(int64(total),
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			var msg *string
			var times int
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if msg == nil {
					select {
					case m := <-msgCh:
						msg = &m
						times = 10
						nextCh <- struct{}{}
					default:
					}
					return base.Fill(w, st)
				}
				switch {
				case times == 0, st.Completed, st.Aborted:
					defer func() {
						msg = nil
					}()
					resumeCh <- struct{}{}
				default:
					times--
				}
				_, err := io.WriteString(w, runewidth.Truncate(*msg, st.AvailableWidth, "…"))
				nextCh <- struct{}{}
				return err
			})
		}),
		mpb.PrependDecorators(decor.Name("my bar:")),
		mpb.AppendDecorators(newCustomPercentage(nextCh)),
	)
	// simulating some work
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		max := 100 * time.Millisecond
		for i := 0; i < total; i++ {
			time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
			if ew.isErr() {
				msgCh <- fmt.Sprintf("%s at %d, retrying...", ew.Error(), i)
				i--
				bar.SetRefill(int64(i))
				ew.reset(nil)
				<-resumeCh
				continue
			}
			bar.Increment()
		}
	}()

	p.Wait()
}

type errorWrapper struct {
	sync.RWMutex
	err error
}

func (ew *errorWrapper) Error() string {
	ew.RLock()
	defer ew.RUnlock()
	return ew.err.Error()
}

func (ew *errorWrapper) isErr() bool {
	ew.RLock()
	defer ew.RUnlock()
	return ew.err != nil
}

func (ew *errorWrapper) reset(err error) {
	ew.Lock()
	ew.err = err
	ew.Unlock()
}

type percentage struct {
	decor.Decorator
	suspend <-chan struct{}
}

func (d percentage) Decor(s decor.Statistics) (string, int) {
	select {
	case <-d.suspend:
		return d.Format("")
	default:
		return d.Decorator.Decor(s)
	}
}

func newCustomPercentage(nextCh <-chan struct{}) decor.Decorator {
	return percentage{
		Decorator: decor.Percentage(),
		suspend:   nextCh,
	}
}
