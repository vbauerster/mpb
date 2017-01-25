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
	fillCh, emptyCh, tipCh, leftEndCh, rightEndCh chan byte

	widthCh    chan int
	etaAlphaCh chan float64

	incrCh        chan int64
	trimLeftCh    chan bool
	trimRightCh   chan bool
	refillCh      chan *refill
	stateReqCh    chan chan state
	decoratorCh   chan *decorator
	flushedCh     chan struct{}
	removeReqCh   chan struct{}
	completeReqCh chan struct{}
	done          chan struct{}

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
	refill struct {
		char byte
		till int64
	}
	state struct {
		fill           byte
		empty          byte
		tip            byte
		leftEnd        byte
		rightEnd       byte
		etaAlpha       float64
		barWidth       int
		total          int64
		current        int64
		trimLeftSpace  bool
		trimRightSpace bool
		timeElapsed    time.Duration
		timePerItem    time.Duration
		appendFuncs    []DecoratorFunc
		prependFuncs   []DecoratorFunc
		simpleSpinner  func() byte
		refill         *refill
	}
)

func newBar(ctx context.Context, wg *sync.WaitGroup, total int64, barWidth int) *Bar {
	b := &Bar{
		fillCh:        make(chan byte),
		emptyCh:       make(chan byte),
		tipCh:         make(chan byte),
		leftEndCh:     make(chan byte),
		rightEndCh:    make(chan byte),
		etaAlphaCh:    make(chan float64),
		incrCh:        make(chan int64, 1),
		widthCh:       make(chan int),
		trimLeftCh:    make(chan bool),
		trimRightCh:   make(chan bool),
		refillCh:      make(chan *refill),
		stateReqCh:    make(chan chan state, 1),
		decoratorCh:   make(chan *decorator),
		flushedCh:     make(chan struct{}, 1),
		removeReqCh:   make(chan struct{}),
		completeReqCh: make(chan struct{}),
		done:          make(chan struct{}),
	}
	go b.server(ctx, wg, total, barWidth)
	return b
}

// SetWidth sets width of the bar
func (b *Bar) SetWidth(n int) *Bar {
	if n < 2 || IsClosed(b.done) {
		return b
	}
	b.widthCh <- n
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
	if IsClosed(b.done) {
		return b
	}
	b.fillCh <- c
	return b
}

// SetTip sets character representing tip of progress.
// Defaults to '>'
func (b *Bar) SetTip(c byte) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.tipCh <- c
	return b
}

// SetEmpty sets character representing the empty progress
// Defaults to '-'
func (b *Bar) SetEmpty(c byte) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.emptyCh <- c
	return b
}

// SetLeftEnd sets character representing the left most border
// Defaults to '['
func (b *Bar) SetLeftEnd(c byte) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.leftEndCh <- c
	return b
}

// SetRightEnd sets character representing the right most border
// Defaults to ']'
func (b *Bar) SetRightEnd(c byte) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.rightEndCh <- c
	return b
}

// SetEtaAlpha sets alfa for exponential-weighted-moving-average ETA estimator
// Defaults to 0.25
// Normally you shouldn't touch this
func (b *Bar) SetEtaAlpha(a float64) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.etaAlphaCh <- a
	return b
}

// ProxyReader wrapper for io operations, like io.Copy
func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Incr increments progress bar
func (b *Bar) Incr(n int) {
	if n < 1 || IsClosed(b.done) {
		return
	}
	b.incrCh <- int64(n)
}

