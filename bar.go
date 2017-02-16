package mpb

import (
	"context"
	"io"
	"sync"
	"time"
	"unicode/utf8"
)

type formatRunes [5]rune

// Bar represents a progress Bar
type Bar struct {
	formatElementCh chan runeFormatElement

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

// Statistics represents statistics of the progress bar.
// Cantains: Total, Current, TimeElapsed, TimePerItemEstimate
type Statistics struct {
	Total, Current                   int64
	TimeElapsed, TimePerItemEstimate time.Duration
}

// Eta returns exponential-weighted-moving-average ETA estimator
func (s *Statistics) Eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

type (
	runeFormatElement struct {
		char  rune
		index uint8
	}
	refill struct {
		char rune
		till int64
	}
	state struct {
		id             int64
		format         formatRunes
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

func newBar(ctx context.Context, wg *sync.WaitGroup, id, total int64, width int, format string) *Bar {
	b := &Bar{
		formatElementCh: make(chan runeFormatElement),
		etaAlphaCh:      make(chan float64),
		incrCh:          make(chan int64, 1),
		widthCh:         make(chan int),
		trimLeftCh:      make(chan bool),
		trimRightCh:     make(chan bool),
		refillCh:        make(chan *refill),
		stateReqCh:      make(chan chan state, 1),
		decoratorCh:     make(chan *decorator),
		flushedCh:       make(chan struct{}, 1),
		removeReqCh:     make(chan struct{}),
		completeReqCh:   make(chan struct{}),
		done:            make(chan struct{}),
	}
	go b.server(ctx, wg, id, total, width, format)
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
func (b *Bar) SetFill(r rune) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.formatElementCh <- runeFormatElement{r, 1}
	return b
}

// SetTip sets character representing tip of progress.
// Defaults to '>'
func (b *Bar) SetTip(r rune) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.formatElementCh <- runeFormatElement{r, 2}
	return b
}

// SetEmpty sets character representing the empty progress
// Defaults to '-'
func (b *Bar) SetEmpty(r rune) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.formatElementCh <- runeFormatElement{r, 3}
	return b
}

// SetLeftEnd sets character representing the left most border
// Defaults to '['
func (b *Bar) SetLeftEnd(r rune) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.formatElementCh <- runeFormatElement{r, 0}
	return b
}

// SetRightEnd sets character representing the right most border
// Defaults to ']'
func (b *Bar) SetRightEnd(r rune) *Bar {
	if IsClosed(b.done) {
		return b
	}
	b.formatElementCh <- runeFormatElement{r, 4}
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
func (b *Bar) IncrWithReFill(n int, r rune) {
	if IsClosed(b.done) {
		return
	}
	b.Incr(n)
	b.refillCh <- &refill{r, int64(n)}
}

// GetAppenders returns slice of appender DecoratorFunc
func (b *Bar) GetAppenders() []DecoratorFunc {
	s := b.getState()
	return s.appendFuncs
}

// GetPrependers returns slice of prepender DecoratorFunc
func (b *Bar) GetPrependers() []DecoratorFunc {
	s := b.getState()
	return s.prependFuncs
}

// GetStatistics returs *Statistics, which contains information like Tottal,
// Current, TimeElapsed and TimePerItemEstimate
func (b *Bar) GetStatistics() *Statistics {
	state := b.getState()
	return state.newStat()
}

// GetID returs id of the bar
func (b *Bar) GetID() int64 {
	state := b.getState()
	return state.id
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

func (b *Bar) getState() state {
	if IsClosed(b.done) {
		return b.lastState
	}
	ch := make(chan state, 1)
	b.stateReqCh <- ch
	return <-ch
}

func (b *Bar) bytes(termWidth int) []byte {
	state := b.getState()
	return state.draw(termWidth)
}

func (b *Bar) server(ctx context.Context, wg *sync.WaitGroup, id, total int64, width int, format string) {
	var completed bool
	timeStarted := time.Now()
	blockStartTime := timeStarted
	state := state{
		id:       id,
		format:   formatRunes{'[', '=', '>', '-', ']'},
		etaAlpha: 0.25,
		barWidth: width,
		total:    total,
	}
	if total <= 0 {
		state.simpleSpinner = getSpinner()
	} else if format != "" {
		for i, n := 0, 0; len(format) > 0; i++ {
			state.format[i], n = utf8.DecodeRuneInString(format)
			format = format[n:]
		}
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
		case e := <-b.formatElementCh:
			state.format[e.index] = e.char
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

func (s *state) newStat() *Statistics {
	return &Statistics{
		Total:               s.total,
		Current:             s.current,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}
}

func (s *state) draw(termWidth int) []byte {
	if termWidth <= 0 {
		termWidth = s.barWidth
	}

	stat := s.newStat()

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
	if totalCount > termWidth {
		newWidth := termWidth - prependCount - appendCount
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

	// bar width without leftEnd and rightEnd characters
	barWidth := width - 1
	formatBytes := s.convertFormatRunesToBytes()

	if s.simpleSpinner != nil {
		var buf []byte
		for _, block := range [...][]byte{formatBytes[0], []byte{s.simpleSpinner()}, formatBytes[4]} {
			buf = append(buf, block...)
		}
		return buf
	}

	completedWidth := percentage(s.total, s.current, barWidth)

	buf := make([]byte, 0, width)
	// append leftEnd rune
	buf = append(buf, formatBytes[0]...)

	if s.refill != nil {
		till := percentage(s.total, s.refill.till, barWidth)
		rbytes := make([]byte, utf8.RuneLen(s.refill.char))
		utf8.EncodeRune(rbytes, s.refill.char)
		for i := 0; i < till; i++ {
			buf = append(buf, rbytes...)
		}
		for i := till; i < completedWidth-1; i++ {
			// append fill rune
			buf = append(buf, formatBytes[1]...)
		}
	} else {
		for i := 0; i < completedWidth-1; i++ {
			// append fill rune
			buf = append(buf, formatBytes[1]...)
		}
	}
	// set tip bit
	if completedWidth > 0 && completedWidth < barWidth {
		buf = append(buf, formatBytes[2]...)
	}

	for i := completedWidth + 1; i < barWidth; i++ {
		// append empty rune
		buf = append(buf, formatBytes[3]...)
	}
	// append rightEnd rune
	buf = append(buf, formatBytes[4]...)
	return buf
}

func (s *state) convertFormatRunesToBytes() [5][]byte {
	var formatBytes [5][]byte
	for i, r := range s.format {
		buf := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(buf, r)
		formatBytes[i] = buf
	}
	return formatBytes
}

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
