package mpb

import (
	"bytes"
	"fmt"
	"io"
	"strings"
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

// Bar represents a progress Bar
type Bar struct {
	priority int
	index    int

	// the flag is set from Progress monitor goroutine only
	completed bool

	operateState chan func(*bState)
	done         chan struct{}
	shutdown     chan struct{}

	// it's guaranteed that cacheState isn't nil, after done is closed
	cacheState *bState
}

type (
	bState struct {
		id                   int
		width                int
		format               fmtRunes
		etaAlpha             float64
		total                int64
		current              int64
		totalAutoIncrTrigger int64
		totalAutoIncrBy      int64
		trimLeftSpace        bool
		trimRightSpace       bool
		completed            bool
		removed              bool
		dynamic              bool
		startTime            time.Time
		timeElapsed          time.Duration
		blockStartTime       time.Time
		timePerItem          time.Duration
		aDecorators          []decor.DecoratorFunc
		pDecorators          []decor.DecoratorFunc
		refill               *refill
		bufP, bufB, bufA     *bytes.Buffer
		panicMsg             string
	}
	refill struct {
		char rune
		till int64
	}
	renderedReader struct {
		io.Reader
		toComplete bool
		toRemove   bool
	}
)

func newBar(id int, total int64, cancel <-chan struct{}, options ...BarOption) *Bar {
	if total <= 0 {
		total = time.Now().Unix()
	}

	s := &bState{
		id:       id,
		total:    total,
		etaAlpha: etaAlpha,
	}

	for _, opt := range options {
		opt(s)
	}

	s.bufP = bytes.NewBuffer(make([]byte, 0, s.width/2))
	s.bufB = bytes.NewBuffer(make([]byte, 0, s.width))
	s.bufA = bytes.NewBuffer(make([]byte, 0, s.width/2))

	b := &Bar{
		priority:     id,
		operateState: make(chan func(*bState)),
		done:         make(chan struct{}),
		shutdown:     make(chan struct{}),
	}

	go b.serve(s, cancel)
	return b
}

// RemoveAllPrependers removes all prepend functions
func (b *Bar) RemoveAllPrependers() {
	select {
	case b.operateState <- func(s *bState) { s.pDecorators = nil }:
	case <-b.done:
	}
}

// RemoveAllAppenders removes all append functions
func (b *Bar) RemoveAllAppenders() {
	select {
	case b.operateState <- func(s *bState) { s.aDecorators = nil }:
	case <-b.done:
	}
}

// ProxyReader wrapper for io operations, like io.Copy
func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Increment is a shorthand for b.IncrBy(1)
func (b *Bar) Increment() {
	b.IncrBy(1)
}

// IncrBy increments progress bar by amount of n
func (b *Bar) IncrBy(n int) {
	if n < 1 {
		return
	}
	select {
	case b.operateState <- func(s *bState) {
		if s.completed {
			return
		}
		next := time.Now()
		if s.current == 0 {
			s.startTime = next
			s.blockStartTime = next
		} else {
			now := time.Now()
			s.updateTimePerItemEstimate(n, now, next)
			s.timeElapsed = now.Sub(s.startTime)
		}
		s.current += int64(n)
		if s.dynamic {
			curp := decor.CalcPercentage(s.total, s.current, 100)
			if 100-curp <= s.totalAutoIncrTrigger {
				s.total += s.totalAutoIncrBy
			}
		} else if s.current >= s.total {
			s.current = s.total
			s.completed = true
		}
	}:
	case <-b.done:
	}
}

// ResumeFill fills bar with different r rune,
// from 0 to till amount of progress.
func (b *Bar) ResumeFill(r rune, till int64) {
	if till < 1 {
		return
	}
	select {
	case b.operateState <- func(s *bState) { s.refill = &refill{r, till} }:
	case <-b.done:
	}
}

// NumOfAppenders returns current number of append decorators
func (b *Bar) NumOfAppenders() int {
	result := make(chan int, 1)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.aDecorators) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.aDecorators)
	}
}

// NumOfPrependers returns current number of prepend decorators
func (b *Bar) NumOfPrependers() int {
	result := make(chan int, 1)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.pDecorators) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.pDecorators)
	}
}

// ID returs id of the bar
func (b *Bar) ID() int {
	result := make(chan int, 1)
	select {
	case b.operateState <- func(s *bState) { result <- s.id }:
		return <-result
	case <-b.done:
		return b.cacheState.id
	}
}

// Current returns bar's current number, in other words sum of all increments.
func (b *Bar) Current() int64 {
	result := make(chan int64, 1)
	select {
	case b.operateState <- func(s *bState) { result <- s.current }:
		return <-result
	case <-b.done:
		return b.cacheState.current
	}
}

// Total returns bar's total number.
func (b *Bar) Total() int64 {
	result := make(chan int64, 1)
	select {
	case b.operateState <- func(s *bState) { result <- s.total }:
		return <-result
	case <-b.done:
		return b.cacheState.total
	}
}

// SetTotal sets total dynamically. The final param indicates the very last set,
// in other words you should set it to true when total is determined.
func (b *Bar) SetTotal(total int64, final bool) {
	select {
	case b.operateState <- func(s *bState) {
		s.total = total
		s.dynamic = !final
	}:
	case <-b.done:
	}
}

// Completed reports whether the bar is in completed state
func (b *Bar) Completed() bool {
	result := make(chan bool, 1)
	select {
	case b.operateState <- func(s *bState) { result <- s.completed }:
		return <-result
	case <-b.done:
		return b.cacheState.completed
	}
}

