package mpb

import (
	"container/heap"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/cwriter"
)

const (
	// default RefreshRate
	prr = 120 * time.Millisecond
	// default width
	pwidth = 80
	// default format
	pformat = "[=>-]"
)

// Progress represents the container that renders Progress bars
type Progress struct {
	operateState chan func(*pState)
	done         chan struct{}
}

type (
	// progress state, which may contain several bars
	pState struct {
		bHeap       *priorityQueue
		heapUpdated bool
		zeroWait    bool
		idCounter   int
		width       int
		format      string
		rr          time.Duration
		cw          *cwriter.Writer
		ticker      *time.Ticker

		// following are provided by user
		uwg              *sync.WaitGroup
		cancel           <-chan struct{}
		shutdownNotifier chan struct{}
		interceptors     []func(io.Writer)
	}
	widthSyncer struct {
		// Public for easy testing
		Accumulator []chan int
		Distributor []chan int
	}
	barRendering struct {
		bar   *Bar
		ready <-chan *renderedReader
	}
)

// New creates new Progress instance, which orchestrates bars rendering process.
// Accepts mpb.ProgressOption funcs for customization.
func New(options ...ProgressOption) *Progress {
	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	s := &pState{
		bHeap:  &pq,
		width:  pwidth,
		format: pformat,
		cw:     cwriter.New(os.Stdout),
		rr:     prr,
		ticker: time.NewTicker(prr),
	}

	for _, opt := range options {
		opt(s)
	}

	p := &Progress{
		operateState: make(chan func(*pState)),
		done:         make(chan struct{}),
	}
	go p.serve(s)
	return p
}

// AddBar creates a new progress bar and adds to the container.
func (p *Progress) AddBar(total int64, options ...BarOption) *Bar {
	result := make(chan *Bar, 1)
	select {
	case p.operateState <- func(s *pState) {
		options = append(options, barWidth(s.width), barFormat(s.format))
		b := newBar(s.idCounter, total, s.cancel, options...)
		heap.Push(s.bHeap, b)
		s.heapUpdated = true
		s.idCounter++
		result <- b
	}:
		return <-result
	case <-p.done:
		// fail early
		return nil
	}
}

// RemoveBar removes the bar at next render cycle
func (p *Progress) RemoveBar(b *Bar) bool {
	return b.askToComplete(true)
}

// UpdateBarPriority provides a way to change bar's order position.
// Zero is highest priority, i.e. bar will be on top.
func (p *Progress) UpdateBarPriority(b *Bar, priority int) {
	select {
	case p.operateState <- func(s *pState) { s.bHeap.update(b, priority) }:
	case <-p.done:
	}
}

// BarCount returns bars count
func (p *Progress) BarCount() int {
	result := make(chan int, 1)
	select {
	case p.operateState <- func(s *pState) { result <- s.bHeap.Len() }:
		return <-result
	case <-p.done:
		return 0
	}
}

// Wait first waits for all bars to complete, then waits for user provided WaitGroup, if any.
// It's optional to call, in other words if you don't call Progress.Wait(),
// it's not guaranteed that all bars will be flushed completely to the underlying io.Writer.
func (p *Progress) Wait() {
	if p.BarCount() == 0 {
		select {
		case p.operateState <- func(s *pState) { s.zeroWait = true }:
		case <-p.done:
		}
		return
	}
	<-p.done
}

func newWidthSyncer(timeout <-chan struct{}, numBars, numColumn int) *widthSyncer {
	ws := &widthSyncer{
		Accumulator: make([]chan int, numColumn),
		Distributor: make([]chan int, numColumn),
	}
	for i := 0; i < numColumn; i++ {
		ws.Accumulator[i] = make(chan int, numBars)
		ws.Distributor[i] = make(chan int, numBars)
	}
	for i := 0; i < numColumn; i++ {
		go func(accumulator <-chan int, distributor chan<- int) {
			defer close(distributor)
			widths := make([]int, 0, numBars)
		loop:
			for {
				select {
				case w := <-accumulator:
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
			maxWidth := calcMax(widths)
			for i := 0; i < len(widths); i++ {
				distributor <- maxWidth
			}
		}(ws.Accumulator[i], ws.Distributor[i])
	}
	return ws
}

func (s *pState) writeAndFlush(tw, numP, numA int) (err error) {
	timeout := make(chan struct{})
	pSyncer := newWidthSyncer(timeout, s.bHeap.Len(), numP)
	aSyncer := newWidthSyncer(timeout, s.bHeap.Len(), numA)
	time.AfterFunc(s.rr-s.rr/12, func() {
		close(timeout)
	})

	for _, br := range s.renderByPriority(tw, pSyncer, aSyncer) {
		r := <-br.ready
		_, err = s.cw.ReadFrom(r)
		if !br.bar.completed && r.toComplete {
			close(br.bar.shutdown)
			br.bar.completed = true
		}
		if r.toRemove {
			s.heapUpdated = heap.Remove(s.bHeap, br.bar.index) != nil
		}
	}

	for _, interceptor := range s.interceptors {
		interceptor(s.cw)
	}

	if e := s.cw.Flush(); err == nil {
		err = e
	}
	return
}

func (s *pState) renderByPriority(tw int, pSyncer, aSyncer *widthSyncer) []*barRendering {
	slice := make([]*barRendering, 0, s.bHeap.Len())
	for s.bHeap.Len() > 0 {
		b := heap.Pop(s.bHeap).(*Bar)
		defer heap.Push(s.bHeap, b)
		slice = append(slice, &barRendering{
			bar:   b,
			ready: b.render(tw, pSyncer, aSyncer),
		})
	}
	return slice
}

func (s *pState) waitAll() {
	for s.bHeap.Len() > 0 {
		b := heap.Pop(s.bHeap).(*Bar)
		<-b.done
	}
	if s.uwg != nil {
		s.uwg.Wait()
	}
}

func calcMax(slice []int) int {
	if len(slice) == 0 {
		return 0
	}

	max := slice[0]
	for i := 1; i < len(slice); i++ {
		if slice[i] > max {
			max = slice[i]
		}
	}
	return max
}
