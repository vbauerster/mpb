package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
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
			newCustomPercentage(decor.Percentage(), nextCh),
		),
	)
	// simulating some work
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		max := 100 * time.Millisecond
		var err error
		for i := 0; i < total; i++ {
			if err != nil {
				bar.SetRefill(int64(i))
				time.Sleep(3 * time.Second)
				resumeCh <- struct{}{}
				err = nil
			} else {
				time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
			}
			if i == 33 {
				err = errors.New("some error")
				msgCh <- fmt.Sprintf("%s retrying...", err.Error())
			}
			bar.Increment()
		}
	}()

	p.Wait()
}

type customFiller struct {
	mpb.Filler
	base mpb.Filler
}

// implementing mpb.BaseFiller, so bar.SetRefill works
func (cf *customFiller) BaseFiller() mpb.Filler {
	return cf.base
}

func newCustomFiller(ch <-chan string, resume <-chan struct{}) (mpb.Filler, <-chan struct{}) {
	base := mpb.NewBarFiller(mpb.DefaultBarStyle, false)
	nextCh := make(chan struct{}, 1)
	var msg *string
	filler := mpb.FillerFunc(func(w io.Writer, width int, st *decor.Statistics) {
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
	cf := &customFiller{
		Filler: filler,
		base:   base,
	}
	return cf, nextCh
}

type myPercentageDecorator struct {
	decor.Decorator
	ch <-chan struct{}
}

func (d *myPercentageDecorator) Decor(st *decor.Statistics) string {
	select {
	case <-d.ch:
		return ""
	default:
		return d.Decorator.Decor(st)
	}
}

func newCustomPercentage(base decor.Decorator, ch <-chan struct{}) decor.Decorator {
	return &myPercentageDecorator{
		Decorator: base,
		ch:        ch,
	}
}
