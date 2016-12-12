package uiprogress

import (
	"fmt"
	"os"
	"sync"
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

	incrRequestCh chan *incrRequest

	redrawRequestCh chan *redrawRequest

	decoratorCh chan *decorator

	timePerItemEstimate time.Duration

	flushedCh chan struct{}

	stopCh chan struct{}
	done   chan struct{}

	// wg *sync.WaitGroup
}

type Statistics struct {
	Total, Completed    int
	TimePerItemEstimate time.Duration
}

type redrawRequest struct {
	bufCh chan []byte
}

type incrRequest struct {
	amount int
	result chan bool
}

// NewBar returns a new progress bar
func newBar(total int, wg *sync.WaitGroup) *Bar {
	b := &Bar{
		Alpha:           0.25,
		total:           total,
		Width:           Width,
		LeftEnd:         LeftEnd,
		RightEnd:        RightEnd,
		Head:            Head,
		Fill:            Fill,
		Empty:           Empty,
		incrRequestCh:   make(chan *incrRequest),
		redrawRequestCh: make(chan *redrawRequest),
		decoratorCh:     make(chan *decorator),
		flushedCh:       make(chan struct{}),
		stopCh:          make(chan struct{}),
		done:            make(chan struct{}),
	}
	go b.server(wg)
	return b
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

func (b *Bar) PrependPercentage() *Bar {
	b.PrependFunc(func(s *Statistics) string {
		completed := int(100 * float64(s.Completed) / float64(s.Total))
		return fmt.Sprintf("%3d %%", completed)
	})
	return b
}

// String returns the string representation of the bar
func (b *Bar) String() string {
	if b.IsStopped() {
		return "bar stopped"
	}
	bufCh := make(chan []byte)
	b.redrawRequestCh <- &redrawRequest{bufCh}
	return string(<-bufCh)
}

func (b *Bar) flushed() {
	if !b.IsStopped() {
		b.flushedCh <- struct{}{}
	}
}

func (b *Bar) Incr(n int) bool {
	if b.IsStopped() {
		return false
	}
	result := make(chan bool)
	b.incrRequestCh <- &incrRequest{n, result}
	return <-result
}

func (b *Bar) server(wg *sync.WaitGroup) {
	var completed int
	blockStartTime := time.Now()
	buf := make([]byte, b.Width, b.Width+24)
	var appendFuncs []DecoratorFunc
	var prependFuncs []DecoratorFunc
	var done bool
	for {
		select {
		case r := <-b.incrRequestCh:
			n := completed + r.amount
			fmt.Fprintf(os.Stderr, "n = %+v\n", n)
			if n > b.total {
				r.result <- false
				completed = b.total
				fmt.Fprintln(os.Stderr, "n > b.total = return false")
				break // breaks out of select
			}
			// r.result <- true
			b.updateTimePerItemEstimate(r.amount, blockStartTime)
			completed = n
			blockStartTime = time.Now()
			if completed == b.total && !done {
				fmt.Fprintln(os.Stderr, "completed == b.total")
				done = true
				wg.Done()
			}
			r.result <- true
		case d := <-b.decoratorCh:
			switch d.kind {
			case decoratorAppend:
				appendFuncs = append(appendFuncs, d.f)
			case decoratorPrepend:
				prependFuncs = append(prependFuncs, d.f)
			}
		case r := <-b.redrawRequestCh:
			// fmt.Fprintln(os.Stderr, "redraw")
			r.bufCh <- b.draw(buf, completed, appendFuncs, prependFuncs)
		// case <-b.flushedCh:
		// 	if completed == b.total && !done {
		// 		fmt.Fprintln(os.Stderr, "wg.Done")
		// 		done = true
		// 		wg.Done()
		// 	} else {
		// 		fmt.Fprintln(os.Stderr, "wg.Done not done")
		// 	}
		case <-b.stopCh:
			fmt.Fprintln(os.Stderr, "received stop signal")
			// close(b.incrRequestCh)
			close(b.done)
			if !done {
				done = true
				wg.Done()
			}
			// close(b.redrawRequestCh)
			// close(b.flushedCh)
			return
		}
	}
}

func (b *Bar) Stop() {
	if !b.IsStopped() {
		fmt.Fprintln(os.Stderr, "sending to stopCh")
		b.stopCh <- struct{}{}
	} else {
		fmt.Fprintln(os.Stderr, "Stop: already stopped")
	}
}

func (b *Bar) IsStopped() bool {
	select {
	case <-b.done:
		return true
	default:
		return false
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
