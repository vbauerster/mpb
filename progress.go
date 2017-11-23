package mpb

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/vbauerster/mpb/cwriter"
)

type (
	// BeforeRender is a func, which gets called before render process
	BeforeRender func([]*Bar)

	widthSync struct {
		Listen []chan int
		Result []chan int
	}

	// progress config, fields are adjustable by user indirectly
	pConf struct {
		bars []*Bar

		idCounter    int
		width        int
		format       string
		rr           time.Duration
		ewg          *sync.WaitGroup
		cw           *cwriter.Writer
		ticker       *time.Ticker
		beforeRender BeforeRender
		interceptors []func(io.Writer)

		shutdownNotifier chan struct{}
		cancel           <-chan struct{}
	}
)

const (
	// default RefreshRate
	prr = 100 * time.Millisecond
	// default width
	pwidth = 80
	// default format
	pformat = "[=>-]"
)

// Progress represents the container that renders Progress bars
type Progress struct {
	// wg for internal rendering sync
	wg *sync.WaitGroup
	// External wg
	ewg *sync.WaitGroup

	// quit channel to request p.server to quit
	quit chan struct{}
	// done channel is receiveable after p.server has been quit
	done chan struct{}
	ops  chan func(*pConf)
}

// New creates new Progress instance, which orchestrates bars rendering process.
// Accepts mpb.ProgressOption funcs for customization.
func New(options ...ProgressOption) *Progress {
	// defaults
	conf := pConf{
		bars:   make([]*Bar, 0, 3),
		width:  pwidth,
		format: pformat,
		cw:     cwriter.New(os.Stdout),
		rr:     prr,
		ticker: time.NewTicker(prr),
		cancel: make(chan struct{}),
	}

	for _, opt := range options {
		opt(&conf)
	}

	p := &Progress{
		ewg:  conf.ewg,
		wg:   new(sync.WaitGroup),
		done: make(chan struct{}),
		ops:  make(chan func(*pConf)),
		quit: make(chan struct{}),
	}
	go p.server(conf)
	return p
}

// AddBar creates a new progress bar and adds to the container.
func (p *Progress) AddBar(total int64, options ...BarOption) *Bar {
	p.wg.Add(1)
	result := make(chan *Bar, 1)
	select {
	case p.ops <- func(c *pConf) {
		options = append(options, barWidth(c.width), barFormat(c.format))
		b := newBar(c.idCounter, total, p.wg, c.cancel, options...)
		c.bars = append(c.bars, b)
		c.idCounter++
		result <- b
	}:
		return <-result
	case <-p.quit:
		return new(Bar)
	}
}

// RemoveBar removes bar at any time.
func (p *Progress) RemoveBar(b *Bar) bool {
	result := make(chan bool, 1)
	op := func(c *pConf) {
		var ok bool
		for i, bar := range c.bars {
			if bar == b {
				bar.Complete()
				c.bars = append(c.bars[:i], c.bars[i+1:]...)
				ok = true
				break
			}
		}
		result <- ok
	}
	select {
	case p.ops <- op:
		return <-result
	case <-p.quit:
		return false
	}
}

// BarCount returns bars count
func (p *Progress) BarCount() int {
	result := make(chan int, 1)
	op := func(c *pConf) {
		result <- len(c.bars)
	}
	select {
	case p.ops <- op:
		return <-result
	case <-p.quit:
		return 0
	}
}

// Stop is a way to gracefully shutdown mpb's rendering goroutine.
// It is NOT for cancelation (use mpb.WithContext for cancelation purposes).
// If *sync.WaitGroup has been provided via mpb.WithWaitGroup(), its Wait()
// method will be called first.
func (p *Progress) Stop() {
	if p.ewg != nil {
		p.ewg.Wait()
	}
	select {
	case <-p.quit:
		return
	default:
		// wait for all bars to quit
		p.wg.Wait()
		// request p.server to quit
		p.quitRequest()
		// wait for p.server to quit
		<-p.done
	}
}

func (p *Progress) quitRequest() {
	select {
	case <-p.quit:
	default:
		close(p.quit)
	}
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(conf pConf) {
	defer func() {
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
		}
		close(p.done)
	}()

	numP, numA := -1, -1

	for {
		select {
		case op := <-p.ops:
			op(&conf)
		case <-conf.ticker.C:
			if len(conf.bars) == 0 {
				runtime.Gosched()
				break
			}
			b0 := conf.bars[0]
			if numP == -1 {
				numP = b0.NumOfPrependers()
			}
			if numA == -1 {
				numA = b0.NumOfAppenders()
			}
			err := conf.writeAndFlush(numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case <-conf.cancel:
			conf.ticker.Stop()
			conf.cancel = nil
		case <-p.quit:
			if conf.cancel != nil {
				conf.ticker.Stop()
			}
			return
		}
	}
}

func newWidthSync(timeout <-chan struct{}, numBars, numColumn int) *widthSync {
	ws := &widthSync{
		Listen: make([]chan int, numColumn),
		Result: make([]chan int, numColumn),
	}
	for i := 0; i < numColumn; i++ {
		ws.Listen[i] = make(chan int, numBars)
		ws.Result[i] = make(chan int, numBars)
	}
	for i := 0; i < numColumn; i++ {
		go func(listenCh <-chan int, resultCh chan<- int) {
			defer close(resultCh)
			widths := make([]int, 0, numBars)
		loop:
			for {
				select {
				case w := <-listenCh:
					widths = append(widths, w)
					if len(widths) == numBars {
						break loop
					}
				case <-timeout:
					if len(widths) == 0 {
						return
					}
					break loop
				}
			}
			result := max(widths)
			for i := 0; i < len(widths); i++ {
				resultCh <- result
			}
		}(ws.Listen[i], ws.Result[i])
	}
	return ws
}

func (p *pConf) writeAndFlush(numP, numA int) (err error) {
	if p.beforeRender != nil {
		p.beforeRender(p.bars)
	}

	wSyncTimeout := make(chan struct{})
	time.AfterFunc(p.rr, func() {
		close(wSyncTimeout)
	})

	prependWs := newWidthSync(wSyncTimeout, len(p.bars), numP)
	appendWs := newWidthSync(wSyncTimeout, len(p.bars), numA)

	tw, _, _ := cwriter.TermSize()

	sequence := make([]<-chan *writeBuf, len(p.bars))
	for i, b := range p.bars {
		sequence[i] = b.render(tw, prependWs, appendWs)
	}

	var i int
	for b := range fanIn(sequence...) {
		_, err = p.cw.Write(b.buf)
		defer func(bar *Bar, complete bool) {
			if complete {
				bar.Complete()
			}
		}(p.bars[i], b.completeAfterFlush)
		i++
	}

	for _, interceptor := range p.interceptors {
		interceptor(p.cw)
	}

	if e := p.cw.Flush(); err == nil {
		err = e
	}
	return
}

func fanIn(inputs ...<-chan *writeBuf) <-chan *writeBuf {
	ch := make(chan *writeBuf)

	go func() {
		defer close(ch)
		for _, input := range inputs {
			ch <- <-input
		}
	}()

	return ch
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
