package mpb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8/cwriter"
)

const (
	defaultRefreshRate = 150 * time.Millisecond
)

// DoneError represents an error when `*mpb.Progress` is done but its functionality is requested.
var DoneError = fmt.Errorf("%T instance can't be reused after it's done!", (*Progress)(nil))

// Progress represents a container that renders one or more progress bars.
type Progress struct {
	ctx          context.Context
	uwg          *sync.WaitGroup
	bwg          *sync.WaitGroup
	operateState chan func(*pState)
	interceptIO  chan func(io.Writer)
	done         chan struct{}
	shutdown     chan struct{}
	cancel       func()
}

// pState holds bars in its priorityQueue, it gets passed to (*Progress).serve monitor goroutine.
type pState struct {
	hm   heapManager
	rows []io.Reader

	// following are provided/overrided by user
	refreshRate        time.Duration
	idCount            int
	reqWidth           int
	popPriority        int
	popCompleted       bool
	disableAutoRefresh bool
	forceAutoRefresh   bool
	manualRefresh      chan interface{}
	renderDelay        <-chan struct{}
	shutdownNotifier   chan<- interface{}
	queueBars          map[*Bar]*Bar
	output             io.Writer
	debugOut           io.Writer
	uwg                *sync.WaitGroup
}

// New creates new Progress container instance. It's not possible to
// reuse instance after (*Progress).Wait method has been called.
func New(options ...ContainerOption) *Progress {
	return NewWithContext(context.Background(), options...)
}

// NewWithContext creates new Progress container instance with provided
// context. It's not possible to reuse instance after (*Progress).Wait
// method has been called.
func NewWithContext(ctx context.Context, options ...ContainerOption) *Progress {
	s := &pState{
		hm:          make(heapManager),
		rows:        make([]io.Reader, 32),
		refreshRate: defaultRefreshRate,
		popPriority: math.MinInt32,
		queueBars:   make(map[*Bar]*Bar),
		output:      os.Stdout,
		debugOut:    io.Discard,
	}

	for _, opt := range options {
		if opt != nil {
			opt(s)
		}
	}

	if s.manualRefresh == nil {
		s.manualRefresh = make(chan interface{})
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &Progress{
		ctx:          ctx,
		uwg:          s.uwg,
		bwg:          new(sync.WaitGroup),
		operateState: make(chan func(*pState)),
		interceptIO:  make(chan func(io.Writer)),
		done:         make(chan struct{}),
		shutdown:     make(chan struct{}),
		cancel:       cancel,
	}

	go p.serve(s, cwriter.New(s.output))
	return p
}

// AddBar creates a bar with default bar filler.
func (p *Progress) AddBar(total int64, options ...BarOption) *Bar {
	return p.New(total, BarStyle(), options...)
}

// AddSpinner creates a bar with default spinner filler.
func (p *Progress) AddSpinner(total int64, options ...BarOption) *Bar {
	return p.New(total, SpinnerStyle(), options...)
}

// New creates a bar by calling `Build` method on provided `BarFillerBuilder`.
func (p *Progress) New(total int64, builder BarFillerBuilder, options ...BarOption) *Bar {
	return p.AddFiller(total, builder.Build(), options...)
}

// AddFiller creates a bar which renders itself by provided filler.
// If `total <= 0` triggering complete event by increment methods is disabled.
// Panics if *Progress instance is done, i.e. called after (*Progress).Wait().
func (p *Progress) AddFiller(total int64, filler BarFiller, options ...BarOption) *Bar {
	if filler == nil {
		filler = NopStyle().Build()
	}
	p.bwg.Add(1)
	result := make(chan *Bar)
	select {
	case p.operateState <- func(ps *pState) {
		bs := ps.makeBarState(total, filler, options...)
		bar := newBar(p, bs)
		if bs.wait.bar != nil {
			ps.queueBars[bs.wait.bar] = bar
		} else {
			ps.hm.push(bar, true)
		}
		ps.idCount++
		result <- bar
	}:
		bar := <-result
		return bar
	case <-p.done:
		p.bwg.Done()
		panic(DoneError)
	}
}

func (p *Progress) traverseBars(cb func(b *Bar) bool) {
	iter, drop := make(chan *Bar), make(chan struct{})
	select {
	case p.operateState <- func(s *pState) { s.hm.iter(iter, drop) }:
		for b := range iter {
			if cb(b) {
				close(drop)
				break
			}
		}
	case <-p.done:
	}
}

// UpdateBarPriority same as *Bar.SetPriority(int).
func (p *Progress) UpdateBarPriority(b *Bar, priority int) {
	select {
	case p.operateState <- func(s *pState) {
		if b.index < 0 {
			return
		}
		b.priority = priority
		s.hm.fix(b.index)
	}:
	case <-p.done:
	}
}

// Write is implementation of io.Writer.
// Writing to `*mpb.Progress` will print lines above a running bar.
// Writes aren't flushed immediately, but at next refresh cycle.
// If Write is called after `*mpb.Progress` is done, `mpb.DoneError`
// is returned.
func (p *Progress) Write(b []byte) (int, error) {
	type result struct {
		n   int
		err error
	}
	ch := make(chan result)
	select {
	case p.interceptIO <- func(w io.Writer) {
		n, err := w.Write(b)
		ch <- result{n, err}
	}:
		res := <-ch
		return res.n, res.err
	case <-p.done:
		return 0, DoneError
	}
}

// Wait waits for all bars to complete and finally shutdowns container. After
// this method has been called, there is no way to reuse (*Progress) instance.
func (p *Progress) Wait() {
	// wait for user wg, if any
	if p.uwg != nil {
		p.uwg.Wait()
	}

	p.bwg.Wait()
	p.Shutdown()
}

// Shutdown cancels any running bar immediately and then shutdowns (*Progress)
// instance. Normally this method shouldn't be called unless you know what you
// are doing. Proper way to shutdown is to call (*Progress).Wait() instead.
func (p *Progress) Shutdown() {
	p.cancel()
	<-p.shutdown
}

func (p *Progress) serve(s *pState, cw *cwriter.Writer) {
	var err error
	render := func() error { return s.render(cw) }
	tickerC := s.newTicker(p.ctx, cw.IsTerminal(), p.done)

	go s.hm.run()

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case fn := <-p.interceptIO:
			fn(cw)
		case <-tickerC:
			e := render()
			if e != nil {
				p.cancel() // cancel all bars
				render = func() error { return nil }
				err = e
			}
		case <-p.done:
			update := make(chan bool)
			for err == nil {
				s.hm.state(update)
				if <-update {
					err = render()
				} else {
					break
				}
			}
			if err != nil {
				_, _ = fmt.Fprintln(s.debugOut, err.Error())
			}
			s.hm.end(s.shutdownNotifier)
			close(p.shutdown)
			return
		}
	}
}

