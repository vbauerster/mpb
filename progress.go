package mpb

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
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
	// Context for canceling bars rendering
	ctx context.Context
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
	done           chan struct{}
}

// New creates new Progress instance, which will orchestrate bars rendering
// process. It acceepts context.Context, for cancellation.
// If you don't plan to cancel, it is safe to feed with nil
func New(ctx context.Context) *Progress {
	if ctx == nil {
		ctx = context.Background()
	}
	p := &Progress{
		width:          pwidth,
		operationCh:    make(chan *operation),
		rrChangeReqCh:  make(chan time.Duration),
		outChangeReqCh: make(chan io.Writer),
		barCountReqCh:  make(chan chan int),
		brCh:           make(chan BeforeRender),
		done:           make(chan struct{}),
		wg:             new(sync.WaitGroup),
		ctx:            ctx,
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
	bar := newBar(p.ctx, p.wg, id, total, p.width, p.format)
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
	respCh := make(chan int, 1)
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
	p.wg.Wait()
	if isClosed(p.done) {
		return
	}
	close(p.operationCh)
}

type widthSync struct {
	listen []chan int
	result []chan int
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(cw *cwriter.Writer, t *time.Ticker) {
	defer func() {
		t.Stop()
		close(p.done)
	}()
	bars := make([]*Bar, 0, 3)
	var beforeRender BeforeRender
	var wg sync.WaitGroup
	recoverIfPanic := func() {
		if p := recover(); p != nil {
			logger.Printf("unexpected panic: %+v\n", p)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			os.Stderr.Write(buf[:n])
		}
		wg.Done()
	}
	var numPrependers int
	// var numAppenders int
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
			if len(bars) > 0 {
				numPrependers = len(bars[0].GetPrependers())
				// numAppenders = len(bars[0].GetAppenders())
			}
		case respCh := <-p.barCountReqCh:
			respCh <- len(bars)
		case beforeRender = <-p.brCh:
		case <-t.C:
			numBars := len(bars)

			if numBars == 0 {
				break
			}

			if beforeRender != nil {
				beforeRender(bars)
			}

			prepWidthSync := &widthSync{
				listen: make([]chan int, numPrependers),
				result: make([]chan int, numPrependers),
			}
			for i := 0; i < numPrependers; i++ {
				prepWidthSync.listen[i] = make(chan int, numBars)
				prepWidthSync.result[i] = make(chan int, numBars)
			}
			stopWidthListen := make(chan struct{})
			for i, listenCh := range prepWidthSync.listen {
				go func(listenCh <-chan int, resultCh chan<- int) {
					widths := make([]int, 0, numBars)
				loop:
					for {
						select {
						case w := <-listenCh:
							widths = append(widths, w)
							if len(widths) == numBars {
								break loop
							}
						case <-stopWidthListen:
							return
						}
					}
					result := max(widths)
					for i := 0; i < numBars; i++ {
						resultCh <- result
					}
					// close(resultCh)
				}(listenCh, prepWidthSync.result[i])
			}

			width, _, _ := cwriter.GetTermSize()
			ibars := iBarsGen(bars, width)
			ibbCh := make(chan indexedBarBuffer)
			wg.Add(numBars)
			for i := 0; i < numBars; i++ {
				go func() {
					defer recoverIfPanic()
					drawer(ibars, ibbCh, prepWidthSync)
				}()
			}
			go func() {
				wg.Wait()
				close(ibbCh)
				close(stopWidthListen)
				for _, ch := range prepWidthSync.result {
					close(ch)
				}
				for _, ch := range prepWidthSync.listen {
					close(ch)
				}
			}()

			m := make(map[int][]byte, len(bars))
			for ibb := range ibbCh {
				m[ibb.index] = ibb.buf
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
		case <-p.ctx.Done():
			return
		}
	}
}

func drawer(ibars <-chan indexedBar, ibbCh chan<- indexedBarBuffer) {
	for b := range ibars {
		buf := b.bar.bytes(b.termWidth, ws)
		buf = append(buf, '\n')
		ibbCh <- indexedBarBuffer{b.index, buf}
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

func max(slice []int) int {
	max := slice[0]

	for i := 1; i < len(slice); i++ {
		if slice[i] > max {
			max = slice[i]
		}
	}

	return max
}