// Complete stops bar's progress tracking, but doesn't remove the bar from rendering queue.
// If you need to remove, invoke Progress.RemoveBar(*Bar) instead.
func (b *Bar) Complete() {
	b.askToComplete(false)
}

func (b *Bar) askToComplete(toRemove bool) bool {
	result := make(chan bool, 1)
	select {
	case b.operateState <- func(s *bState) {
		s.removed = toRemove
		s.completed = true
		result <- true
	}:
		return <-result
	case <-b.done:
		return false
	}
}

func (b *Bar) serve(s *bState, cancel <-chan struct{}) {
	for {
		select {
		case op := <-b.operateState:
			op(s)
		case <-cancel:
			s.completed = true
			cancel = nil
		case <-b.shutdown:
			b.cacheState = s
			close(b.done)
			return
		}
	}
}

func (b *Bar) render(tw int, pSyncer, aSyncer *widthSyncer) <-chan *renderedReader {
	ch := make(chan *renderedReader, 1)

	go func() {
		select {
		case b.operateState <- func(s *bState) {
			var r io.Reader
			defer func() {
				// recovering if external decorators panic
				if p := recover(); p != nil {
					s.panicMsg = fmt.Sprintf("b#%02d panic: %v\n", s.id, p)
					s.pDecorators = nil
					s.aDecorators = nil
					s.completed = true
					r = strings.NewReader(s.panicMsg)
				}
				ch <- &renderedReader{r, s.completed, s.removed}
			}()
			s.draw(tw, pSyncer, aSyncer)
			r = io.MultiReader(s.bufP, s.bufB, s.bufA)
		}:
		case <-b.done:
			s := b.cacheState
			var r io.Reader
			if s.panicMsg != "" {
				r = strings.NewReader(s.panicMsg)
			} else {
				s.draw(tw, pSyncer, aSyncer)
				r = io.MultiReader(s.bufP, s.bufB, s.bufA)
			}
			ch <- &renderedReader{r, s.completed, s.removed}
		}
	}()

	return ch
}

func (s *bState) updateTimePerItemEstimate(amount int, now, next time.Time) {
	lastBlockTime := now.Sub(s.blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(amount)
	s.timePerItem = time.Duration((s.etaAlpha * lastItemEstimate) + (1-s.etaAlpha)*float64(s.timePerItem))
	s.blockStartTime = next
}

func (s *bState) draw(termWidth int, pSyncer, aSyncer *widthSyncer) {
	if termWidth <= 0 {
		termWidth = s.width
	}

	stat := newStatistics(s)

	// render prepend functions to the left of the bar
	s.bufP.Reset()
	for i, f := range s.pDecorators {
		s.bufP.WriteString(f(stat, pSyncer.Accumulator[i], pSyncer.Distributor[i]))
	}

	if !s.trimLeftSpace {
		s.bufP.WriteByte(' ')
	}

	// render append functions to the right of the bar
	s.bufA.Reset()
	if !s.trimRightSpace {
		s.bufA.WriteByte(' ')
	}

	for i, f := range s.aDecorators {
		s.bufA.WriteString(f(stat, aSyncer.Accumulator[i], aSyncer.Distributor[i]))
	}

	prependCount := utf8.RuneCount(s.bufP.Bytes())
	appendCount := utf8.RuneCount(s.bufA.Bytes())

	if termWidth > s.width {
		s.fillBar(s.width)
	} else {
		s.fillBar(termWidth - prependCount - appendCount)
	}
	barCount := utf8.RuneCount(s.bufB.Bytes())
	totalCount := prependCount + barCount + appendCount
	if totalCount > termWidth {
		s.fillBar(termWidth - prependCount - appendCount)
	}
	s.bufA.WriteByte('\n')
}

func (s *bState) fillBar(width int) {
	s.bufB.Reset()
	s.bufB.WriteRune(s.format[rLeft])
	if width <= 2 {
		s.bufB.WriteRune(s.format[rRight])
		return
	}

	// bar s.width without leftEnd and rightEnd runes
	barWidth := width - 2

	completedWidth := decor.CalcPercentage(s.total, s.current, int64(barWidth))

	if s.refill != nil {
		till := decor.CalcPercentage(s.total, s.refill.till, int64(barWidth))
		// append refill rune
		var i int64
		for i = 0; i < till; i++ {
			s.bufB.WriteRune(s.refill.char)
		}
		for i = till; i < completedWidth; i++ {
			s.bufB.WriteRune(s.format[rFill])
		}
	} else {
		var i int64
		for i = 0; i < completedWidth; i++ {
			s.bufB.WriteRune(s.format[rFill])
		}
	}

	if completedWidth < int64(barWidth) && completedWidth > 0 {
		_, size := utf8.DecodeLastRune(s.bufB.Bytes())
		s.bufB.Truncate(s.bufB.Len() - size)
		s.bufB.WriteRune(s.format[rTip])
	}

	for i := completedWidth; i < int64(barWidth); i++ {
		s.bufB.WriteRune(s.format[rEmpty])
	}

	s.bufB.WriteRune(s.format[rRight])
}

func newStatistics(s *bState) *decor.Statistics {
	return &decor.Statistics{
		ID:                  s.id,
		Completed:           s.completed,
		Removed:             s.removed,
		Total:               s.total,
		Current:             s.current,
		StartTime:           s.startTime,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}
}

func (s *bState) updateFormat(format string) {
	for i, n := 0, 0; len(format) > 0; i++ {
		s.format[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
}
