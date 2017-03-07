package mpb

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/cwriter"
)

var logger = log.New(os.Stderr, "mpb: ", log.LstdFlags|log.Lshortfile)

// ErrCallAfterStop thrown by panic, if Progress methods like (*Progress).AddBar()
// are called after (*Progress).Stop() has been called
var ErrCallAfterStop = errors.New("method call on stopped Progress instance")

type (
	// BeforeRender is a func, which gets called before render process
	BeforeRender func([]*Bar)
	barOpType    uint

	operation struct {
		kind   barOpType
		bar    *Bar
		result chan bool
	}

	indexedBarBuffer struct {
		index int
		buf   []byte
	}

	indexedBar struct {
		index     int
		termWidth int
		bar       *Bar
	}
)

const (
	barAdd barOpType = iota
	barRemove
)

const (
	// default RefreshRate
	rr = 100
	// default width
	pwidth = 70
	// number of format runes for bar
	numFmtRunes = 5
)

// Progress represents the container that renders Progress bars
type Progress struct {
	// WaitGroup for internal rendering sync
	wg *sync.WaitGroup

	out    io.Writer
	width  int
	format string

	operationCh    chan *operation
	rrChangeReqCh  chan time.Duration
	outChangeReqCh chan io.Writer
	barCountReqCh  chan chan int
	brCh           chan BeforeRender
	stopCh         chan struct{}
	done           chan struct{}
}

// New Progress instance, it orchestrates the rendering of progress bars.
func New() *Progress {
	p := &Progress{
		width:          pwidth,
		operationCh:    make(chan *operation),
		rrChangeReqCh:  make(chan time.Duration),
		outChangeReqCh: make(chan io.Writer),
		barCountReqCh:  make(chan chan int),
		brCh:           make(chan BeforeRender),
		stopCh:         make(chan struct{}),
		done:           make(chan struct{}),
		wg:             new(sync.WaitGroup),
	}
	go p.server(cwriter.New(os.Stdout), time.NewTicker(rr*time.Millisecond))
	return p
}

// SetWidth overrides default (70) width of bar(s)
func (p *Progress) SetWidth(n int) *Progress {
	if n <= 0 {
		return p
	}
	p.width = n
	return p
}

// SetOut sets underlying writer of progress. Default is os.Stdout
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) SetOut(w io.Writer) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if w == nil {
		return p
	}
	p.outChangeReqCh <- w
	return p
}

// RefreshRate overrides default (100ms) refresh rate value
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	p.rrChangeReqCh <- d
	return p
}

// BeforeRenderFunc accepts a func, which gets called before render process.
func (p *Progress) BeforeRenderFunc(f BeforeRender) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	p.brCh <- f
	return p
}

// AddBar creates a new progress bar and adds to the container
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) AddBar(total int64) *Bar {
	return p.AddBarWithID(0, total)
}

// AddBarWithID creates a new progress bar and adds to the container
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) AddBarWithID(id int, total int64) *Bar {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	result := make(chan bool)
	bar := newBar(p.wg, id, total, p.width, p.format)
	p.operationCh <- &operation{barAdd, bar, result}
	if <-result {
		p.wg.Add(1)
	}
	return bar
}

// RemoveBar removes bar at any time.
// Pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) RemoveBar(b *Bar) bool {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	result := make(chan bool)
	p.operationCh <- &operation{barRemove, b, result}
	return <-result
}

// BarCount returns bars count in the container.
// Pancis if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) BarCount() int {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	respCh := make(chan int)
	p.barCountReqCh <- respCh
	return <-respCh
}

// Format sets custom format for underlying bar(s).
// The default one is "[=>-]"
func (p *Progress) Format(format string) *Progress {
	if utf8.RuneCountInString(format) != numFmtRunes {
		return p
	}
	p.format = format
	return p
}

// Stop waits for bars to finish rendering and stops the rendering goroutine
func (p *Progress) Stop() {
	if isClosed(p.done) {
		return
	}
	p.stopCh <- struct{}{}
	close(p.operationCh)
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(cw *cwriter.Writer, t *time.Ticker) {
	defer func() {
		t.Stop()
		close(p.done)
	}()
	const numDrawers = 3
	bars := make([]*Bar, 0, 4)
	var beforeRender BeforeRender
	var wg sync.WaitGroup
	recoverIfPanic := func() {
		if e := recover(); e != nil {
			logger.Printf("unexpected panic: %+v\n", e)
		}
		wg.Done()
	}

	for {
		select {
		case w := <-p.outChangeReqCh:
			cw.Flush()
			cw = cwriter.New(w)
		case op, ok := <-p.operationCh:
			if !ok {
				return
			}
			switch op.kind {
			case barAdd:
				bars = append(bars, op.bar)
				op.result <- true
			case barRemove:
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
		case beforeRender = <-p.brCh:
		case <-t.C:
			if beforeRender != nil {
				beforeRender(bars)
			}

			width, _, _ := cwriter.GetTermSize()
			ibars := iBarsGen(bars, width)
			c := make(chan indexedBarBuffer)
			wg.Add(numDrawers)
			for i := 0; i < numDrawers; i++ {
				go func() {
					defer recoverIfPanic()
					drawer(ibars, c)
				}()
			}
			go func() {
				wg.Wait()
				close(c)
			}()

			m := make(map[int][]byte, len(bars))
			for r := range c {
				m[r.index] = r.buf
			}
			for i := 0; i < len(bars); i++ {
				cw.Write(m[i])
			}

			cw.Flush()

			for _, b := range bars {
				b.flushed()
			}
		case d := <-p.rrChangeReqCh:
			t.Stop()
			t = time.NewTicker(d)
		case <-p.stopCh:
			for _, b := range bars {
				b.Stop()
			}
		case <-p.done:
			return
		}
	}
}

func drawer(ibars <-chan indexedBar, c chan<- indexedBarBuffer) {
	for b := range ibars {
		buf := b.bar.bytes(b.termWidth)
		buf = append(buf, '\n')
		c <- indexedBarBuffer{b.index, buf}
	}
}

func iBarsGen(bars []*Bar, width int) <-chan indexedBar {
	ibars := make(chan indexedBar)
	go func() {
		defer close(ibars)
		for i, b := range bars {
			ibars <- indexedBar{i, width, b}
		}
	}()
	return ibars
}

// isClosed check if ch closed
// caution see: http://www.tapirgames.com/blog/golang-channel-closing
func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