// IncrWithReFill increments pb with different fill character
func (b *Bar) IncrWithReFill(n int, c byte) {
	if IsClosed(b.done) {
		return
	}
	b.Incr(n)
	b.refillCh <- &refill{c, int64(n)}
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

// Completed signals to the bar, that process has been completed.
// You should call this method when total is unknown and you've reached the point
// of process completion.
func (b *Bar) Completed() {
	if IsClosed(b.done) {
		return
	}
	b.completeReqCh <- struct{}{}
}

func (b *Bar) bytes(termWidth int) []byte {
	if IsClosed(b.done) {
		return b.lastState.draw(termWidth)
	}
	ch := make(chan state, 1)
	b.stateReqCh <- ch
	s := <-ch
	return s.draw(termWidth)
}

func (b *Bar) server(ctx context.Context, wg *sync.WaitGroup, total int64, barWidth int) {
	var completed bool
	timeStarted := time.Now()
	blockStartTime := timeStarted
	state := state{
		fill:     '=',
		empty:    '-',
		tip:      '>',
		leftEnd:  '[',
		rightEnd: ']',
		etaAlpha: 0.25,
		barWidth: barWidth,
		total:    total,
	}
	if total <= 0 {
		state.simpleSpinner = getSpinner()
	}
	defer func() {
		b.stop(&state)
		wg.Done()
	}()
	for {
		select {
		case i := <-b.incrCh:
			n := state.current + i
			if total > 0 && n > total {
				state.current = total
				completed = true
				blockStartTime = time.Now()
				break // break out of select
			}
			state.timeElapsed = time.Since(timeStarted)
			state.timePerItem = calcTimePerItemEstimate(state.timePerItem, blockStartTime, state.etaAlpha, i)
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
		case state.fill = <-b.fillCh:
		case state.empty = <-b.emptyCh:
		case state.tip = <-b.tipCh:
		case state.leftEnd = <-b.leftEndCh:
		case state.rightEnd = <-b.rightEndCh:
		case state.barWidth = <-b.widthCh:
		case state.refill = <-b.refillCh:
		case state.trimLeftSpace = <-b.trimLeftCh:
		case state.trimRightSpace = <-b.trimRightCh:
		case <-b.flushedCh:
			if completed {
				return
			}
		case <-b.completeReqCh:
			return
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

func (s *state) draw(termWidth int) []byte {
	if termWidth <= 0 {
		termWidth = s.barWidth
	}
	stat := &Statistics{
		Total:               s.total,
		Current:             s.current,
		TermWidth:           termWidth,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}

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

	barBlock := s.fillBar(s.barWidth)
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
		newWidth := termWidth - prependCount - appendCount - 1
		barBlock = s.fillBar(newWidth)
	}

	buf := make([]byte, 0, termWidth)
	for _, block := range [...][]byte{prependBlock, leftSpace, barBlock, rightSpace, appendBlock} {
		buf = append(buf, block...)
	}

	return buf
}

func (s *state) fillBar(width int) []byte {
	if width < 2 {
		return []byte{}
	}

	if s.simpleSpinner != nil {
		return []byte{s.leftEnd, s.simpleSpinner(), s.rightEnd}
	}

	buf := make([]byte, width)
	completedWidth := percentage(s.total, s.current, width)

	if s.refill != nil {
		till := percentage(s.total, s.refill.till, width)
		for i := 1; i < till; i++ {
			buf[i] = s.refill.char
		}
		for i := till; i < completedWidth; i++ {
			buf[i] = s.fill
		}
	} else {
		for i := 1; i < completedWidth; i++ {
			buf[i] = s.fill
		}
	}

	for i := completedWidth; i < width-1; i++ {
		buf[i] = s.empty
	}
	// set tip bit
	if completedWidth > 0 && completedWidth < s.barWidth {
		buf[completedWidth-1] = s.tip
	}
	// set left and right ends bits
	buf[0], buf[width-1] = s.leftEnd, s.rightEnd

	return buf
}

func (b *Bar) status() int {
	var total, current int64
	if IsClosed(b.done) {
		total = b.lastState.total
		current = b.lastState.current
	} else {
		ch := make(chan state, 1)
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
	if total <= 0 {
		return 0
	}
	return int(float64(ratio) * float64(current) / float64(total))
}

func getSpinner() func() byte {
	chars := []byte(`-\|/`)
	repeat := len(chars) - 1
	index := repeat
	return func() byte {
		if index == repeat {
			index = -1
		}
		index++
		return chars[index]
	}
}
