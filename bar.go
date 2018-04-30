package mpb

import (
	"bytes"
	"fmt"
	"io"
	"strings"
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

type barRunes [formatLen]rune

// Bar represents a progress Bar
type Bar struct {
	priority int
	index    int

	// pointer to running bar, which this bar should replace
	runningBar *Bar

	// completed is set from master Progress goroutine only
	completed bool

	removeOnComplete bool

	operateState chan func(*bState)
	// done is closed by Bar's goroutine, after cacheState is written
	done chan struct{}
	// shutdown is closed from master Progress goroutine only
	shutdown chan struct{}

	cacheState *bState
}

type (
	bState struct {
		id                   int
		width                int
		runes                barRunes
		etaAlpha             float64
		total                int64
		current              int64
		totalAutoIncrTrigger int64
		totalAutoIncrBy      int64
		trimLeftSpace        bool
		trimRightSpace       bool
		toComplete           bool
		dynamic              bool
		noBarOnComplete      bool
		startTime            time.Time
		timeElapsed          time.Duration
		blockStartTime       time.Time
		timePerItem          time.Duration
		aDecorators          []decor.DecoratorFunc
		pDecorators          []decor.DecoratorFunc
		refill               *refill
		bufP, bufB, bufA     *bytes.Buffer
		panicMsg             string

		// following options are assigned to the *Bar
		priority         int
		removeOnComplete bool
		runningBar       *Bar
	}
	refill struct {
		char rune
		till int64
	}
	bFrame struct {
		bar        *Bar
		reader     io.Reader
		toComplete bool
	}
)

func newBar(wg *sync.WaitGroup, id int, total int64, cancel <-chan struct{}, options ...BarOption) *Bar {
	if total <= 0 {
		total = time.Now().Unix()
	}

	s := &bState{
		id:       id,
		priority: id,
		total:    total,
		etaAlpha: etaAlpha,
	}

	for _, opt := range options {
		if opt != nil {
			opt(s)
		}
	}

	s.bufP = bytes.NewBuffer(make([]byte, 0, s.width))
	s.bufB = bytes.NewBuffer(make([]byte, 0, s.width))
	s.bufA = bytes.NewBuffer(make([]byte, 0, s.width))

	b := &Bar{
		priority:         s.priority,
		removeOnComplete: s.removeOnComplete,
		runningBar:       s.runningBar,
		operateState:     make(chan func(*bState)),
		done:             make(chan struct{}),
		shutdown:         make(chan struct{}),
	}

	if b.runningBar != nil {
		b.priority = b.runningBar.priority
	}

	go b.serve(wg, s, cancel)
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
		if total != 0 {
			s.total = total
		}
		s.dynamic = !final
	}:
	case <-b.done:
	}
}

// IncrBy increments progress bar by amount of n
func (b *Bar) IncrBy(n int) {
	if n < 1 {
		return
	}
	now := time.Now()
	select {
	case b.operateState <- func(s *bState) {
		if s.toComplete {
			return
		}
		if s.current == 0 {
			s.startTime = now
			s.blockStartTime = now
		} else {
			s.updateTimePerItemEstimate(n, now)
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
			s.toComplete = true
		}
	}:
	case <-b.done:
	}
}

// Completed reports whether the bar is in completed state
func (b *Bar) Completed() bool {
	result := make(chan bool, 1)
	select {
	case b.operateState <- func(s *bState) { result <- s.toComplete }:
		return <-result
	case <-b.done:
		return b.cacheState.toComplete
	}
}

func (b *Bar) serve(wg *sync.WaitGroup, s *bState, cancel <-chan struct{}) {
	defer wg.Done()
	for {
		select {
		case op := <-b.operateState:
			op(s)
		case <-cancel:
			s.toComplete = true
			cancel = nil
		case <-b.shutdown:
			b.cacheState = s
			close(b.done)
			return
		}
	}
}

