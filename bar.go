package mpb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/v4/decor"
)

// Filler interface.
// Bar renders by calling Filler's Fill method. You can literally have
// any bar kind, by implementing this interface and passing it to the
// Add method.
type Filler interface {
	Fill(w io.Writer, width int, stat *decor.Statistics)
}

// FillerFunc is function type adapter to convert function into Filler.
type FillerFunc func(w io.Writer, width int, stat *decor.Statistics)

func (f FillerFunc) Fill(w io.Writer, width int, stat *decor.Statistics) {
	f(w, width, stat)
}

// Bar represents a progress Bar.
type Bar struct {
	priority int
	index    int

	runningBar   *Bar
	cacheState   *bState
	operateState chan func(*bState)
	bFrameCh     chan *bFrame
	syncTableCh  chan [][]chan int
	completed    chan bool
	forceRefresh chan<- time.Time

	// done is closed by Bar's goroutine, after cacheState is written
	done chan struct{}
	// shutdown is closed from master Progress goroutine only
	shutdown chan struct{}

	arbitraryCurrent struct {
		sync.Mutex
		current int64
	}

	dlogger *log.Logger
}

type (
	bState struct {
		filler             Filler
		extender           Filler
		id                 int
		width              int
		total              int64
		current            int64
		trimSpace          bool
		toComplete         bool
		removeOnComplete   bool
		barClearOnComplete bool
		completeFlushed    bool
		aDecorators        []decor.Decorator
		pDecorators        []decor.Decorator
		amountReceivers    []decor.AmountReceiver
		shutdownListeners  []decor.ShutdownListener
		bufP, bufB, bufA   *bytes.Buffer
		bufE               *bytes.Buffer
		panicMsg           string

		// following options are assigned to the *Bar
		priority   int
		runningBar *Bar
	}
	bFrame struct {
		rd               io.Reader
		extendedLines    int
		toShutdown       bool
		removeOnComplete bool
	}
)

func newBar(
	ctx context.Context,
	wg *sync.WaitGroup,
	forceRefresh chan<- time.Time,
	bs *bState,
	dlogger *log.Logger,
) *Bar {

	bs.bufP = bytes.NewBuffer(make([]byte, 0, bs.width))
	bs.bufB = bytes.NewBuffer(make([]byte, 0, bs.width))
	bs.bufA = bytes.NewBuffer(make([]byte, 0, bs.width))
	if bs.extender != nil {
		bs.bufE = bytes.NewBuffer(make([]byte, 0, bs.width))
	}

	bar := &Bar{
		priority:     bs.priority,
		runningBar:   bs.runningBar,
		operateState: make(chan func(*bState)),
		bFrameCh:     make(chan *bFrame, 1),
		syncTableCh:  make(chan [][]chan int),
		completed:    make(chan bool),
		done:         make(chan struct{}),
		shutdown:     make(chan struct{}),
		forceRefresh: forceRefresh,
		dlogger:      dlogger,
	}

	go bar.serve(ctx, wg, bs)
	return bar
}

// RemoveAllPrependers removes all prepend functions.
func (b *Bar) RemoveAllPrependers() {
	select {
	case b.operateState <- func(s *bState) { s.pDecorators = nil }:
	case <-b.done:
	}
}

// RemoveAllAppenders removes all append functions.
func (b *Bar) RemoveAllAppenders() {
	select {
	case b.operateState <- func(s *bState) { s.aDecorators = nil }:
	case <-b.done:
	}
}

// ProxyReader wraps r with metrics required for progress tracking.
func (b *Bar) ProxyReader(r io.Reader) io.ReadCloser {
	if r == nil {
		return nil
	}
	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(r)
	}
	return &proxyReader{rc, b, time.Now()}
}

// ID returs id of the bar.
func (b *Bar) ID() int {
	result := make(chan int)
	select {
	case b.operateState <- func(s *bState) { result <- s.id }:
		return <-result
	case <-b.done:
		return b.cacheState.id
	}
}

// Current returns bar's current number, in other words sum of all increments.
func (b *Bar) Current() int64 {
	result := make(chan int64)
	select {
	case b.operateState <- func(s *bState) { result <- s.current }:
		return <-result
	case <-b.done:
		return b.cacheState.current
	}
}

// SetRefill sets refill, if supported by underlying Filler.
func (b *Bar) SetRefill(amount int64) {
	b.operateState <- func(s *bState) {
		if f, ok := s.filler.(interface{ SetRefill(int64) }); ok {
			f.SetRefill(amount)
		}
	}
}

// SetTotal sets total dynamically.
// Set final to true, when total is known, it will trigger bar complete event.
func (b *Bar) SetTotal(total int64, final bool) bool {
	select {
	case b.operateState <- func(s *bState) {
		if total > 0 {
			s.total = total
		}
		if final && !s.toComplete {
			s.current = s.total
			s.toComplete = true
			go b.refreshNowTillShutdown()
		}
	}:
		return true
	case <-b.done:
		return false
	}
}

// SetCurrent sets progress' current to arbitrary amount.
func (b *Bar) SetCurrent(current int64, wdd ...time.Duration) {
	if current <= 0 {
		return
	}
	b.arbitraryCurrent.Lock()
	last := b.arbitraryCurrent.current
	b.IncrBy(int(current-last), wdd...)
	b.arbitraryCurrent.current = current
	b.arbitraryCurrent.Unlock()
}

// Increment is a shorthand for b.IncrBy(1).
func (b *Bar) Increment() {
	b.IncrBy(1)
}

