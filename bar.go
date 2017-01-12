package mpb

import (
	"context"
	"io"
	"sync"
	"time"
	"unicode/utf8"
)

// Bar represents a progress Bar
type Bar struct {
	width     int
	termWidth int
	etaAlpha  float64

	fill     byte
	empty    byte
	tip      byte
	leftEnd  byte
	rightEnd byte

	incrCh      chan int64
	trimLeftCh  chan bool
	trimRightCh chan bool
	stateReqCh  chan chan state
	decoratorCh chan *decorator
	flushedCh   chan struct{}
	removeReqCh chan struct{}
	done        chan struct{}

	refill    *reFill
	lastState state
}

// Statistics represents statistics of the progress bar
// instance of this, sent to DecoratorFunc, as param
type Statistics struct {
	Total, Current                   int64
	TermWidth                        int
	TimeElapsed, TimePerItemEstimate time.Duration
}

func (s *Statistics) Eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

type (
	state struct {
		total, current                int64
		timeElapsed, timePerItem      time.Duration
		appendFuncs, prependFuncs     []DecoratorFunc
		trimLeftSpace, trimRightSpace bool
	}
	reFill struct {
		c    byte
		till int64
	}
)

func newBar(ctx context.Context, wg *sync.WaitGroup, total int64, width int) *Bar {
	b := &Bar{
		fill:     '=',
		empty:    '-',
		tip:      '>',
		leftEnd:  '[',
		rightEnd: ']',
		etaAlpha: 0.25,
		width:    width,

		incrCh:      make(chan int64, 1),
		trimLeftCh:  make(chan bool),
		trimRightCh: make(chan bool),
		stateReqCh:  make(chan chan state, 1),
		decoratorCh: make(chan *decorator),
		flushedCh:   make(chan struct{}, 1),
		removeReqCh: make(chan struct{}),
		done:        make(chan struct{}),
	}
	go b.server(ctx, wg, total)
	return b
}

// SetWidth sets width of the bar
func (b *Bar) SetWidth(n int) *Bar {
	if n < 2 {
		return b
	}
	b.width = n
	return b
}

// TrimLeftSpace removes space befor LeftEnd charater
func (b *Bar) TrimLeftSpace() *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.trimLeftCh <- true
	return b
}

// TrimRightSpace removes space after RightEnd charater
func (b *Bar) TrimRightSpace() *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.trimRightCh <- true
	return b
}

// SetFill sets character representing completed progress.
// Defaults to '='
func (b *Bar) SetFill(c byte) *Bar {
	b.fill = c
	return b
}

// SetTip sets character representing tip of progress.
// Defaults to '>'
func (b *Bar) SetTip(c byte) *Bar {
	b.tip = c
	return b
}

// SetEmpty sets character representing the empty progress
// Defaults to '-'
func (b *Bar) SetEmpty(c byte) *Bar {
	b.empty = c
	return b
}

// SetLeftEnd sets character representing the left most border
// Defaults to '['
func (b *Bar) SetLeftEnd(c byte) *Bar {
	b.leftEnd = c
	return b
}

// SetRightEnd sets character representing the right most border
// Defaults to ']'
func (b *Bar) SetRightEnd(c byte) *Bar {
	b.rightEnd = c
	return b
}

// SetEtaAlpha sets alfa for exponential-weighted-moving-average ETA estimator
// Defaults to 0.25
// Normally you shouldn't touch this
func (b *Bar) SetEtaAlpha(a float64) *Bar {
	b.etaAlpha = a
	return b
}

// ProxyReader wrapper for io operations, like io.Copy
func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Incr increments progress bar
func (b *Bar) Incr(n int) {
	if IsClosed(b.done) {
		return
	}
	if n > 0 {
		b.incrCh <- int64(n)
	}
}

// IncrWithReFill increments pb with different fill character
func (b *Bar) IncrWithReFill(n int, c byte) {
	if n > 0 {
		b.Incr(n)
		b.refill = &reFill{c, int64(n)}
	}
}

// Current returns the actual current.
func (b *Bar) Current() int64 {
	if IsClosed(b.done) {
		return b.lastState.current
	}
	ch := make(chan state, 1)
	b.stateReqCh <- ch
	state := <-ch
	return state.current
}

// InProgress returns true, while progress is running
// Can be used as condition in for loop
func (b *Bar) InProgress() bool {
	return !IsClosed(b.done)
}

// PrependFunc prepends DecoratorFunc
func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.decoratorCh <- &decorator{decPrepend, f}
	return b
}

// RemoveAllPrependers removes all prepend functions
func (b *Bar) RemoveAllPrependers() {
	if IsClosed(b.done) {
		return
	}
	b.decoratorCh <- &decorator{decPrependZero, nil}
}

// AppendFunc appends DecoratorFunc
func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.decoratorCh <- &decorator{decAppend, f}
	return b
}

// RemoveAllAppenders removes all append functions
func (b *Bar) RemoveAllAppenders() {
	if IsClosed(b.done) {
		return
	}
	b.decoratorCh <- &decorator{decAppendZero, nil}
}

func (b *Bar) bytes(width int) []byte {
	if width <= 0 {
		width = b.width
	}
	if IsClosed(b.done) {
		return b.draw(b.lastState, width)
	}
	ch := make(chan state, 1)
	b.stateReqCh <- ch
	return b.draw(<-ch, width)
}

