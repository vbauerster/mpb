package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func main() {
	p := mpb.New()

	total := 100
	msgCh := make(chan string)
	resumeCh := make(chan struct{})
	nextCh := make(chan struct{}, 1)
	bar := p.AddBar(int64(total),
		mpb.BarFillerMiddleware(func(base mpb.BarFiller) mpb.BarFiller {
			var msg *string
			return mpb.BarFillerFunc(func(w io.Writer, reqWidth int, st decor.Statistics) {
				select {
				case m := <-msgCh:
					defer func() {
						msg = &m
					}()
					nextCh <- struct{}{}
				case <-resumeCh:
					msg = nil
				default:
				}
				if msg != nil {
					io.WriteString(w, runewidth.Truncate(*msg, st.AvailableWidth, "…"))
					nextCh <- struct{}{}
				} else {
					base.Fill(w, reqWidth, st)
				}
			})
		}),
		mpb.PrependDecorators(
			decor.Name("my bar:"),
		),
		mpb.AppendDecorators(
			newCustomPercentage(nextCh),
		),
	)
	ew := &errorWrapper{}
	time.AfterFunc(2*time.Second, func() {
		ew.reset(errors.New("timeout"))
	})
	// simulating some work
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		max := 100 * time.Millisecond
		for i := 0; i < total; i++ {
			time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
			if ew.isErr() {
				msgCh <- fmt.Sprintf("%s at %d, retrying...", ew.Error(), i)
				go ew.reset(nil)
				i--
				bar.SetRefill(int64(i))
				time.Sleep(3 * time.Second)
				resumeCh <- struct{}{}
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

func newCustomPercentage(nextCh <-chan struct{}) decor.Decorator {
	base := decor.Percentage()
	fn := func(s decor.Statistics) string {
		select {
		case <-nextCh:
			return ""
		default:
			return base.Decor(s)
		}
	}
	return decor.Any(fn)
}