func (b *Bar) render(debugOut io.Writer, tw int, pSyncer, aSyncer *widthSyncer) <-chan *bFrame {
	ch := make(chan *bFrame, 1)

	go func() {
		select {
		case b.operateState <- func(s *bState) {
			var r io.Reader
			defer func() {
				// recovering if external decorators panic
				if p := recover(); p != nil {
					s.panicMsg = fmt.Sprintf("panic: %v", p)
					s.pDecorators = nil
					s.aDecorators = nil
					s.toComplete = true
					// truncate panic msg to one tw line, if necessary
					r = strings.NewReader(fmt.Sprintf(fmt.Sprintf("%%.%ds\n", tw), s.panicMsg))
					fmt.Fprintf(debugOut, "%s %s bar id %02d %v\n", "[mpb]", time.Now(), s.id, s.panicMsg)
				}
				ch <- &bFrame{b, r, s.toComplete}
			}()
			r = s.draw(tw, pSyncer, aSyncer)
		}:
		case <-b.done:
			s := b.cacheState
			var r io.Reader
			if s.panicMsg != "" {
				r = strings.NewReader(fmt.Sprintf(fmt.Sprintf("%%.%ds\n", tw), s.panicMsg))
			} else {
				r = s.draw(tw, pSyncer, aSyncer)
			}
			ch <- &bFrame{b, r, s.toComplete}
		}
	}()

	return ch
}

func (s *bState) draw(termWidth int, pSyncer, aSyncer *widthSyncer) io.Reader {
	defer s.bufA.WriteByte('\n')

	if termWidth <= 0 {
		termWidth = s.width
	}

	stat := newStatistics(s)

	// render prepend functions to the left of the bar
	for i, f := range s.pDecorators {
		s.bufP.WriteString(f(stat, pSyncer.Accumulator[i], pSyncer.Distributor[i]))
	}

	for i, f := range s.aDecorators {
		s.bufA.WriteString(f(stat, aSyncer.Accumulator[i], aSyncer.Distributor[i]))
	}

	prependCount := utf8.RuneCount(s.bufP.Bytes())
	appendCount := utf8.RuneCount(s.bufA.Bytes())

	if s.toComplete && s.noBarOnComplete {
		return io.MultiReader(s.bufP, s.bufA)
	}

	s.fillBar(s.width)
	barCount := utf8.RuneCount(s.bufB.Bytes())
	totalCount := prependCount + barCount + appendCount
	if spaceCount := 0; totalCount > termWidth {
		if !s.trimLeftSpace {
			spaceCount++
		}
		if !s.trimRightSpace {
			spaceCount++
		}
		s.fillBar(termWidth - prependCount - appendCount - spaceCount)
	}

	return io.MultiReader(s.bufP, s.bufB, s.bufA)
}

func (s *bState) fillBar(width int) {
	defer func() {
		s.bufB.WriteRune(s.runes[rRight])
		if !s.trimRightSpace {
			s.bufB.WriteByte(' ')
		}
	}()

	s.bufB.Reset()
	if !s.trimLeftSpace {
		s.bufB.WriteByte(' ')
	}
	s.bufB.WriteRune(s.runes[rLeft])
	if width <= 2 {
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
			s.bufB.WriteRune(s.runes[rFill])
		}
	} else {
		var i int64
		for i = 0; i < completedWidth; i++ {
			s.bufB.WriteRune(s.runes[rFill])
		}
	}

	if completedWidth < int64(barWidth) && completedWidth > 0 {
		_, size := utf8.DecodeLastRune(s.bufB.Bytes())
		s.bufB.Truncate(s.bufB.Len() - size)
		s.bufB.WriteRune(s.runes[rTip])
	}

	for i := completedWidth; i < int64(barWidth); i++ {
		s.bufB.WriteRune(s.runes[rEmpty])
	}
}

func (s *bState) updateTimePerItemEstimate(amount int, now time.Time) {
	lastBlockTime := now.Sub(s.blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(amount)
	s.timePerItem = time.Duration((s.etaAlpha * lastItemEstimate) + (1-s.etaAlpha)*float64(s.timePerItem))
	s.blockStartTime = now
}

func newStatistics(s *bState) *decor.Statistics {
	return &decor.Statistics{
		ID:                  s.id,
		Completed:           s.toComplete,
		Total:               s.total,
		Current:             s.current,
		StartTime:           s.startTime,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}
}

func strToBarRunes(format string) (array barRunes) {
	for i, n := 0, 0; len(format) > 0; i++ {
		array[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
	return
}