func (b *Bar) server(ctx context.Context, wg *sync.WaitGroup, total int64) {
	var completed bool
	timeStarted := time.Now()
	blockStartTime := timeStarted
	state := state{total: total}
	defer func() {
		b.stop(&state)
		wg.Done()
	}()
	for {
		select {
		case i := <-b.incrCh:
			n := state.current + i
			if n > total {
				state.current = total
				completed = true
				blockStartTime = time.Now()
				break // break out of select
			}
			state.timeElapsed = time.Since(timeStarted)
			state.timePerItem = calcTimePerItemEstimate(state.timePerItem, blockStartTime, b.etaAlpha, i)
			if n == total {
				completed = true
			}
			state.current = n
			blockStartTime = time.Now()
		case d := <-b.decoratorCh:
			switch d.kind {
			case decAppend:
				state.appendFuncs = append(state.appendFuncs, d.f)
			case decAppendZero:
				state.appendFuncs = nil
			case decPrepend:
				state.prependFuncs = append(state.prependFuncs, d.f)
			case decPrependZero:
				state.prependFuncs = nil
			}
		case ch := <-b.stateReqCh:
			ch <- state
		case state.trimLeftSpace = <-b.trimLeftCh:
		case state.trimRightSpace = <-b.trimRightCh:
		case <-b.flushedCh:
			if completed {
				return
			}
		case <-b.removeReqCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bar) stop(s *state) {
	b.lastState = *s
	close(b.done)
}

func (b *Bar) flushed() {
	if IsClosed(b.done) {
		return
	}
	b.flushedCh <- struct{}{}
}

func (b *Bar) remove() {
	if IsClosed(b.done) {
		return
	}
	b.removeReqCh <- struct{}{}
}

func (b *Bar) draw(s state, termWidth int) []byte {
	buf := make([]byte, 0, termWidth)

	stat := &Statistics{
		Total:               s.total,
		Current:             s.current,
		TermWidth:           termWidth,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}

	barBlock := b.fillBar(s.total, s.current, b.width)

	// render append functions to the right of the bar
	var appendBlock []byte
	for _, f := range s.appendFuncs {
		appendBlock = append(appendBlock, []byte(f(stat))...)
	}

	// render prepend functions to the left of the bar
	var prependBlock []byte
	for _, f := range s.prependFuncs {
		prependBlock = append(prependBlock, []byte(f(stat))...)
	}

	prependCount := utf8.RuneCount(prependBlock)
	barCount := utf8.RuneCount(barBlock)
	appendCount := utf8.RuneCount(appendBlock)

	var leftSpace, rightSpace []byte
	space := []byte{' '}

	if !s.trimLeftSpace {
		prependCount++
		leftSpace = space
	}
	if !s.trimRightSpace {
		appendCount++
		rightSpace = space
	}

	totalCount := prependCount + barCount + appendCount
	if totalCount >= termWidth {
		newWidth := termWidth - prependCount - appendCount
		barBlock = b.fillBar(s.total, s.current, newWidth-1)
	}

	for _, block := range [...][]byte{prependBlock, leftSpace, barBlock, rightSpace, appendBlock} {
		buf = append(buf, block...)
	}

	return buf
}

func (b *Bar) fillBar(total, current int64, width int) []byte {
	if width < 2 {
		return []byte{}
	}

	buf := make([]byte, width)
	completedWidth := percentage(total, current, width)

	if b.refill != nil {
		till := percentage(total, b.refill.till, width)
		for i := 1; i < till; i++ {
			buf[i] = b.refill.c
		}
		for i := till; i < completedWidth; i++ {
			buf[i] = b.fill
		}
	} else {
		for i := 1; i < completedWidth; i++ {
			buf[i] = b.fill
		}
	}

	for i := completedWidth; i < width-1; i++ {
		buf[i] = b.empty
	}
	// set tip bit
	if completedWidth > 0 && completedWidth < width {
		buf[completedWidth-1] = b.tip
	}
	// set left and right ends bits
	buf[0], buf[width-1] = b.leftEnd, b.rightEnd

	return buf
}

func (b *Bar) status() int {
	var total, current int64
	if IsClosed(b.done) {
		total = b.lastState.total
		current = b.lastState.current
	} else {
		ch := make(chan state)
		b.stateReqCh <- ch
		state := <-ch
		total = state.total
		current = state.current
	}
	return percentage(total, current, 100)
}

// SortableBarSlice satisfies sort interface
type SortableBarSlice []*Bar

func (p SortableBarSlice) Len() int { return len(p) }

func (p SortableBarSlice) Less(i, j int) bool { return p[i].status() < p[j].status() }

func (p SortableBarSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func calcTimePerItemEstimate(tpie time.Duration, blockStartTime time.Time, alpha float64, items int64) time.Duration {
	lastBlockTime := time.Since(blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(items)
	return time.Duration((alpha * lastItemEstimate) + (1-alpha)*float64(tpie))
}

func percentage(total, current int64, ratio int) int {
	if total == 0 {
		return 0
	}
	return int(float64(ratio) * float64(current) / float64(total))
}
