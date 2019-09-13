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
	filler, nextCh := makeCustomFiller(msgCh, resumeCh)
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

func makeCustomFiller(ch <-chan string, resume <-chan struct{}) (mpb.FillerFunc, <-chan struct{}) {
	type refiller interface {
		SetRefill(int64)
	}
	filler := mpb.NewBarFiller()
	nextCh := make(chan struct{}, 1)
	var msg *string
	return func(w io.Writer, width int, st *decor.Statistics) {
		select {
		case m := <-ch:
			if f, ok := filler.(refiller); ok {
				defer f.SetRefill(st.Current)
			}
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
			filler.Fill(w, width, st)
		}
	}, nextCh
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
