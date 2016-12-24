package mpb

import (
	"context"
	"errors"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/mpb/cwriter"
)

// ErrCallAfterStop thrown by panic, if Progress methods like AddBar() are called
// after Stop() has been called
var ErrCallAfterStop = errors.New("method call on stopped Progress instance")

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

// default RefreshRate
const rr = 100

// Progress represents the container that renders Progress bars
type Progress struct {
	// Context for canceling bars rendering
	ctx context.Context
	// WaitGroup for internal rendering sync
	wg *sync.WaitGroup

	out   io.Writer
	width int
	sort  SortType

	operationCh    chan *operation
	rrChangeReqCh  chan time.Duration
	outChangeReqCh chan io.Writer
	barCountReqCh  chan chan int
	done           chan struct{}
}

type operation struct {
	kind   opType
	bar    *Bar
	result chan bool
}

// New creates new Progress instance, which will orchestrate bars rendering
// process. It acceepts context.Context, for cancellation.
// If you don't plan to cancel, it is safe to feed with nil
func New(ctx context.Context) *Progress {
	if ctx == nil {
		ctx = context.Background()
	}
	p := &Progress{
		width:          70,
		operationCh:    make(chan *operation),
		rrChangeReqCh:  make(chan time.Duration),
		outChangeReqCh: make(chan io.Writer),
		barCountReqCh:  make(chan chan int),
		done:           make(chan struct{}),
		wg:             new(sync.WaitGroup),
		ctx:            ctx,
	}
	go p.server(cwriter.New(os.Stdout), time.NewTicker(rr*time.Millisecond))
	return p
}

// SetWidth sets the width for all underlying bars
func (p *Progress) SetWidth(n int) *Progress {
	if n <= 0 {
		return p
	}
	p.width = n
	return p
}

// SetOut sets underlying writer of progress. Default is os.Stdout
// pancis, if called on stopped Progress instance, i.e after Stop()
func (p *Progress) SetOut(w io.Writer) *Progress {
	if p.isDone() {
		panic(ErrCallAfterStop)
	}
	if w == nil {
		return p
	}
	p.outChangeReqCh <- w
	return p
}

// RefreshRate overrides default (30ms) refreshRate value
// pancis, if called on stopped Progress instance, i.e after Stop()
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	if p.isDone() {
		panic(ErrCallAfterStop)
	}
	p.rrChangeReqCh <- d
	return p
}

// func (p *Progress) WithContext(ctx context.Context) *Progress {
// 	if p.BarCount() > 0 {
// 		panic("cannot apply ctx after AddBar has been called")
// 	}
// 	if ctx == nil {
// 		panic("nil context")
// 	}
// 	p.ctx = ctx
// 	return p
// }

// WithSort sorts the bars, while redering
func (p *Progress) WithSort(sort SortType) *Progress {
	p.sort = sort
	return p
}

// AddBar creates a new progress bar and adds to the container
// pancis, if called on stopped Progress instance, i.e after Stop()
func (p *Progress) AddBar(total int64) *Bar {
	if p.isDone() {
		panic(ErrCallAfterStop)
	}
	result := make(chan bool)
	bar := newBar(p.ctx, p.wg, total, p.width)
	p.operationCh <- &operation{opBarAdd, bar, result}
	if <-result {
		p.wg.Add(1)
	}
	return bar
}

// RemoveBar removes bar at any time
// pancis, if called on stopped Progress instance, i.e after Stop()
func (p *Progress) RemoveBar(b *Bar) bool {
	if p.isDone() {
		panic(ErrCallAfterStop)
	}
	result := make(chan bool)
	p.operationCh <- &operation{opBarRemove, b, result}
	return <-result
}

// BarCount returns bars count in the container.
// Pancis if called on stopped Progress instance, i.e after Stop()
func (p *Progress) BarCount() int {
	if p.isDone() {
		panic(ErrCallAfterStop)
	}
	respCh := make(chan int)
	p.barCountReqCh <- respCh
	return <-respCh
}

// Stop waits for bars to finish rendering and stops the rendering goroutine
func (p *Progress) Stop() {
	p.wg.Wait()
	if !p.isDone() {
		close(p.done)
		close(p.operationCh)
	}
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(cw *cwriter.Writer, t *time.Ticker) {
	const numDigesters = 4
	bars := make([]*Bar, 0, 4)
	for {
		select {
		case w := <-p.outChangeReqCh:
			cw.Flush()
			cw = cwriter.New(w)
		case op, ok := <-p.operationCh:
			if !ok {
				t.Stop()
				return
			}
			switch op.kind {
			case opBarAdd:
				bars = append(bars, op.bar)
				op.result <- true
			case opBarRemove:
				var ok bool
				for i, b := range bars {
					if b == op.bar {
						bars = append(bars[:i], bars[i+1:]...)
						ok = true
						b.remove()
						break
					}
				}
				op.result <- ok
			}
		case respCh := <-p.barCountReqCh:
			respCh <- len(bars)
		case <-t.C:
			switch p.sort {
			case SortTop:
				sort.Sort(sort.Reverse(SortableBarSlice(bars)))
			case SortBottom:
				sort.Sort(SortableBarSlice(bars))
			}

			width, _ := cwriter.TerminalWidth()
			ibars := iBarsGen(p.ctx.Done(), bars, width)
			c := make(chan *indexedBarBuffer)
			var wg sync.WaitGroup
			wg.Add(numDigesters)
			for i := 0; i < numDigesters; i++ {
				go func() {
					drawer(p.ctx.Done(), ibars, c)
					wg.Done()
				}()
			}
			go func() {
				wg.Wait()
				close(c)
			}()

			m := make(map[int][]byte, len(bars))
			for r := range c {
				m[r.index] = r.buff
			}
			for i := 0; i < len(bars); i++ {
				m[i] = append(m[i], '\n')
				cw.Write(m[i])
			}
			cw.Flush()
			go flushed(p.ctx.Done(), bars)

		case d := <-p.rrChangeReqCh:
			t.Stop()
			t = time.NewTicker(d)
		case <-p.ctx.Done():
			t.Stop()
			close(p.done)
			return
		}
	}
}

type indexedBarBuffer struct {
	index int
	buff  []byte
}

type indexedBar struct {
	index int
	width int
	bar   *Bar
}

func drawer(done <-chan struct{}, ibars <-chan *indexedBar, c chan<- *indexedBarBuffer) {
	for b := range ibars {
		select {
		case c <- &indexedBarBuffer{b.index, b.bar.bytes(b.width)}:
		case <-done:
			return
		}
	}
}

func iBarsGen(done <-chan struct{}, bars []*Bar, width int) <-chan *indexedBar {
	ibars := make(chan *indexedBar)
	go func() {
		defer close(ibars)
		for i, b := range bars {
			select {
			case ibars <- &indexedBar{i, width, b}:
			case <-done:
				return
			}
		}
	}()
	return ibars
}

func flushed(done <-chan struct{}, bars []*Bar) {
	for _, b := range bars {
		select {
		case <-done:
			return
		default:
			b.flushDone()
		}
	}
}

func (p *Progress) isDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}
