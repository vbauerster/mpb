package uiprogress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/uilive"
)

type opType uint

const (
	barAdd opType = iota
	barRemove
)

const refreshRate = 60

// progress represents the container that renders progress bars
type progress struct {
	// out is the writer to render progress bars to
	out io.Writer

	// Width is the width of the progress bars
	// Width int

	op chan *operation

	// new refresh interval to be send over this channel
	interval chan time.Duration

	wg *sync.WaitGroup

	stopped bool
}

type operation struct {
	kind   opType
	bar    *Bar
	result chan bool
}

// New returns a new progress bar with defaults
func New() *progress {
	p := &progress{
		out:      os.Stdout,
		op:       make(chan *operation),
		interval: make(chan time.Duration),
		wg:       new(sync.WaitGroup),
	}
	go p.server()
	return p
}

// SetOut sets underlying writer of progress
// default is os.Stdout
func (p *progress) SetOut(w io.Writer) *progress {
	if w == nil {
		return p
	}
	p.out = w
	return p
}

// AddBar creates a new progress bar and adds to the container
func (p *progress) AddBar(total int) *Bar {
	p.wg.Add(1)
	bar := newBar(total, p.wg)
	// bar.Width = p.Width
	p.op <- &operation{barAdd, bar, nil}
	return bar
}

func (p *progress) RemoveBar(b *Bar) bool {
	result := make(chan bool)
	p.op <- &operation{barRemove, b, result}
	return <-result
}

// RefreshRate overrides default (30ms) refreshRate value
func (p *progress) RefreshRate(d time.Duration) *progress {
	p.interval <- d
	return p
}

// WaitAndStop stops listening
func (p *progress) WaitAndStop() {
	if !p.stopped {
		// fmt.Fprintln(os.Stderr, "p.WaitAndStop")
		p.stopped = true
		p.wg.Wait()
		close(p.op)
	}
}

// server monitors underlying channels and renders any progress bars
func (p *progress) server() {
	t := time.NewTicker(refreshRate * time.Millisecond)
	bars := make([]*Bar, 0, 4)
	lw := uilive.New(p.out)
	for {
		select {
		case op, ok := <-p.op:
			if !ok {
				// fmt.Fprintln(os.Stderr, "Sopping bars")
				for _, b := range bars {
					b.Stop()
				}
				t.Stop()
				return
			}
			switch op.kind {
			case barAdd:
				bars = append(bars, op.bar)
			case barRemove:
				var ok bool
				for i, b := range bars {
					if b == op.bar {
						bars = append(bars[:i], bars[i+1:]...)
						ok = true
						b.Stop()
						break
					}
				}
				op.result <- ok
			}
		case <-t.C:
			for _, b := range bars {
				// cannot parallel this, because order matters
				fmt.Fprintln(lw, b)
			}
			lw.Flush()
			for _, b := range bars {
				go func(b *Bar) {
					b.flushed()
				}(b)
			}
		case d := <-p.interval:
			t.Stop()
			t = time.NewTicker(d)
		}
	}
}
