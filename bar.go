package mpb

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/mattn/go-runewidth"
	"github.com/vbauerster/mpb/v8/decor"
)

// Bar represents a progress bar.
type Bar struct {
	index        int // used by heap
	priority     int // used by heap
	frameCh      chan *renderFrame
	operateState chan func(*bState)
	done         chan struct{}
	container    *Progress
	bs           *bState
	cancel       func()
}

type extenderFunc func([]io.Reader, decor.Statistics) ([]io.Reader, error)

// bState is actual bar's state.
type bState struct {
	id                int
	priority          int
	reqWidth          int
	total             int64
	current           int64
	refill            int64
	trimSpace         bool
	completed         bool
	aborted           bool
	recovered         bool
	triggerComplete   bool
	dropOnComplete    bool
	noPop             bool
	aDecorators       []decor.Decorator
	pDecorators       []decor.Decorator
	averageDecorators []decor.AverageDecorator
	ewmaDecorators    []decor.EwmaDecorator
	shutdownListeners []decor.ShutdownListener
	buffers           [3]*bytes.Buffer
	filler            BarFiller
	middleware        func(BarFiller) BarFiller
	extender          extenderFunc
	manualRefresh     chan interface{}

	wait struct {
		bar  *Bar // key for (*pState).queueBars
		sync bool
	}
}

type renderFrame struct {
	rows     []io.Reader
	shutdown int
	err      error
}

func newBar(container *Progress, bs *bState) *Bar {
	ctx, cancel := context.WithCancel(container.ctx)

	bar := &Bar{
		priority:     bs.priority,
		frameCh:      make(chan *renderFrame, 1),
		operateState: make(chan func(*bState)),
		done:         make(chan struct{}),
		container:    container,
		cancel:       cancel,
	}

	go bar.serve(ctx, bs)
	return bar
}

// ProxyReader wraps io.Reader with metrics required for progress tracking.
// If `r` is 'unknown total/size' reader it's mandatory to call
// (*Bar).SetTotal(-1, true) method after (io.Reader).Read returns io.EOF.
// If bar is already completed or aborted, returns nil.
// Panics if `r` is nil.
func (b *Bar) ProxyReader(r io.Reader) io.ReadCloser {
	if r == nil {
		panic("expected non nil io.Reader")
	}
	result := make(chan bool)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.ewmaDecorators) != 0 }:
		return newProxyReader(r, b, <-result)
	case <-b.done:
		return nil
	}
}

// ProxyWriter wraps io.Writer with metrics required for progress tracking.
// If bar is already completed or aborted, returns nil.
// Panics if `w` is nil.
func (b *Bar) ProxyWriter(w io.Writer) io.WriteCloser {
	if w == nil {
		panic("expected non nil io.Writer")
	}
	result := make(chan bool)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.ewmaDecorators) != 0 }:
		return newProxyWriter(w, b, <-result)
	case <-b.done:
		return nil
	}
}

// ID returs id of the bar.
func (b *Bar) ID() int {
	result := make(chan int)
	select {
	case b.operateState <- func(s *bState) { result <- s.id }:
		return <-result
	case <-b.done:
		return b.bs.id
	}
}

// Current returns bar's current value, in other words sum of all increments.
func (b *Bar) Current() int64 {
	result := make(chan int64)
	select {
	case b.operateState <- func(s *bState) { result <- s.current }:
		return <-result
	case <-b.done:
		return b.bs.current
	}
}

// SetRefill sets refill flag with specified amount.
// The underlying BarFiller will change its visual representation, to
// indicate refill event. Refill event may be referred to some retry
// operation for example.
func (b *Bar) SetRefill(amount int64) {
	select {
	case b.operateState <- func(s *bState) {
		s.refill = amount
	}:
	case <-b.done:
	}
}

// TraverseDecorators traverses all available decorators and calls cb func on each.
func (b *Bar) TraverseDecorators(cb func(decor.Decorator)) {
	sync := make(chan struct{})
	select {
	case b.operateState <- func(s *bState) {
		for _, decorators := range [][]decor.Decorator{
			s.pDecorators,
			s.aDecorators,
		} {
			for _, d := range decorators {
				cb(extractBaseDecorator(d))
			}
		}
		close(sync)
	}:
		<-sync
	case <-b.done:
	}
}

