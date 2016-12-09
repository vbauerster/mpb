package uiprogress

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gosuri/uilive"
)

type opType uint

const (
	add opType = iota
	remove
)

const refreshRate = 100

// progress represents the container that renders progress bars
type progress struct {
	// out is the writer to render progress bars to
	out io.Writer

	// Width is the width of the progress bars
	// Width int

	lw *uilive.Writer

	op chan *operation

	// new refresh interval to be send over this channel
	interval chan time.Duration
}

type operation struct {
	kind opType
	bar  *Bar
	ok   chan bool
}

// New returns a new progress bar with defaults
func New() *progress {
	p := &progress{
		out:      os.Stdout,
		lw:       uilive.New(),
		op:       make(chan *operation),
		interval: make(chan time.Duration),
	}
	go p.server()
	return p
}

// RefreshRate overrides default (30ms) refreshRate value
func (p *progress) RefreshRate(d time.Duration) *progress {
	p.interval <- d
	return p
}

// SetOut sets underlying writer of progress
// default is os.Stdout
func (p *progress) SetOut(w io.Writer) *progress {
	p.out = w
	return p
}

// AddBar creates a new progress bar and adds to the container
func (p *progress) AddBar(total int) *Bar {
	bar := NewBar(total)
	// bar.Width = p.Width
	p.op <- &operation{add, bar, nil}
	return bar
}

func (p *progress) RemoveBar(b *Bar) bool {
	result := make(chan bool)
	p.op <- &operation{remove, b, result}
	return <-result
}

// Bypass returns a writer which allows non-buffered data to be written to the underlying output
func (p *progress) Bypass() io.Writer {
	return p.lw.Bypass()
}

// Stop stops listening
func (p *progress) Stop() {
	close(p.op)
}

// server monitors underlying channels and renders any progress bars
func (p *progress) server() {
	t := time.NewTicker(refreshRate * time.Millisecond)
	bars := make([]*Bar, 0)
	p.lw.Out = p.out
	for {
		select {
		case op, ok := <-p.op:
			if !ok {
				t.Stop()
				close(p.interval)
				return
			}
			switch op.kind {
			case add:
				bars = append(bars, op.bar)
			case remove:
				var ok bool
				for i, b := range bars {
					if b == op.bar {
						bars = append(bars[:i], bars[i+1:]...)
						ok = true
						break
					}
				}
				op.ok <- ok
			}
		case <-t.C:
			for _, b := range bars {
				fmt.Fprintln(p.lw, b.String())
			}
			p.lw.Flush()
		case d := <-p.interval:
			t.Stop()
			t = time.NewTicker(d)
		}
	}
}