func (s *pState) newTicker(ctx context.Context, isTerminal bool, done chan struct{}) chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		var autoRefresh <-chan time.Time
		if (isTerminal || s.forceAutoRefresh) && !s.disableAutoRefresh {
			if s.renderDelay != nil {
				<-s.renderDelay
			}
			ticker := time.NewTicker(s.refreshRate)
			defer ticker.Stop()
			autoRefresh = ticker.C
		}
		for {
			select {
			case t := <-autoRefresh:
				ch <- t
			case x := <-s.manualRefresh:
				if t, ok := x.(time.Time); ok {
					ch <- t
				} else {
					ch <- time.Now()
				}
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()
	return ch
}

func (s *pState) render(cw *cwriter.Writer) (err error) {
	var width, height int
	if cw.IsTerminal() {
		width, height, err = cw.GetTermSize()
		if err != nil {
			return err
		}
	} else {
		width = s.reqWidth
		height = 100
	}

	s.hm.sync()
	iter := make(chan *Bar)
	s.hm.iter(iter, nil)
	for b := range iter {
		go b.render(width)
	}

	return s.flush(cw, height)
}

func (s *pState) flush(cw *cwriter.Writer, height int) error {
	wg := new(sync.WaitGroup)
	defer wg.Wait() // waiting for all s.hm.push to complete

	var popCount int
	s.rows = s.rows[:0]

	iter, drop := make(chan *Bar), make(chan struct{})
	s.hm.drain(iter, drop)

	for b := range iter {
		frame := <-b.frameCh
		if frame.err != nil {
			close(drop)
			b.cancel()
			return frame.err // b.frameCh is buffered it's ok to return here
		}
		var usedRows int
		for i := len(frame.rows) - 1; i >= 0; i-- {
			if row := frame.rows[i]; len(s.rows) < height {
				s.rows = append(s.rows, row)
				usedRows++
			} else {
				_, _ = io.Copy(io.Discard, row)
			}
		}
		if frame.shutdown {
			b.Wait() // waiting for b.done, so it's safe to read b.bs
			if qb, ok := s.queueBars[b]; ok {
				delete(s.queueBars, b)
				qb.priority = b.priority
				wg.Add(1)
				go func(b *Bar) {
					s.hm.push(b, true)
					wg.Done()
				}(qb)
				continue
			}
			if s.popCompleted && !b.bs.noPop {
				switch b.bs.shutdown++; b.bs.shutdown {
				case 1:
					b.priority = s.popPriority
					s.popPriority++
				default:
					if b.bs.dropOnComplete {
						popCount += usedRows
						continue
					}
				}
			} else if b.bs.dropOnComplete {
				continue
			}
		}
		wg.Add(1)
		go func(b *Bar) {
			s.hm.push(b, false)
			wg.Done()
		}(b)
	}

	for i := len(s.rows) - 1; i >= 0; i-- {
		_, err := cw.ReadFrom(s.rows[i])
		if err != nil {
			return err
		}
	}

	return cw.Flush(len(s.rows) - popCount)
}

func (s *pState) makeBarState(total int64, filler BarFiller, options ...BarOption) *bState {
	bs := &bState{
		id:            s.idCount,
		priority:      s.idCount,
		reqWidth:      s.reqWidth,
		total:         total,
		filler:        filler,
		manualRefresh: s.manualRefresh,
	}

	if total > 0 {
		bs.triggerComplete = true
	}

	for _, opt := range options {
		if opt != nil {
			opt(bs)
		}
	}

	if bs.middleware != nil {
		bs.filler = bs.middleware(filler)
		bs.middleware = nil
	}

	for i := 0; i < len(bs.buffers); i++ {
		bs.buffers[i] = bytes.NewBuffer(make([]byte, 0, 512))
	}

	bs.subscribeDecorators()

	return bs
}