// EnableTriggerComplete enables triggering complete event. It's
// effective only for bar which was constructed with `total <= 0` and
// after total has been set with (*Bar).SetTotal(int64, false). If bar
// has been incremented to the total, complete event is triggered right
// away.
func (b *Bar) EnableTriggerComplete() {
	select {
	case b.operateState <- func(s *bState) {
		if s.triggerComplete || s.total <= 0 {
			return
		}
		if s.current >= s.total {
			s.current = s.total
			s.completed = true
			go b.forceRefresh(s.manualRefresh)
		} else {
			s.triggerComplete = true
		}
	}:
	case <-b.done:
	}
}

// SetTotal sets total to an arbitrary value. It's effective only for
// bar which was constructed with `total <= 0`. Setting total to negative
// value is equivalent to (*Bar).SetTotal((*Bar).Current(), bool).
// If triggerCompleteNow is true, total value is set to current and
// complete event is triggered right away.
func (b *Bar) SetTotal(total int64, triggerCompleteNow bool) {
	select {
	case b.operateState <- func(s *bState) {
		if s.triggerComplete {
			return
		}
		if total < 0 {
			s.total = s.current
		} else {
			s.total = total
		}
		if triggerCompleteNow {
			s.current = s.total
			s.completed = true
			go b.forceRefresh(s.manualRefresh)
		}
	}:
	case <-b.done:
	}
}

// SetCurrent sets progress' current to an arbitrary value.
func (b *Bar) SetCurrent(current int64) {
	if current < 0 {
		return
	}
	select {
	case b.operateState <- func(s *bState) {
		s.current = current
		if s.triggerComplete && s.current >= s.total {
			s.current = s.total
			s.completed = true
			go b.forceRefresh(s.manualRefresh)
		}
	}:
	case <-b.done:
	}
}

// Increment is a shorthand for b.IncrInt64(1).
func (b *Bar) Increment() {
	b.IncrInt64(1)
}

// IncrBy is a shorthand for b.IncrInt64(int64(n)).
func (b *Bar) IncrBy(n int) {
	b.IncrInt64(int64(n))
}

// IncrInt64 increments progress by amount of n.
func (b *Bar) IncrInt64(n int64) {
	if n <= 0 {
		return
	}
	select {
	case b.operateState <- func(s *bState) {
		s.current += n
		if s.triggerComplete && s.current >= s.total {
			s.current = s.total
			s.completed = true
			go b.forceRefresh(s.manualRefresh)
		}
	}:
	case <-b.done:
	}
}

// EwmaIncrement is a shorthand for b.EwmaIncrInt64(1, dur).
func (b *Bar) EwmaIncrement(dur time.Duration) {
	b.EwmaIncrInt64(1, dur)
}

// EwmaIncrBy is a shorthand for b.EwmaIncrInt64(int64(n), dur).
func (b *Bar) EwmaIncrBy(n int, dur time.Duration) {
	b.EwmaIncrInt64(int64(n), dur)
}

// EwmaIncrInt64 increments progress by amount of n and updates EWMA based
// decorators by dur of a single iteration.
func (b *Bar) EwmaIncrInt64(n int64, dur time.Duration) {
	if n <= 0 {
		return
	}
	select {
	case b.operateState <- func(s *bState) {
		s.ewmaUpdate(n, dur)
		s.current += n
		if s.triggerComplete && s.current >= s.total {
			s.current = s.total
			s.completed = true
			go b.forceRefresh(s.manualRefresh)
		}
	}:
	case <-b.done:
	}
}

// DecoratorAverageAdjust adjusts all average based decorators. Call
// if you need to adjust start time of all average based decorators
// or after progress resume.
func (b *Bar) DecoratorAverageAdjust(start time.Time) {
	select {
	case b.operateState <- func(s *bState) {
		s.decoratorAverageAdjust(start)
	}:
	case <-b.done:
	}
}

// SetPriority changes bar's order among multiple bars. Zero is highest
// priority, i.e. bar will be on top. If you don't need to set priority
// dynamically, better use BarPriority option.
func (b *Bar) SetPriority(priority int) {
	b.container.UpdateBarPriority(b, priority)
}

