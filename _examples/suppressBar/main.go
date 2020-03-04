package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

func main() {
	p := mpb.New()

	total := 100
	msgCh := make(chan string)
	resumeCh := make(chan struct{})
	filler, nextCh := newCustomFiller(msgCh, resumeCh)
	bar := p.Add(int64(total), filler,
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

type myBarFiller struct {
	mpb.BarFiller
	base mpb.BarFiller
}

func (cf *myBarFiller) Base() mpb.BarFiller {
	return cf.base
}

func newCustomFiller(ch <-chan string, resume <-chan struct{}) (mpb.BarFiller, <-chan struct{}) {
	base := mpb.NewBarFiller(mpb.DefaultBarStyle, false)
	nextCh := make(chan struct{}, 1)
	var msg *string
	filler := mpb.BarFillerFunc(func(w io.Writer, width int, st *decor.Statistics) {
		select {
		case m := <-ch:
			defer func() {
				msg = &m
			}()
			nextCh <- struct{}{}
		case <-resume:
			msg = nil
		default:
		}
		if msg != nil {
			limitFmt := fmt.Sprintf("%%.%ds", width)
			fmt.Fprintf(w, limitFmt, *msg)
			nextCh <- struct{}{}
		} else {
			base.Fill(w, width, st)
		}
	})
	cf := &myBarFiller{
		BarFiller: filler,
		base:      base,
	}
	return cf, nextCh
}

func newCustomPercentage(ch <-chan struct{}) decor.Decorator {
	base := decor.Percentage()
	f := func(s *decor.Statistics) string {
		select {
		case <-ch:
			return ""
		default:
			return base.Decor(s)
		}
	}
	return decor.Any(f)
}
