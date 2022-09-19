package mpb

import (
	"bytes"
	"container/heap"
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
	prr = 150 * time.Millisecond // default RefreshRate
)

// DoneError represents an error when `*mpb.Progress` is done but its functionality is requested.
var DoneError = fmt.Errorf("%T instance can't be reused after it's done!", (*Progress)(nil))

// Progress represents a container that renders one or more progress bars.
type Progress struct {
	ctx          context.Context
	uwg          *sync.WaitGroup
	cwg          *sync.WaitGroup
	bwg          *sync.WaitGroup
	operateState chan func(*pState)
	interceptIo  chan func(io.Writer)
	done         chan struct{}
	once         sync.Once
	cancel       func()
}

// pState holds bars in its priorityQueue, it gets passed to (*Progress).serve monitor goroutine.
type pState struct {
	bHeap       priorityQueue
	heapUpdated bool
	pMatrix     map[int][]chan int
	aMatrix     map[int][]chan int

	// for reuse purposes
	rows []io.Reader
	pool []*Bar

	// following are provided/overrided by user
	idCount            int
	reqWidth           int
	popPriority        int
	popCompleted       bool
	outputDiscarded    bool
	disableAutoRefresh bool
	rr                 time.Duration
	uwg                *sync.WaitGroup
	externalRefresh    chan interface{}
	renderDelay        <-chan struct{}
	shutdownNotifier   chan struct{}
	queueBars          map[*Bar]*Bar
	output             io.Writer
	debugOut           io.Writer
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
	ctx, cancel := context.WithCancel(ctx)
	s := &pState{
		rr:              prr,
		bHeap:           priorityQueue{},
		rows:            make([]io.Reader, 0, 64),
		pool:            make([]*Bar, 0, 64),
		externalRefresh: make(chan interface{}),
		queueBars:       make(map[*Bar]*Bar),
		output:          os.Stdout,
		popPriority:     math.MinInt32,
		debugOut:        io.Discard,
	}

	for _, opt := range options {
		if opt != nil {
			opt(s)
		}
	}

	p := &Progress{
		ctx:          ctx,
		uwg:          s.uwg,
		cwg:          new(sync.WaitGroup),
		bwg:          new(sync.WaitGroup),
		operateState: make(chan func(*pState)),
		interceptIo:  make(chan func(io.Writer)),
		done:         make(chan struct{}),
		cancel:       cancel,
	}

	p.cwg.Add(1)
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
			heap.Push(&ps.bHeap, bar)
			ps.heapUpdated = true
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
	sync := make(chan struct{})
	select {
	case p.operateState <- func(s *pState) {
		for i := 0; i < s.bHeap.Len(); i++ {
			bar := s.bHeap[i]
			if !cb(bar) {
				break
			}
		}
		close(sync)
	}:
		<-sync
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
		heap.Fix(&s.bHeap, b.index)
	}:
	case <-p.done:
	}
}

// BarCount returns bars count.
func (p *Progress) BarCount() int {
	result := make(chan int)
	select {
	case p.operateState <- func(s *pState) { result <- s.bHeap.Len() }:
		return <-result
	case <-p.done:
		return 0
	}
}

// Write is implementation of io.Writer.
// Writing to `*mpb.Progress` will print lines above a running bar.
// Writes aren't flushed immediatly, but at next refresh cycle.
// If Write is called after `*mpb.Progress` is done, `mpb.DoneError`
// is returned.
func (p *Progress) Write(b []byte) (int, error) {
	type result struct {
		n   int
		err error
	}
	ch := make(chan *result)
	select {
	case p.interceptIo <- func(w io.Writer) {
		n, err := w.Write(b)
		ch <- &result{n, err}
	}:
		res := <-ch
		return res.n, res.err
	case <-p.done:
		return 0, DoneError
	}
}

// Wait waits for all bars to complete and finally shutdowns container.
// After this method has been called, there is no way to reuse *Progress
// instance.
func (p *Progress) Wait() {
	// wait for user wg, if any
	if p.uwg != nil {
		p.uwg.Wait()
	}

	// wait for bars to quit, if any
	p.bwg.Wait()

	p.once.Do(p.shutdown)

	// wait for container to quit
	p.cwg.Wait()
}

func (p *Progress) shutdown() {
	close(p.done)
}

func (p *Progress) serve(s *pState, cw *cwriter.Writer) {
	defer p.cwg.Done()

	refreshCh := s.newTicker(p.done)

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case fn := <-p.interceptIo:
			fn(cw)
		case <-refreshCh:
			err := s.render(cw)
			if err != nil {
				p.cancel() // cancel all bars
				p.once.Do(p.shutdown)
				s.heapUpdated = false
				refreshCh = nil
				_, _ = fmt.Fprintln(s.debugOut, err)
			}
		case <-s.shutdownNotifier:
			for s.heapUpdated {
				err := s.render(cw)
				if err != nil {
					_, _ = fmt.Fprintln(s.debugOut, err)
					return
				}
			}
			return
		}
	}
}

