package uiprogress

import (
	"fmt"
	"time"
)

var (
	// Fill is the default character representing completed progress
	Fill byte = '='

	// Head is the default character that moves when progress is updated
	Head byte = '>'

	// Empty is the default character that represents the empty progress
	Empty byte = '-'

	// LeftEnd is the default character in the left most part of the progress indicator
	LeftEnd byte = '['

	// RightEnd is the default character in the right most part of the progress indicator
	RightEnd byte = ']'

	// Width is the default width of the progress bar
	Width = 70
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(s *Statistics) string

type decoratorFuncType uint

const (
	decoratorAppend decoratorFuncType = iota
	decoratorPrepend
)

type decorator struct {
	kind decoratorFuncType
	f    DecoratorFunc
}

// Bar represents a progress bar
type Bar struct {
	// total of the total  for the progress bar
	total int

	// LeftEnd is character in the left most part of the progress indicator. Defaults to '['
	LeftEnd byte

	// RightEnd is character in the right most part of the progress indicator. Defaults to ']'
	RightEnd byte

	// Fill is the character representing completed progress. Defaults to '='
	Fill byte

	// Head is the character that moves when progress is updated.  Defaults to '>'
	Head byte

	// Empty is the character that represents the empty progress. Default is '-'
	Empty byte

	// Width is the width of the progress bar
	Width int

	Alpha float64

	currentIncrCh chan int

	redrawRequestCh chan *redrawRequest

	decoratorCh chan *decorator

	timePerItemEstimate time.Duration
}

type Statistics struct {
	Total, Completed    int
	TimePerItemEstimate time.Duration
}

type redrawRequest struct {
	bufch chan []byte
}

// NewBar returns a new progress bar
func NewBar(total int) *Bar {
	b := &Bar{
		Alpha:           0.25,
		total:           total,
		Width:           Width,
		LeftEnd:         LeftEnd,
		RightEnd:        RightEnd,
		Head:            Head,
		Fill:            Fill,
		Empty:           Empty,
		currentIncrCh:   make(chan int),
		redrawRequestCh: make(chan *redrawRequest),
		decoratorCh:     make(chan *decorator),
	}
	go b.server()
	return b
}

func (b *Bar) Incr(n int) {
	b.currentIncrCh <- n
}

func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorPrepend, f}
	return b
}

func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorAppend, f}
	return b
}

func (b *Bar) PrependETA() *Bar {
	b.PrependFunc(func(s *Statistics) string {
		eta := time.Duration(s.Total-s.Completed) * s.TimePerItemEstimate
		return fmt.Sprintf("ETA %-5v", time.Duration(eta.Seconds())*time.Second)
	})
	return b
}

func (b *Bar) AppendETA() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		eta := time.Duration(s.Total-s.Completed) * s.TimePerItemEstimate
		return fmt.Sprintf("ETA %v", time.Duration(eta.Seconds())*time.Second)
	})
	return b
}

// String returns the string representation of the bar
func (b *Bar) String() string {
	bufch := make(chan []byte)
	b.redrawRequestCh <- &redrawRequest{bufch}
	return string(<-bufch)
}

func (b *Bar) server() {
	var current int
	blockStartTime := time.Now()
	buf := make([]byte, b.Width, b.Width+24)
	var appendFuncs []DecoratorFunc
	var prependFuncs []DecoratorFunc
	for {
		select {
		case i := <-b.currentIncrCh:
			n := current + i
			if n > b.total {
				return
			}
			b.updateTimePerItemEstimate(i, blockStartTime)
			current = n
			blockStartTime = time.Now()
		case d := <-b.decoratorCh:
			switch d.kind {
			case decoratorAppend:
				appendFuncs = append(appendFuncs, d.f)
			case decoratorPrepend:
				prependFuncs = append(prependFuncs, d.f)
			}
		case r := <-b.redrawRequestCh:
			r.bufch <- b.draw(buf, current, appendFuncs, prependFuncs)
		}
	}
}

func (b *Bar) draw(buf []byte, current int, appendFuncs, prependFuncs []DecoratorFunc) []byte {
	completedWidth := current * b.Width / b.total

	for i := 0; i < completedWidth; i++ {
		buf[i] = b.Fill
	}
	for i := completedWidth; i < b.Width; i++ {
		buf[i] = b.Empty
	}
	// set head bit
	if completedWidth > 0 && completedWidth < b.Width {
		buf[completedWidth-1] = b.Head
	}

	// set left and right ends bits
	buf[0], buf[len(buf)-1] = b.LeftEnd, b.RightEnd

	s := &Statistics{b.total, current, b.timePerItemEstimate}

	// render append functions to the right of the bar
	for _, f := range appendFuncs {
		buf = append(buf, ' ')
		buf = append(buf, []byte(f(s))...)
	}

	// render prepend functions to the left of the bar
	for _, f := range prependFuncs {
		args := []byte(f(s))
		args = append(args, ' ')
		buf = append(args, buf...)
	}
	return buf
}

func (b *Bar) updateTimePerItemEstimate(items int, blockStartTime time.Time) {
	lastBlockTime := time.Since(blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(items)
	b.timePerItemEstimate = time.Duration((b.Alpha * lastItemEstimate) + (1-b.Alpha)*float64(b.timePerItemEstimate))
}
