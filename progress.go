package mpb

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/uilive"
)

type opType uint

const (
	opBarAdd opType = iota
	opBarRemove
)

type SortType uint

const (
	SortNone SortType = iota
	SortTop
	SortBottom
)

const refreshRate = 60

// Progress represents the container that renders Progress bars
type Progress struct {
	out     io.Writer
	width   int
	sort    SortType
	stopped bool

	op            chan *operation
	rrChangeReqCh chan time.Duration

	wg *sync.WaitGroup
}

type operation struct {
	kind   opType
	bar    *Bar
	result chan bool
}

// New returns a new progress bar with defaults
func New() *Progress {
	p := &Progress{
		out:           os.Stdout,
		width:         70,
		op:            make(chan *operation),
		rrChangeReqCh: make(chan time.Duration),
		wg:            new(sync.WaitGroup),
	}
	go p.server()
	return p
}

func (p *Progress) SetWidth(n int) *Progress {
	if n <= 0 {
		return p
	}
	p.width = n
	return p
}

// SetOut sets underlying writer of progress
// default is os.Stdout
func (p *Progress) SetOut(w io.Writer) *Progress {
	if w == nil {
		return p
	}
	p.out = w
	return p
}

// RefreshRate overrides default (30ms) refreshRate value
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	p.rrChangeReqCh <- d
	return p
}

func (p *Progress) WithSort(sort SortType) *Progress {
	p.sort = sort
	return p
}

// AddBar creates a new progress bar and adds to the container
func (p *Progress) AddBar(total int) *Bar {
	p.wg.Add(1)
	bar := newBar(total, p.width, p.wg)
	p.op <- &operation{opBarAdd, bar, nil}
	return bar
}

func (p *Progress) RemoveBar(b *Bar) bool {
	result := make(chan bool)
	p.op <- &operation{opBarRemove, b, result}
	return <-result
}

// WaitAndStop stops listening
func (p *Progress) WaitAndStop() {
	if !p.stopped {
		// fmt.Fprintln(os.Stderr, "p.WaitAndStop")
		p.stopped = true
		p.wg.Wait()
		close(p.op)
	}
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server() {
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
			case opBarAdd:
				bars = append(bars, op.bar)
			case opBarRemove:
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
			switch p.sort {
			case SortTop:
				sort.Sort(sort.Reverse(SortableBarSlice(bars)))
			case SortBottom:
				sort.Sort(SortableBarSlice(bars))
			}
			for _, b := range bars {
				// cannot parallel this, because order matters
				fmt.Fprintln(lw, b)
			}
			lw.Flush()
			for _, b := range bars {
				go func(b *Bar) {
					b.flushedCh <- struct{}{}
				}(b)
			}
		case d := <-p.rrChangeReqCh:
			t.Stop()
			t = time.NewTicker(d)
		}
	}
}