func (s *pState) render(cw *cwriter.Writer) error {
	if s.heapUpdated {
		s.updateSyncMatrix()
		s.heapUpdated = false
	}
	syncWidth(s.pMatrix)
	syncWidth(s.aMatrix)

	width, height, err := cw.GetTermSize()
	if err != nil {
		width = s.reqWidth
		height = s.bHeap.Len()
	}
	for i := 0; i < s.bHeap.Len(); i++ {
		bar := s.bHeap[i]
		go bar.render(width)
	}

	return s.flush(cw, height)
}

func (s *pState) flush(cw *cwriter.Writer, height int) (err error) {
	var wg sync.WaitGroup
	var popCount int

	for s.bHeap.Len() > 0 {
		b := heap.Pop(&s.bHeap).(*Bar)
		frame := <-b.frameCh
		if frame.err != nil {
			if err == nil {
				err = frame.err
			}
			continue
		}
		var usedRows int
		for i := len(frame.rows) - 1; i >= 0; i-- {
			if row := frame.rows[i]; len(s.rows) < height {
				s.rows = append(s.rows, row)
				usedRows++
			} else {
				wg.Add(1)
				go func() {
					_, _ = io.Copy(io.Discard, row)
					wg.Done()
				}()
			}
		}
		if frame.shutdown != 0 {
			b.Wait() // waiting for b.done, so it's safe to read b.bs
			drop := b.bs.dropOnComplete
			if qb, ok := s.queueBars[b]; ok {
				delete(s.queueBars, b)
				qb.priority = b.priority
				s.pool = append(s.pool, qb)
				drop = true
			} else if s.popCompleted && !b.bs.noPop {
				if frame.shutdown > 1 {
					popCount += usedRows
					drop = true
				} else {
					s.popPriority++
					b.priority = s.popPriority
				}
			}
			if drop {
				s.heapUpdated = true
				continue
			}
		}
		s.pool = append(s.pool, b)
	}

	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		for _, b := range s.pool {
			heap.Push(&s.bHeap, b)
		}
		wg.Done()
	}()

	for i := len(s.rows) - 1; i >= 0; i-- {
		_, err := cw.ReadFrom(s.rows[i])
		if err != nil {
			wg.Wait()
			return err
		}
	}

	err = cw.Flush(len(s.rows) - popCount)
	wg.Wait()
	s.rows = s.rows[:0]
	s.pool = s.pool[:0]
	return err
}

func (s *pState) newTicker(done <-chan struct{}) chan time.Time {
	ch := make(chan time.Time)
	if s.shutdownNotifier == nil {
		s.shutdownNotifier = make(chan struct{})
	}
	go func() {
		var autoRefresh <-chan time.Time
		if !s.disableAutoRefresh {
			if !s.outputDiscarded {
				if s.renderDelay != nil {
					<-s.renderDelay
				}
				ticker := time.NewTicker(s.rr)
				defer ticker.Stop()
				autoRefresh = ticker.C
			}
		}
		for {
			select {
			case t := <-autoRefresh:
				ch <- t
			case x := <-s.externalRefresh:
				if t, ok := x.(time.Time); ok {
					ch <- t
				} else {
					ch <- time.Now()
				}
			case <-done:
				close(s.shutdownNotifier)
				return
			}
		}
	}()
	return ch
}

func (s *pState) updateSyncMatrix() {
	s.pMatrix = make(map[int][]chan int)
	s.aMatrix = make(map[int][]chan int)
	for i := 0; i < s.bHeap.Len(); i++ {
		bar := s.bHeap[i]
		table := bar.wSyncTable()
		pRow, aRow := table[0], table[1]

		for i, ch := range pRow {
			s.pMatrix[i] = append(s.pMatrix[i], ch)
		}

		for i, ch := range aRow {
			s.aMatrix[i] = append(s.aMatrix[i], ch)
		}
	}
}

func (s *pState) makeBarState(total int64, filler BarFiller, options ...BarOption) *bState {
	bs := &bState{
		id:        s.idCount,
		priority:  s.idCount,
		reqWidth:  s.reqWidth,
		total:     total,
		filler:    filler,
		refreshCh: s.externalRefresh,
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

func syncWidth(matrix map[int][]chan int) {
	for _, column := range matrix {
		go maxWidthDistributor(column)
	}
}

func maxWidthDistributor(column []chan int) {
	var maxWidth int
	for _, ch := range column {
		if w := <-ch; w > maxWidth {
			maxWidth = w
		}
	}
	for _, ch := range column {
		ch <- maxWidth
	}
}
