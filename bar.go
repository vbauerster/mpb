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
	"github.com/vbauerster/mpb/internal"
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
)

type barRunes [formatLen]rune

// Bar represents a progress Bar
type Bar struct {
	priority int
	index    int

	runningBar    *Bar
	cacheState    *bState
	operateState  chan func(*bState)
	frameReaderCh chan io.Reader

	// done is closed by Bar's goroutine, after cacheState is written
	done chan struct{}
	// shutdown is closed from master Progress goroutine only
	shutdown chan struct{}
}

type (
	bState struct {
		id                   int
		width                int
		runes                barRunes
		total                int64
		current              int64
		totalAutoIncrTrigger int64
		totalAutoIncrBy      int64
		trimLeftSpace        bool
		trimRightSpace       bool
		toComplete           bool
		dynamic              bool
		removeOnComplete     bool
		barClearOnComplete   bool
		completeFlushed      bool
		aDecorators          []decor.Decorator
		pDecorators          []decor.Decorator
		amountReceivers      []decor.AmountReceiver
		shutdownListeners    []decor.ShutdownListener
		refill               *refill
		bufP, bufB, bufA     *bytes.Buffer
		panicMsg             string

		// following options are assigned to the *Bar
		priority   int
		runningBar *Bar
	}
	refill struct {
		char rune
		till int64
	}
	frameReader struct {
		io.Reader
		toShutdown       bool
		removeOnComplete bool
	}
)

func newBar(wg *sync.WaitGroup, id int, total int64, cancel <-chan struct{}, options ...BarOption) *Bar {
	dynamic := total <= 0
	if dynamic {
		total = time.Now().Unix()
	}

	s := &bState{
		id:       id,
		priority: id,
		total:    total,
		dynamic:  dynamic,
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
		priority:      s.priority,
		runningBar:    s.runningBar,
		operateState:  make(chan func(*bState)),
		frameReaderCh: make(chan io.Reader, 1),
		done:          make(chan struct{}),
		shutdown:      make(chan struct{}),
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
//
//	`r` io.Reader to be wrapped
//
//	`sbChannels` optional start block channels
func (b *Bar) ProxyReader(r io.Reader, sbChannels ...chan<- time.Time) *Reader {
	proxyReader := &Reader{
		Reader:     r,
		bar:        b,
		sbChannels: sbChannels,
	}
	return proxyReader
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
	result := make(chan int)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.aDecorators) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.aDecorators)
	}
}

// NumOfPrependers returns current number of prepend decorators
func (b *Bar) NumOfPrependers() int {
	result := make(chan int)
	select {
	case b.operateState <- func(s *bState) { result <- len(s.pDecorators) }:
		return <-result
	case <-b.done:
		return len(b.cacheState.pDecorators)
	}
}

// ID returs id of the bar
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

// Total returns bar's total number.
func (b *Bar) Total() int64 {
	result := make(chan int64)
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
	b.operateState <- func(s *bState) {
		if total != 0 {
			s.total = total
		}
		s.dynamic = !final
	}
}

// Increment is a shorthand for b.IncrBy(1)
func (b *Bar) Increment() {
	b.IncrBy(1)
}

// IncrBy increments progress bar by amount of n
func (b *Bar) IncrBy(n int) {
	select {
	case b.operateState <- func(s *bState) {
		s.current += int64(n)
		if s.dynamic {
			curp := internal.Percentage(s.total, s.current, 100)
			if 100-curp <= s.totalAutoIncrTrigger {
				s.total += s.totalAutoIncrBy
			}
		} else if s.current >= s.total {
			s.current = s.total
			s.toComplete = true
		}
		for _, ar := range s.amountReceivers {
			ar.NextAmount(n)
		}
	}:
	case <-b.done:
	}
}

// Completed reports whether the bar is in completed state
func (b *Bar) Completed() bool {
	result := make(chan bool)
	b.operateState <- func(s *bState) { result <- s.toComplete }
	return <-result
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
			for _, sl := range s.shutdownListeners {
				sl.Shutdown()
			}
			return
		}
	}
}

func (b *Bar) render(debugOut io.Writer, tw int, pSyncer, aSyncer *widthSyncer) {
	var r io.Reader
	select {
	case b.operateState <- func(s *bState) {
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
			b.frameReaderCh <- &frameReader{
				Reader:           r,
				toShutdown:       s.toComplete && !s.completeFlushed,
				removeOnComplete: s.removeOnComplete,
			}
			s.completeFlushed = s.toComplete
		}()
		r = s.draw(tw, pSyncer, aSyncer)
	}:
	case <-b.done:
		s := b.cacheState
		if s.panicMsg != "" {
			r = strings.NewReader(fmt.Sprintf(fmt.Sprintf("%%.%ds\n", tw), s.panicMsg))
		} else {
			r = s.draw(tw, pSyncer, aSyncer)
		}
		b.frameReaderCh <- &frameReader{
			Reader: r,
		}
	}
}

func (s *bState) draw(termWidth int, pSyncer, aSyncer *widthSyncer) io.Reader {
	defer s.bufA.WriteByte('\n')

	if termWidth <= 0 {
		termWidth = s.width
	}

	stat := newStatistics(s)

	// render prepend functions to the left of the bar
	for i, d := range s.pDecorators {
		s.bufP.WriteString(d.Decor(stat, pSyncer.Accumulator[i], pSyncer.Distributor[i]))
	}

	for i, d := range s.aDecorators {
		s.bufA.WriteString(d.Decor(stat, aSyncer.Accumulator[i], aSyncer.Distributor[i]))
	}

	prependCount := utf8.RuneCount(s.bufP.Bytes())
	appendCount := utf8.RuneCount(s.bufA.Bytes())

	if s.barClearOnComplete && s.completeFlushed {
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

	completedWidth := internal.Percentage(s.total, s.current, int64(barWidth))

	if s.refill != nil {
		till := internal.Percentage(s.total, s.refill.till, int64(barWidth))
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

func newStatistics(s *bState) *decor.Statistics {
	return &decor.Statistics{
		ID:        s.id,
		Completed: s.completeFlushed,
		Total:     s.total,
		Current:   s.current,
	}
}

func strToBarRunes(format string) (array barRunes) {
	for i, n := 0, 0; len(format) > 0; i++ {
		array[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
	return
}
