package mpb

import (
	"bytes"
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
		id               int
		width            int
		format           fmtRunes
		etaAlpha         float64
		total            int64
		current          int64
		dropRatio        int64
		trimLeftSpace    bool
		trimRightSpace   bool
		completed        bool
		aborted          bool
		dynamic          bool
		startTime        time.Time
		timeElapsed      time.Duration
		blockStartTime   time.Time
		timePerItem      time.Duration
		appendFuncs      []decor.DecoratorFunc
		prependFuncs     []decor.DecoratorFunc
		refill           *refill
		bufP, bufB, bufA *bytes.Buffer
	}
)

func newBar(ID int, total int64, wg *sync.WaitGroup, cancel <-chan struct{}, options ...BarOption) *Bar {
	if total <= 0 {
		total = time.Now().Unix()
	}

	s := state{
		id:        ID,
		total:     total,
		etaAlpha:  etaAlpha,
		dropRatio: 10,
	}

	for _, opt := range options {
		opt(&s)
	}

	s.bufP = bytes.NewBuffer(make([]byte, 0, s.width/2))
	s.bufB = bytes.NewBuffer(make([]byte, 0, s.width))
	s.bufA = bytes.NewBuffer(make([]byte, 0, s.width/2))

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
	}
}

// RemoveAllAppenders removes all append functions
func (b *Bar) RemoveAllAppenders() {
	select {
	case b.ops <- func(s *state) {
		s.appendFuncs = nil
	}:
	case <-b.quit:
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
			if s.dynamic {
				sum -= sum * s.dropRatio / 100
			} else {
				s.current = s.total
				s.completed = true
			}
			return
		}
		s.current = sum
		s.blockStartTime = time.Now()
	}:
	case <-b.quit:
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

// SetTotal sets total dynamically. The final param indicates the very last set,
// in other words you should set it to true when total is determined.
// Also you may consider providing your drop ratio via BarDropRatio BarOption func.
func (b *Bar) SetTotal(total int64, final bool) {
	select {
	case b.ops <- func(s *state) {
		s.total = total
		s.dynamic = !final
	}:
	case <-b.quit:
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
		case <-cancel:
			s.aborted = true
			cancel = nil
			b.Complete()
		case <-b.quit:
			s.completed = true
			return
		}
	}
}

func (b *Bar) render(tw int, prependWs, appendWs *widthSync) <-chan []byte {
	ch := make(chan []byte, 1)

	go func() {
		var st state
		defer func() {
			// recovering if external decorators panic
			if p := recover(); p != nil {
				ch <- []byte(fmt.Sprintf("bar%02d panic: %q\n", st.id, p))
			}
			close(ch)
		}()
		result := make(chan state, 1)
		select {
		case b.ops <- func(s *state) { result <- *s }:
			st = <-result
			if st.completed {
				b.Complete()
			}
		case <-b.done:
			st = b.cacheState
		}
		st.draw(tw, prependWs, appendWs)
		buf := make([]byte, 0, st.bufP.Len()+st.bufB.Len()+st.bufA.Len())
		buf = concatenateBlocks(buf, st.bufP.Bytes(), st.bufB.Bytes(), st.bufA.Bytes())
		buf = append(buf, '\n')
		ch <- buf
	}()

	return ch
}

func (s *state) updateTimePerItemEstimate(amount int) {
	lastBlockTime := time.Since(s.blockStartTime) // shorthand for time.Now().Sub(t)
	lastItemEstimate := float64(lastBlockTime) / float64(amount)
	s.timePerItem = time.Duration((s.etaAlpha * lastItemEstimate) + (1-s.etaAlpha)*float64(s.timePerItem))
}

func (s *state) draw(termWidth int, prependWs, appendWs *widthSync) {
	if termWidth <= 0 {
		termWidth = s.width
	}

	stat := newStatistics(s)

	// render prepend functions to the left of the bar
	s.bufP.Reset()
	for i, f := range s.prependFuncs {
		s.bufP.WriteString(f(stat, prependWs.Listen[i], prependWs.Result[i]))
	}

	if !s.trimLeftSpace {
		s.bufP.WriteByte(' ')
	}

	// render append functions to the right of the bar
	s.bufA.Reset()
	if !s.trimRightSpace {
		s.bufA.WriteByte(' ')
	}

	for i, f := range s.appendFuncs {
		s.bufA.WriteString(f(stat, appendWs.Listen[i], appendWs.Result[i]))
	}

	prependCount := utf8.RuneCount(s.bufP.Bytes())
	appendCount := utf8.RuneCount(s.bufA.Bytes())

	s.fillBar(s.width)
	barCount := utf8.RuneCount(s.bufB.Bytes())
	totalCount := prependCount + barCount + appendCount
	if totalCount > termWidth {
		shrinkWidth := termWidth - prependCount - appendCount
		s.fillBar(shrinkWidth)
	}
}

func (s *state) fillBar(width int) {
	s.bufB.Reset()
	if width <= 2 {
		return
	}

	// bar s.width without leftEnd and rightEnd runes
	barWidth := width - 2

	completedWidth := decor.CalcPercentage(s.total, s.current, barWidth)

	s.bufB.WriteRune(s.format[rLeft])

	if s.refill != nil {
		till := decor.CalcPercentage(s.total, s.refill.till, barWidth)
		// append refill rune
		for i := 0; i < till; i++ {
			s.bufB.WriteRune(s.refill.char)
		}
		for i := till; i < completedWidth; i++ {
			s.bufB.WriteRune(s.format[rFill])
		}
	} else {
		for i := 0; i < completedWidth; i++ {
			s.bufB.WriteRune(s.format[rFill])
		}
	}

	if completedWidth < barWidth && completedWidth > 0 {
		_, size := utf8.DecodeLastRune(s.bufB.Bytes())
		s.bufB.Truncate(s.bufB.Len() - size)
		s.bufB.WriteRune(s.format[rTip])
	}

	for i := completedWidth; i < barWidth; i++ {
		s.bufB.WriteRune(s.format[rEmpty])
	}

	s.bufB.WriteRune(s.format[rRight])
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

func concatenateBlocks(buf []byte, blocks ...[]byte) []byte {
	for _, block := range blocks {
		buf = append(buf, block...)
	}
	return buf
}

func (s *state) updateFormat(format string) {
	for i, n := 0, 0; len(format) > 0; i++ {
		s.format[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
}
