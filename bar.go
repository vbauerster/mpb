package mpb

import (
	"fmt"
	"io"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/decor"
)

const (
	rLeft = iota
	rFill
	rTip
	rEmpty
	rRight
)

const (
	formatLen = 5
	etaAlpha  = 0.25
)

type fmtRunes [formatLen]rune
type fmtByteSegments [formatLen][]byte

// Bar represents a progress Bar
type Bar struct {
	// quit channel to request b.server to quit
	quit chan struct{}
	// done channel is receiveable after b.server has been quit
	done chan struct{}
	ops  chan func(*state)

	// following are used after b.done is receiveable
	cacheState state
}

type (
	refill struct {
		char rune
		till int64
	}
	state struct {
		id             int
		width          int
		format         fmtRunes
		etaAlpha       float64
		total          int64
		current        int64
		trimLeftSpace  bool
		trimRightSpace bool
		completed      bool
		aborted        bool
		startTime      time.Time
		timeElapsed    time.Duration
		blockStartTime time.Time
		timePerItem    time.Duration
		appendFuncs    []decor.DecoratorFunc
		prependFuncs   []decor.DecoratorFunc
		simpleSpinner  func() byte
		refill         *refill
	}
)

func newBar(total int64, wg *sync.WaitGroup, cancel <-chan struct{}, options ...BarOption) *Bar {
	s := state{
		total:    total,
		etaAlpha: etaAlpha,
	}

	if total <= 0 {
		s.simpleSpinner = getSpinner()
	}

	for _, opt := range options {
		opt(&s)
	}

	b := &Bar{
		quit: make(chan struct{}),
		done: make(chan struct{}),
		ops:  make(chan func(*state)),
	}

	go b.server(s, wg, cancel)
	return b
}

// RemoveAllPrependers removes all prepend functions
func (b *Bar) RemoveAllPrependers() {
	select {
	case b.ops <- func(s *state) {
		s.prependFuncs = nil
	}:
	case <-b.quit:
		return
	}
}

// RemoveAllAppenders removes all append functions
func (b *Bar) RemoveAllAppenders() {
	select {
	case b.ops <- func(s *state) {
		s.appendFuncs = nil
	}:
	case <-b.quit:
		return
	}
}

// ProxyReader wrapper for io operations, like io.Copy
func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Increment shorthand for b.Incr(1)
func (b *Bar) Increment() {
	b.Incr(1)
}

// Incr increments progress bar
func (b *Bar) Incr(n int) {
	if n < 1 {
		return
	}
	select {
	case b.ops <- func(s *state) {
		if s.current == 0 {
			s.startTime = time.Now()
			s.blockStartTime = s.startTime
		}
		sum := s.current + int64(n)
		s.timeElapsed = time.Since(s.startTime)
		s.updateTimePerItemEstimate(n)
		if s.total > 0 && sum >= s.total {
			s.current = s.total
			s.completed = true
			return
		}
		s.current = sum
		s.blockStartTime = time.Now()
	}:
	case <-b.quit:
		return
	}
}

// ResumeFill fills bar with different r rune,
// from 0 to till amount of progress.
func (b *Bar) ResumeFill(r rune, till int64) {
	if till < 1 {
		return
	}
	select {
	case b.ops <- func(s *state) {
		s.refill = &refill{r, till}
	}:
	case <-b.quit:
		return
	}
}

func (b *Bar) NumOfAppenders() int {
	result := make(chan int, 1)
	select {
	case b.ops <- func(s *state) { result <- len(s.appendFuncs) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.appendFuncs)
	}
}

func (b *Bar) NumOfPrependers() int {
	result := make(chan int, 1)
	select {
	case b.ops <- func(s *state) { result <- len(s.prependFuncs) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.prependFuncs)
	}
}

// ID returs id of the bar
func (b *Bar) ID() int {
	result := make(chan int, 1)
	select {
	case b.ops <- func(s *state) { result <- s.id }:
		return <-result
	case <-b.done:
		return b.cacheState.id
	}
}

func (b *Bar) Current() int64 {
	result := make(chan int64, 1)
	select {
	case b.ops <- func(s *state) { result <- s.current }:
		return <-result
	case <-b.done:
		return b.cacheState.current
	}
}

func (b *Bar) Total() int64 {
	result := make(chan int64, 1)
	select {
	case b.ops <- func(s *state) { result <- s.total }:
		return <-result
	case <-b.done:
		return b.cacheState.total
	}
}

// InProgress returns true, while progress is running.
// Can be used as condition in for loop
func (b *Bar) InProgress() bool {
	select {
	case <-b.quit:
		return false
	default:
		return true
	}
}

// Complete signals to the bar, that process has been completed.
// You should call this method when total is unknown and you've reached the point
// of process completion. If you don't call this method, it will be called
// implicitly, upon p.Stop() call.
func (b *Bar) Complete() {
	select {
	case <-b.quit:
	default:
		close(b.quit)
	}
}

func (b *Bar) complete() {
	select {
	case b.ops <- func(s *state) {
		if !s.completed {
			b.Complete()
		}
	}:
	case <-time.After(prr):
		return
	}
}