// Abort interrupts bar's running goroutine. Abort won't be engaged
// if bar is already in complete state. If drop is true bar will be
// removed as well. To make sure that bar has been removed call
// (*Bar).Wait method.
func (b *Bar) Abort(drop bool) {
	select {
	case b.operateState <- func(s *bState) {
		if s.completed || s.aborted {
			return
		}
		s.aborted = true
		s.dropOnComplete = drop
		go b.forceRefresh(s.manualRefresh)
	}:
	case <-b.done:
	}
}

// Aborted reports whether the bar is in aborted state.
func (b *Bar) Aborted() bool {
	result := make(chan bool)
	select {
	case b.operateState <- func(s *bState) { result <- s.aborted }:
		return <-result
	case <-b.done:
		return b.bs.aborted
	}
}

// Completed reports whether the bar is in completed state.
func (b *Bar) Completed() bool {
	result := make(chan bool)
	select {
	case b.operateState <- func(s *bState) { result <- s.completed }:
		return <-result
	case <-b.done:
		return b.bs.completed
	}
}

// Wait blocks until bar is completed or aborted.
func (b *Bar) Wait() {
	<-b.done
}

func (b *Bar) serve(ctx context.Context, bs *bState) {
	defer b.container.bwg.Done()
	if bs.wait.bar != nil && bs.wait.sync {
		bs.wait.bar.Wait()
	}
	for {
		select {
		case op := <-b.operateState:
			op(bs)
		case <-ctx.Done():
			bs.aborted = !bs.completed
			bs.decoratorShutdownNotify()
			b.bs = bs
			close(b.done)
			return
		}
	}
}

func (b *Bar) render(tw int) {
	var done bool
	fn := func(s *bState) {
		var rows []io.Reader
		stat := newStatistics(tw, s)
		r, err := s.draw(stat)
		if err != nil {
			b.frameCh <- &renderFrame{err: err}
			return
		}
		rows = append(rows, r)
		if s.extender != nil {
			rows, err = s.extender(rows, stat)
			if err != nil {
				b.frameCh <- &renderFrame{err: err}
				return
			}
		}
		frame := &renderFrame{rows: rows}
		if !done && (s.completed || s.aborted) {
			frame.shutdown++
			b.cancel()
		}
		b.frameCh <- frame
	}
	select {
	case b.operateState <- fn:
	case <-b.done:
		done = true
		fn(b.bs)
	}
}

func (b *Bar) forceRefresh(refreshCh chan interface{}) {
	var anyOtherRunning bool
	b.container.traverseBars(func(bar *Bar) bool {
		anyOtherRunning = b != bar && bar.IsRunning()
		return !anyOtherRunning
	})
	if !anyOtherRunning {
		for {
			select {
			case refreshCh <- time.Now():
				time.Sleep(prr)
			case <-b.done:
				return
			}
		}
	}
}

// IsRunning reports whether the bar is running, i.e. not yet completed
// and not yet aborted.
func (b *Bar) IsRunning() bool {
	result := make(chan bool)
	select {
	case b.operateState <- func(s *bState) {
		result <- !s.completed && !s.aborted
	}:
		return <-result
	case <-b.done:
		return false
	}
}

func (b *Bar) wSyncTable() [][]chan int {
	result := make(chan [][]chan int)
	select {
	case b.operateState <- func(s *bState) { result <- s.wSyncTable() }:
		return <-result
	case <-b.done:
		return b.bs.wSyncTable()
	}
}

func (s *bState) draw(stat decor.Statistics) (io.Reader, error) {
	r, err := s.drawImpl(stat)
	if err != nil {
		for _, b := range s.buffers {
			b.Reset()
		}
		return nil, err
	}
	return io.MultiReader(r, strings.NewReader("\n")), nil
}

