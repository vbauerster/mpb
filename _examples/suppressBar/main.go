package main

import (
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
	filler, nextCh := makeCustomFiller(15, msgCh)
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
		for i := 0; i < total; i++ {
			time.Sleep(time.Duration(rng.Intn(10)+1) * max / 10)
			if i == 33 {
				msgCh <- "some error, retrying..."
			}
			bar.Increment()
		}
	}()

	p.Wait()
}

func makeCustomFiller(maxFlash int, ch <-chan string) (mpb.FillerFunc, <-chan struct{}) {
	type refiller interface {
		SetRefill(int64)
	}
	filler := mpb.NewBarFiller()
	nextCh := make(chan struct{}, 1)
	var msg string
	var count int
	return func(w io.Writer, width int, st *decor.Statistics) {
		if count != 0 {
			nextCh <- struct{}{}
			limitFmt := fmt.Sprintf("%%.%ds", width)
			fmt.Fprintf(w, limitFmt, msg)
			count--
			return
		}
		select {
		case msg = <-ch:
			nextCh <- struct{}{}
			if f, ok := filler.(refiller); ok {
				defer f.SetRefill(st.Current)
			}
			count = maxFlash
		default:
		}
		filler.Fill(w, width, st)
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
