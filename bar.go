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
	stateReqCh  chan state
	incrCh      chan incrReq
	dCommandCh  chan *dCommandData
	flushedCh   chan struct{}
	removeReqCh chan struct{}
	completedCh chan struct{}
	done        chan struct{}
	cancel      <-chan struct{}

	// follawing are used after (*Bar.done) is closed
	width int
	state state
}

// Statistics represents statistics of the progress bar.
// Cantains: Total, Current, TimeElapsed, TimePerItemEstimate
type Statistics struct {
	Completed           bool
	Total               int64
	Current             int64
	StartTime           time.Time
	TimeElapsed         time.Duration
	TimePerItemEstimate time.Duration
}

// Refil is a struct for b.IncrWithReFill
type Refill struct {
	Char rune
	Till int64
}

// Eta returns exponential-weighted-moving-average ETA estimator
func (s *Statistics) Eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

type (
	incrReq struct {
		amount int64
		refill *Refill
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
		completed      bool
		startTime      time.Time
		timeElapsed    time.Duration
		timePerItem    time.Duration
		appendFuncs    []DecoratorFunc
		prependFuncs   []DecoratorFunc
		simpleSpinner  func() byte
		refill         *Refill
	}
)

func newBar(id int, total int64, wg *sync.WaitGroup, conf *userConf) *Bar {
	b := &Bar{
		width:       conf.width,
		stateReqCh:  make(chan state),
		incrCh:      make(chan incrReq),
		dCommandCh:  make(chan *dCommandData),
		flushedCh:   make(chan struct{}),
		removeReqCh: make(chan struct{}),
		completedCh: make(chan struct{}),
		done:        make(chan struct{}),
		cancel:      conf.cancel,
	}

	s := state{
		id:       id,
		total:    total,
		width:    conf.width,
		etaAlpha: 0.25,
	}

	if total <= 0 {
		s.simpleSpinner = getSpinner()
	} else {
		s.updateFormat(conf.format)
	}

	go b.server(wg, s)
	return b
}

// SetWidth overrides width of individual bar
func (b *Bar) SetWidth(n int) *Bar {
	if n < 2 || isClosed(b.done) {
		return b
	}
	s := <-b.stateReqCh
	s.width = n
	b.stateReqCh <- s
	return b
}

// TrimLeftSpace removes space befor LeftEnd charater
func (b *Bar) TrimLeftSpace() *Bar {
	if isClosed(b.done) {
		return b
	}
	// b.trimLeftCh <- true
	s := <-b.stateReqCh
	s.trimLeftSpace = true
	b.stateReqCh <- s
	return b
}

// TrimRightSpace removes space after RightEnd charater
func (b *Bar) TrimRightSpace() *Bar {
	if isClosed(b.done) {
		return b
	}
	// b.trimRightCh <- true
	s := <-b.stateReqCh
	s.trimRightSpace = true
	b.stateReqCh <- s
	return b
}

// Format overrides format of individual bar
func (b *Bar) Format(format string) *Bar {
	if utf8.RuneCountInString(format) != numFmtRunes || isClosed(b.done) {
		return b
	}
	s := <-b.stateReqCh
	s.updateFormat(format)
	b.stateReqCh <- s
	return b
}

// SetEtaAlpha sets alfa for exponential-weighted-moving-average ETA estimator
// Defaults to 0.25
// Normally you shouldn't touch this
func (b *Bar) SetEtaAlpha(a float64) *Bar {
	if isClosed(b.done) {
		return b
	}
	s := <-b.stateReqCh
	s.etaAlpha = a
	b.stateReqCh <- s
	return b
}

// ProxyReader wrapper for io operations, like io.Copy
func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Incr increments progress bar
func (b *Bar) Incr(n int) {
	b.IncrWithReFill(n, nil)
}

// IncrWithReFill increments pb with different fill character
func (b *Bar) IncrWithReFill(n int, refill *Refill) {
	if n < 1 {
		return
	}
	select {
	case b.incrCh <- incrReq{int64(n), refill}:
	case <-b.done:
		return
	}
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
	return b.getState().id
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
	select {
	case b.completedCh <- struct{}{}:
	case <-b.done:
		return
	}
}

func (b *Bar) getState() state {
	select {
	case s := <-b.stateReqCh:
		return s
	case <-b.done:
		return b.state
	}
}

func (b *Bar) server(wg *sync.WaitGroup, s state) {
	var incrStartTime time.Time

	defer func() {
		b.state = s
		wg.Done()
		// fmt.Printf("Exited bar %d\n", s.id)
		close(b.done)
	}()

	for {
		select {
		case b.stateReqCh <- s:
		case s = <-b.stateReqCh:
		case r := <-b.incrCh:
			if s.current == 0 {
				incrStartTime = time.Now()
				s.startTime = incrStartTime
			}
			n := s.current + r.amount
			if s.total > 0 && n > s.total {
				s.current = s.total
				s.completed = true
				break // break out of select
			}
			s.timeElapsed = time.Since(s.startTime)
			s.updateTimePerItemEstimate(incrStartTime, r.amount)
			if n == s.total {
				s.completed = true
			}
			s.current = n
			s.refill = r.refill
			incrStartTime = time.Now()
		case data := <-b.dCommandCh:
			switch data.action {
			case dAppend:
				s.appendFuncs = append(s.appendFuncs, data.f)
			case dAppendZero:
				s.appendFuncs = nil
			case dPrepend:
				s.prependFuncs = append(s.prependFuncs, data.f)
			case dPrependZero:
				s.prependFuncs = nil
			}
		case <-b.flushedCh:
			if s.completed {
				return
			}
		case <-b.completedCh:
			s.completed = true
			return
		case <-b.removeReqCh:
			return
		case <-b.cancel:
			return
		}
	}
}

func (b *Bar) flushed() {
	select {
	case b.flushedCh <- struct{}{}:
	case <-b.done:
		return
	}
}

func (b *Bar) remove() {
	select {
	case b.removeReqCh <- struct{}{}:
	case <-b.done:
		return
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

func (s *state) updateFormat(format string) {
	for i, n := 0, 0; len(format) > 0; i++ {
		s.format[i], n = utf8.DecodeRuneInString(format)
		format = format[n:]
	}
}

func (s *state) updateTimePerItemEstimate(incrStartTime time.Time, amount int64) {
	lastBlockTime := time.Since(incrStartTime) // shorthand for time.Now().Sub(t)
	lastItemEstimate := float64(lastBlockTime) / float64(amount)
	s.timePerItem = time.Duration((s.etaAlpha * lastItemEstimate) + (1-s.etaAlpha)*float64(s.timePerItem))
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
		for _, block := range [...][]byte{fmtBytes[rLeft], {s.simpleSpinner()}, fmtBytes[rRight]} {
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

func fillBar(total, current int64, width int, fmtBytes barFmtBytes, rf *Refill) []byte {
	if width < 2 || total <= 0 {
		return []byte{}
	}

	// bar width without leftEnd and rightEnd runes
	barWidth := width - 2

	completedWidth := percentage(total, current, barWidth)

	buf := make([]byte, 0, width)
	buf = append(buf, fmtBytes[rLeft]...)

	if rf != nil {
		till := percentage(total, rf.Till, barWidth)
		rbytes := make([]byte, utf8.RuneLen(rf.Char))
		utf8.EncodeRune(rbytes, rf.Char)
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
		Completed:           s.completed,
		Total:               s.total,
		Current:             s.current,
		StartTime:           s.startTime,
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

func percentage(total, current int64, ratio int) int {
	if total == 0 || current > total {
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