func (s *bState) drawImpl(stat decor.Statistics) (r io.Reader, err error) {
	type decorResult struct {
		width    int
		truncate bool
	}
	decorFiller := func(buf *bytes.Buffer, decorators []decor.Decorator) (res decorResult, err error) {
		res.width = stat.AvailableWidth
		for _, d := range decorators {
			str := d.Decor(stat)
			if stat.AvailableWidth > 0 && err == nil {
				stat.AvailableWidth -= runewidth.StringWidth(stripansi.Strip(str))
				_, err = buf.WriteString(str)
			}
		}
		res.truncate = stat.AvailableWidth < 0
		return res, err
	}
	bufP, bufB, bufA := s.buffers[0], s.buffers[1], s.buffers[2]

	resP, err := decorFiller(bufP, s.pDecorators)
	if err != nil {
		return nil, err
	}

	resA, err := decorFiller(bufA, s.aDecorators)
	if err != nil {
		return nil, err
	}

	if resP.truncate {
		trunc := strings.NewReader(runewidth.Truncate(stripansi.Strip(bufP.String()), resP.width, "…"))
		bufP.Reset()
		bufA.Reset()
		return trunc, nil
	}

	if resA.truncate {
		trunc := strings.NewReader(runewidth.Truncate(stripansi.Strip(bufA.String()), resA.width, "…"))
		bufA.Reset()
		return io.MultiReader(bufP, trunc), nil
	}

	if !s.trimSpace && stat.AvailableWidth >= 2 {
		stat.AvailableWidth -= 2
		writeFiller := func(buf *bytes.Buffer) error {
			return s.filler.Fill(buf, stat)
		}
		for _, fn := range []func(*bytes.Buffer) error{
			writeSpace,
			writeFiller,
			writeSpace,
		} {
			if err := fn(bufB); err != nil {
				return nil, err
			}
		}
	} else {
		err = s.filler.Fill(bufB, stat)
		if err != nil {
			return nil, err
		}
	}

	return io.MultiReader(bufP, bufB, bufA), nil
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

func (s *bState) subscribeDecorators() {
	for _, decorators := range [...][]decor.Decorator{
		s.pDecorators,
		s.aDecorators,
	} {
		for _, d := range decorators {
			d = extractBaseDecorator(d)
			if d, ok := d.(decor.AverageDecorator); ok {
				s.averageDecorators = append(s.averageDecorators, d)
			}
			if d, ok := d.(decor.EwmaDecorator); ok {
				s.ewmaDecorators = append(s.ewmaDecorators, d)
			}
			if d, ok := d.(decor.ShutdownListener); ok {
				s.shutdownListeners = append(s.shutdownListeners, d)
			}
		}
	}
}

func (s bState) ewmaUpdate(n int64, dur time.Duration) {
	var wg sync.WaitGroup
	for i := 0; i < len(s.ewmaDecorators); i++ {
		switch d := s.ewmaDecorators[i]; i {
		case len(s.ewmaDecorators) - 1:
			d.EwmaUpdate(n, dur)
		default:
			wg.Add(1)
			go func() {
				d.EwmaUpdate(n, dur)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func (s bState) decoratorAverageAdjust(start time.Time) {
	var wg sync.WaitGroup
	for i := 0; i < len(s.averageDecorators); i++ {
		switch d := s.averageDecorators[i]; i {
		case len(s.averageDecorators) - 1:
			d.AverageAdjust(start)
		default:
			wg.Add(1)
			go func() {
				d.AverageAdjust(start)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func (s bState) decoratorShutdownNotify() {
	var wg sync.WaitGroup
	for i := 0; i < len(s.shutdownListeners); i++ {
		switch d := s.shutdownListeners[i]; i {
		case len(s.shutdownListeners) - 1:
			d.Shutdown()
		default:
			wg.Add(1)
			go func() {
				d.Shutdown()
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func newStatistics(tw int, s *bState) decor.Statistics {
	return decor.Statistics{
		AvailableWidth: tw,
		RequestedWidth: s.reqWidth,
		ID:             s.id,
		Total:          s.total,
		Current:        s.current,
		Refill:         s.refill,
		Completed:      s.completed,
		Aborted:        s.aborted,
	}
}

func extractBaseDecorator(d decor.Decorator) decor.Decorator {
	if d, ok := d.(decor.Wrapper); ok {
		return extractBaseDecorator(d.Base())
	}
	return d
}

func writeSpace(buf *bytes.Buffer) error {
	return buf.WriteByte(' ')
}