func (b *Bar) server(s state, wg *sync.WaitGroup, cancel <-chan struct{}) {

	defer func() {
		b.cacheState = s
		close(b.done)
		wg.Done()
	}()

	for {
		select {
		case op := <-b.ops:
			op(&s)
		case <-b.quit:
			s.completed = true
			return
		case <-cancel:
			s.aborted = true
			cancel = nil
			b.Complete()
		}
	}
}

func (b *Bar) render(tw int, flushed chan struct{}, prependWs, appendWs *widthSync) <-chan []byte {
	ch := make(chan []byte, 1)

	go func() {
		defer func() {
			// recovering if external decorators panic
			if p := recover(); p != nil {
				ch <- []byte(fmt.Sprintln(p))
			}
			close(ch)
		}()
		var st state
		result := make(chan state, 1)
		select {
		case b.ops <- func(s *state) {
			result <- *s
			if s.completed {
				<-flushed
				b.Complete()
			}
		}:
			st = <-result
		case <-b.done:
			st = b.cacheState
		}
		buf := draw(&st, tw, prependWs, appendWs)
		buf = append(buf, '\n')
		ch <- buf
	}()

	return ch
}

func (s *state) updateFormat(format string) {
	for i, n := 0, 0; len(format) > 0; i++ {
		s.format[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
}

func (s *state) updateTimePerItemEstimate(amount int) {
	lastBlockTime := time.Since(s.blockStartTime) // shorthand for time.Now().Sub(t)
	lastItemEstimate := float64(lastBlockTime) / float64(amount)
	s.timePerItem = time.Duration((s.etaAlpha * lastItemEstimate) + (1-s.etaAlpha)*float64(s.timePerItem))
}

func draw(s *state, termWidth int, prependWs, appendWs *widthSync) []byte {
	if len(s.prependFuncs) != len(prependWs.Listen) || len(s.appendFuncs) != len(appendWs.Listen) {
		return []byte{}
	}
	if termWidth <= 0 {
		termWidth = s.width
	}

	stat := newStatistics(s)

	// render prepend functions to the left of the bar
	var prependBlock []byte
	for i, f := range s.prependFuncs {
		prependBlock = append(prependBlock,
			[]byte(f(stat, prependWs.Listen[i], prependWs.Result[i]))...)
	}

	// render append functions to the right of the bar
	var appendBlock []byte
	for i, f := range s.appendFuncs {
		appendBlock = append(appendBlock,
			[]byte(f(stat, appendWs.Listen[i], appendWs.Result[i]))...)
	}

	prependCount := utf8.RuneCount(prependBlock)
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

	var barBlock []byte
	buf := make([]byte, 0, termWidth)
	segments := fmtRunesToByteSegments(s.format)

	if s.simpleSpinner != nil {
		for _, block := range [...][]byte{segments[rLeft], {s.simpleSpinner()}, segments[rRight]} {
			barBlock = append(barBlock, block...)
		}
	} else {
		barBlock = fillBar(s.total, s.current, s.width, segments, s.refill)
		barCount := utf8.RuneCount(barBlock)
		totalCount := prependCount + barCount + appendCount
		if totalCount > termWidth {
			shrinkWidth := termWidth - prependCount - appendCount
			barBlock = fillBar(s.total, s.current, shrinkWidth, segments, s.refill)
		}
	}

	return concatenateBlocks(buf, prependBlock, leftSpace, barBlock, rightSpace, appendBlock)
}

func concatenateBlocks(buf []byte, blocks ...[]byte) []byte {
	for _, block := range blocks {
		buf = append(buf, block...)
	}
	return buf
}

func fillBar(total, current int64, width int, fmtBytes fmtByteSegments, rf *refill) []byte {
	if width < 2 || total <= 0 {
		return []byte{}
	}

	// bar width without leftEnd and rightEnd runes
	barWidth := width - 2

	completedWidth := decor.CalcPercentage(total, current, barWidth)

	buf := make([]byte, 0, width)
	buf = append(buf, fmtBytes[rLeft]...)

	if rf != nil {
		till := decor.CalcPercentage(total, rf.till, barWidth)
		rbytes := make([]byte, utf8.RuneLen(rf.char))
		utf8.EncodeRune(rbytes, rf.char)
		// append refill rune
		for i := 0; i < till; i++ {
			buf = append(buf, rbytes...)
		}
		for i := till; i < completedWidth; i++ {
			buf = append(buf, fmtBytes[rFill]...)
		}
	} else {
		for i := 0; i < completedWidth; i++ {
			buf = append(buf, fmtBytes[rFill]...)
		}
	}

	if completedWidth < barWidth && completedWidth > 0 {
		_, size := utf8.DecodeLastRune(buf)
		buf = buf[:len(buf)-size]
		buf = append(buf, fmtBytes[rTip]...)
	}

	for i := completedWidth; i < barWidth; i++ {
		buf = append(buf, fmtBytes[rEmpty]...)
	}

	buf = append(buf, fmtBytes[rRight]...)

	return buf
}

func newStatistics(s *state) *decor.Statistics {
	return &decor.Statistics{
		ID:                  s.id,
		Completed:           s.completed,
		Aborted:             s.aborted,
		Total:               s.total,
		Current:             s.current,
		StartTime:           s.startTime,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}
}

func fmtRunesToByteSegments(format fmtRunes) fmtByteSegments {
	var segments fmtByteSegments
	for i, r := range format {
		buf := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(buf, r)
		segments[i] = buf
	}
	return segments
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
