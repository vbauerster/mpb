package mpb

import (
	"io"
	"math"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	rLeft = iota
	rFill
	rTip
	rEmpty
	rRight
)

type barFmtRunes [numFmtRunes]rune
type barFmtBytes [numFmtRunes][]byte

// Bar represents a progress Bar
type Bar struct {
	stateReqCh    chan chan state
	widthCh       chan int
	formatCh      chan string
	etaAlphaCh    chan float64
	incrCh        chan int64
	trimLeftCh    chan bool
	trimRightCh   chan bool
	refillCh      chan *refill
	dCommandCh    chan *dCommandData
	flushedCh     chan struct{}
	removeReqCh   chan struct{}
	completeReqCh chan struct{}
	done          chan struct{}

	// follawing are used after (*Bar.done) is closed
	width int
	state state
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
		id             int
		width          int
		format         barFmtRunes
		etaAlpha       float64
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

func newBar(id int, total int64, width int, format string, wg *sync.WaitGroup, cancel <-chan struct{}) *Bar {
	b := &Bar{
		stateReqCh:    make(chan chan state),
		widthCh:       make(chan int),
		formatCh:      make(chan string),
		etaAlphaCh:    make(chan float64),
		incrCh:        make(chan int64, 1),
		trimLeftCh:    make(chan bool),
		trimRightCh:   make(chan bool),
		refillCh:      make(chan *refill),
		dCommandCh:    make(chan *dCommandData),
		flushedCh:     make(chan struct{}, 1),
		removeReqCh:   make(chan struct{}),
		completeReqCh: make(chan struct{}),
		done:          make(chan struct{}),
	}
	go b.server(id, total, width, format, wg, cancel)
	return b
}

// SetWidth overrides width of individual bar
func (b *Bar) SetWidth(n int) *Bar {
	if n < 2 || isClosed(b.done) {
		return b
	}
	b.widthCh <- n
	return b
}

// TrimLeftSpace removes space befor LeftEnd charater
func (b *Bar) TrimLeftSpace() *Bar {
	if isClosed(b.done) {
		return b
	}
	b.trimLeftCh <- true
	return b
}

// TrimRightSpace removes space after RightEnd charater
func (b *Bar) TrimRightSpace() *Bar {
	if isClosed(b.done) {
		return b
	}
	b.trimRightCh <- true
	return b
}

// Format overrides format of individual bar
func (b *Bar) Format(format string) *Bar {
	if utf8.RuneCountInString(format) != numFmtRunes || isClosed(b.done) {
		return b
	}
	b.formatCh <- format
	return b
}

// SetEtaAlpha sets alfa for exponential-weighted-moving-average ETA estimator
// Defaults to 0.25
// Normally you shouldn't touch this
func (b *Bar) SetEtaAlpha(a float64) *Bar {
	if isClosed(b.done) {
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
	if n < 1 || isClosed(b.done) {
		return
	}
	b.incrCh <- int64(n)
}

// IncrWithReFill increments pb with different fill character
func (b *Bar) IncrWithReFill(n int, r rune) {
	if isClosed(b.done) {
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

func (b *Bar) NumOfAppenders() int {
	return len(b.GetAppenders())
}

// GetPrependers returns slice of prepender DecoratorFunc
func (b *Bar) GetPrependers() []DecoratorFunc {
	s := b.getState()
	return s.prependFuncs
}

func (b *Bar) NumOfPrependers() int {
	return len(b.GetPrependers())
}

// GetStatistics returs *Statistics, which contains information like Tottal,
// Current, TimeElapsed and TimePerItemEstimate
func (b *Bar) GetStatistics() *Statistics {
	s := b.getState()
	return newStatistics(&s)
}

// GetID returs id of the bar
func (b *Bar) GetID() int {
	state := b.getState()
	return state.id
}

// InProgress returns true, while progress is running
// Can be used as condition in for loop
func (b *Bar) InProgress() bool {
	return !isClosed(b.done)
}

// PrependFunc prepends DecoratorFunc
func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
	if isClosed(b.done) {
		return b
	}
	b.dCommandCh <- &dCommandData{dPrepend, f}
	return b
}

// RemoveAllPrependers removes all prepend functions
func (b *Bar) RemoveAllPrependers() {
	if isClosed(b.done) {
		return
	}
	b.dCommandCh <- &dCommandData{dPrependZero, nil}
}

// AppendFunc appends DecoratorFunc
func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
	if isClosed(b.done) {
		return b
	}
	b.dCommandCh <- &dCommandData{dAppend, f}
	return b
}

// RemoveAllAppenders removes all append functions
func (b *Bar) RemoveAllAppenders() {
	if isClosed(b.done) {
		return
	}
	b.dCommandCh <- &dCommandData{dAppendZero, nil}
}

// Completed signals to the bar, that process has been completed.
// You should call this method when total is unknown and you've reached the point
// of process completion.
func (b *Bar) Completed() {
	if isClosed(b.done) {
		return
	}
	b.completeReqCh <- struct{}{}
}

func (b *Bar) getState() state {
	if isClosed(b.done) {
		return b.state
	}
	ch := make(chan state)
	b.stateReqCh <- ch
	return <-ch
}

func (b *Bar) server(id int, total int64, width int, format string, wg *sync.WaitGroup, cancel <-chan struct{}) {
	var completed bool
	timeStarted := time.Now()
	prevStartTime := timeStarted
	var blockStartTime time.Time
	barState := state{
		id:       id,
		width:    width,
		format:   barFmtRunes{'[', '=', '>', '-', ']'},
		etaAlpha: 0.25,
		total:    total,
	}
	if total <= 0 {
		barState.simpleSpinner = getSpinner()
	} else {
		barState.updateFormat(format)
	}
	defer func() {
		b.stop(&barState, width)
		wg.Done()
	}()
	for {
		select {
		case i := <-b.incrCh:
			blockStartTime = time.Now()
			n := barState.current + i
			if total > 0 && n > total {
				barState.current = total
				completed = true
				break // break out of select
			}
			barState.timeElapsed = time.Since(timeStarted)
			barState.timePerItem = calcTimePerItemEstimate(barState.timePerItem, prevStartTime, barState.etaAlpha, i)
			if n == total {
				completed = true
			}
			barState.current = n
			prevStartTime = blockStartTime
		case data := <-b.dCommandCh:
			switch data.action {
			case dAppend:
				barState.appendFuncs = append(barState.appendFuncs, data.f)
			case dAppendZero:
				barState.appendFuncs = nil
			case dPrepend:
				barState.prependFuncs = append(barState.prependFuncs, data.f)
			case dPrependZero:
				barState.prependFuncs = nil
			}
		case ch := <-b.stateReqCh:
			ch <- barState
		case format := <-b.formatCh:
			barState.updateFormat(format)
		case barState.width = <-b.widthCh:
		case barState.refill = <-b.refillCh:
		case barState.trimLeftSpace = <-b.trimLeftCh:
		case barState.trimRightSpace = <-b.trimRightCh:
		case <-b.flushedCh:
			if completed {
				return
			}
		case <-b.completeReqCh:
			return
		case <-b.removeReqCh:
			return
		case <-cancel:
			return
		}
	}
}

func (b *Bar) stop(s *state, width int) {
	b.state = *s
	b.width = width
	close(b.done)
}

func (b *Bar) flushed() {
	if isClosed(b.done) {
		return
	}
	b.flushedCh <- struct{}{}
}

func (b *Bar) remove() {
	if isClosed(b.done) {
		return
	}
	b.removeReqCh <- struct{}{}
}

func (s *state) updateFormat(format string) {
	if format == "" {
		return
	}
	for i, n := 0, 0; len(format) > 0; i++ {
		s.format[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
}

func (b *Bar) render(rFn func(chan []byte), termWidth int, prependWs, appendWs *widthSync) <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer rFn(ch)
		s := b.getState()
		buf := draw(&s, termWidth, prependWs, appendWs)
		buf = append(buf, '\n')
		ch <- buf
	}()

	return ch
}

func draw(s *state, termWidth int, prependWs, appendWs *widthSync) []byte {
	if len(s.prependFuncs) != len(prependWs.listen) || len(s.appendFuncs) != len(appendWs.listen) {
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
			[]byte(f(stat, prependWs.listen[i], prependWs.result[i]))...)
	}

	// render append functions to the right of the bar
	var appendBlock []byte
	for i, f := range s.appendFuncs {
		appendBlock = append(appendBlock,
			[]byte(f(stat, appendWs.listen[i], appendWs.result[i]))...)
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
	fmtBytes := convertFmtRunesToBytes(s.format)

	if s.simpleSpinner != nil {
		for _, block := range [...][]byte{fmtBytes[rLeft], []byte{s.simpleSpinner()}, fmtBytes[rRight]} {
			barBlock = append(barBlock, block...)
		}
		return concatenateBlocks(buf, prependBlock, leftSpace, barBlock, rightSpace, appendBlock)
	}

	barBlock = fillBar(s.total, s.current, s.width, fmtBytes, s.refill)
	barCount := utf8.RuneCount(barBlock)
	totalCount := prependCount + barCount + appendCount
	if totalCount > termWidth {
		newWidth := termWidth - prependCount - appendCount
		barBlock = fillBar(s.total, s.current, newWidth, fmtBytes, s.refill)
	}

	return concatenateBlocks(buf, prependBlock, leftSpace, barBlock, rightSpace, appendBlock)
}

func concatenateBlocks(buf []byte, blocks ...[]byte) []byte {
	for _, block := range blocks {
		buf = append(buf, block...)
	}
	return buf
}

func fillBar(total, current int64, width int, fmtBytes barFmtBytes, rf *refill) []byte {
	if width < 2 || total <= 0 {
		return []byte{}
	}

	// bar width without leftEnd and rightEnd runes
	barWidth := width - 2

	completedWidth := percentage(total, current, barWidth)

	buf := make([]byte, 0, width)
	buf = append(buf, fmtBytes[rLeft]...)

	if rf != nil {
		till := percentage(total, rf.till, barWidth)
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

func newStatistics(s *state) *Statistics {
	return &Statistics{
		Total:               s.total,
		Current:             s.current,
		TimeElapsed:         s.timeElapsed,
		TimePerItemEstimate: s.timePerItem,
	}
}

func convertFmtRunesToBytes(format barFmtRunes) barFmtBytes {
	var fmtBytes barFmtBytes
	for i, r := range format {
		buf := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(buf, r)
		fmtBytes[i] = buf
	}
	return fmtBytes
}

func calcTimePerItemEstimate(tpie time.Duration, blockStartTime time.Time, alpha float64, items int64) time.Duration {
	lastBlockTime := time.Since(blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(items)
	return time.Duration((alpha * lastItemEstimate) + (1-alpha)*float64(tpie))
}

func percentage(total, current int64, ratio int) int {
	if current > total {
		return 0
	}
	num := float64(ratio) * float64(current) / float64(total)
	ceil := math.Ceil(num)
	diff := ceil - num
	// num = 2.34 will return 2
	// num = 2.44 will return 3
	if math.Max(diff, 0.6) == diff {
		return int(num)
	}
	return int(ceil)
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