// IncrBy increments progress bar by amount of n.
// wdd is optional work duration i.e. time.Since(start), which expected
// to be provided, if any ewma based decorator is used.
func (b *Bar) IncrBy(n int, wdd ...time.Duration) {
	select {
	case b.operateState <- func(s *bState) {
		s.current += int64(n)
		if s.current >= s.total && !s.toComplete {
			s.current = s.total
			s.toComplete = true
			go b.refreshNowTillShutdown()
		}
		for _, ar := range s.amountReceivers {
			ar.NextAmount(n, wdd...)
		}
	}:
	case <-b.done:
	}
}

// Completed reports whether the bar is in completed state.
func (b *Bar) Completed() bool {
	// omit select here, because primary usage of the method is for loop
	// condition, like for !bar.Completed() {...} so when toComplete=true
	// it is called once (at which time, the bar is still alive), then
	// quits the loop and never suppose to be called afterwards.
	return <-b.completed
}

func (b *Bar) wSyncTable() [][]chan int {
	select {
	case b.operateState <- func(s *bState) { b.syncTableCh <- s.wSyncTable() }:
		return <-b.syncTableCh
	case <-b.done:
		return b.cacheState.wSyncTable()
	}
}

func (b *Bar) serve(ctx context.Context, wg *sync.WaitGroup, s *bState) {
	defer wg.Done()
	cancel := ctx.Done()
	for {
		select {
		case op := <-b.operateState:
			op(s)
		case b.completed <- s.toComplete:
		case <-cancel:
			s.toComplete = true
			cancel = nil
		case <-b.shutdown:
			b.cacheState = s
			close(b.done)
			// Notifying decorators about shutdown event
			for _, sl := range s.shutdownListeners {
				sl.Shutdown()
			}
			return
		}
	}
}

func (b *Bar) render(debugOut io.Writer, tw int) {
	select {
	case b.operateState <- func(s *bState) {
		defer func() {
			// recovering if user defined decorator panics for example
			if p := recover(); p != nil {
				s.panicMsg = fmt.Sprintf("panic: %v", p)
				b.dlogger.Println(s.panicMsg)
				b.bFrameCh <- &bFrame{
					rd:         strings.NewReader(fmt.Sprintf(fmt.Sprintf("%%.%ds\n", tw), s.panicMsg)),
					toShutdown: true,
				}
			}
		}()
		r := s.draw(tw)
		var extendedLines int
		if s.extender != nil {
			s.extender.Fill(s.bufE, tw, newStatistics(s))
			extendedLines = countLines(s.bufE.Bytes())
			r = io.MultiReader(r, s.bufE)
		}
		b.bFrameCh <- &bFrame{
			rd:               r,
			extendedLines:    extendedLines,
			toShutdown:       s.toComplete && !s.completeFlushed,
			removeOnComplete: s.removeOnComplete,
		}
		s.completeFlushed = s.toComplete
	}:
	case <-b.done:
		s := b.cacheState
		r := s.draw(tw)
		var extendedLines int
		if s.extender != nil {
			s.extender.Fill(s.bufE, tw, newStatistics(s))
			extendedLines = countLines(s.bufE.Bytes())
			r = io.MultiReader(r, s.bufE)
		}
		b.bFrameCh <- &bFrame{
			rd:            r,
			extendedLines: extendedLines,
		}
	}
}

func (s *bState) draw(termWidth int) io.Reader {
	if s.panicMsg != "" {
		return strings.NewReader(fmt.Sprintf(fmt.Sprintf("%%.%ds\n", termWidth), s.panicMsg))
	}

	stat := newStatistics(s)

	for _, d := range s.pDecorators {
		s.bufP.WriteString(d.Decor(stat))
	}

	for _, d := range s.aDecorators {
		s.bufA.WriteString(d.Decor(stat))
	}

	if s.barClearOnComplete && s.completeFlushed {
		s.bufA.WriteByte('\n')
		return io.MultiReader(s.bufP, s.bufA)
	}

	prependCount := utf8.RuneCount(s.bufP.Bytes())
	appendCount := utf8.RuneCount(s.bufA.Bytes())

	if !s.trimSpace {
		// reserve space for edge spaces
		termWidth -= 2
		s.bufB.WriteByte(' ')
	}

	calcWidth := s.width
	if prependCount+s.width+appendCount > termWidth {
		calcWidth = termWidth - prependCount - appendCount
	}
	s.filler.Fill(s.bufB, calcWidth, stat)

	if !s.trimSpace {
		s.bufB.WriteByte(' ')
	}

	s.bufA.WriteByte('\n')
	return io.MultiReader(s.bufP, s.bufB, s.bufA)
}

func (s *bState) wSyncTable() [][]chan int {
	columns := make([]chan int, 0, len(s.pDecorators)+len(s.aDecorators))
	var pCount int
	for _, d := range s.pDecorators {
		if ch, ok := d.Sync(); ok {
			columns = append(columns, ch)
			pCount++
		}
	}
	var aCount int
	for _, d := range s.aDecorators {
		if ch, ok := d.Sync(); ok {
			columns = append(columns, ch)
			aCount++
		}
	}
	table := make([][]chan int, 2)
	table[0] = columns[0:pCount]
	table[1] = columns[pCount : pCount+aCount : pCount+aCount]
	return table
}

func (b *Bar) refreshNowTillShutdown() {
	for {
		select {
		case b.forceRefresh <- time.Now():
		case <-b.shutdown:
			return
		}
	}
}

func newStatistics(s *bState) *decor.Statistics {
	return &decor.Statistics{
		ID:        s.id,
		Completed: s.completeFlushed,
		Total:     s.total,
		Current:   s.current,
	}
}

func countLines(b []byte) int {
	return bytes.Count(b, []byte("\n"))
}
