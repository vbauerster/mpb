package mpb

import (
	"io"
	"sync"
	"time"
	"unicode/utf8"
)

// Bar represents a progress Bar
type Bar struct {
	width int
	alpha float64

	fill     byte
	empty    byte
	tip      byte
	leftEnd  byte
	rightEnd byte

	trimLeftSpace, trimRightSpace bool

	incrCh       chan int
	redrawReqCh  chan *redrawRequest
	currentReqCh chan chan int
	statusReqCh  chan chan int
	decoratorCh  chan *decorator
	flushedCh    chan struct{}
	stopCh       chan struct{}
	done         chan struct{}
}

// Statistics represents statistics of the progress bar
// instance of this, sent to DecoratorFunc, as param
type Statistics struct {
	Total, Current, TermWidth        int
	TimeElapsed, TimePerItemEstimate time.Duration
}

func (s *Statistics) Eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

func newBar(total, width int, wg *sync.WaitGroup) *Bar {
	b := &Bar{
		fill:     '=',
		empty:    '-',
		tip:      '>',
		leftEnd:  '[',
		rightEnd: ']',
		alpha:    0.25,
		width:    width,

		incrCh:       make(chan int),
		redrawReqCh:  make(chan *redrawRequest),
		currentReqCh: make(chan chan int),
		statusReqCh:  make(chan chan int),
		decoratorCh:  make(chan *decorator),
		flushedCh:    make(chan struct{}),
		stopCh:       make(chan struct{}),
		done:         make(chan struct{}),
	}
	go b.server(wg, total)
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

func (b *Bar) TrimLeftSpace() *Bar {
	b.trimLeftSpace = true
	return b
}

func (b *Bar) TrimRightSpace() *Bar {
	b.trimRightSpace = true
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
	b.alpha = a
	return b
}

func (b *Bar) ProxyReader(r io.Reader) *Reader {
	return &Reader{r, b}
}

// Incr increments progress bar
func (b *Bar) Incr(n int) {
	if !b.isDone() {
		b.incrCh <- n
	}
}

// Current returns the actual current.
func (b *Bar) Current() int {
	respCh := make(chan int)
	b.currentReqCh <- respCh
	return <-respCh
}

// Stop stops rendering the bar
func (b *Bar) Stop() {
	if !b.isDone() {
		b.stopCh <- struct{}{}
	}
}

// InProgress returns true, while progress is running
// Can be used as condition in for loop
func (b *Bar) InProgress() bool {
	return !b.isDone()
}

// PrependFunc prepends DecoratorFunc
func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorPrepend, f}
	return b
}

// AppendFunc appends DecoratorFunc
func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorAppend, f}
	return b
}

type redrawRequest struct {
	width  int
	respCh chan []byte
}

func (b *Bar) Bytes(width int) []byte {
	if width <= 0 {
		width = b.width
	}
	respCh := make(chan []byte)
	b.redrawReqCh <- &redrawRequest{width, respCh}
	return <-respCh
}

func (b *Bar) server(wg *sync.WaitGroup, total int) {
	timeStarted := time.Now()
	blockStartTime := timeStarted
	var tpie time.Duration
	var timeElapsed time.Duration
	var appendFuncs []DecoratorFunc
	var prependFuncs []DecoratorFunc
	var completed bool
	var current int
	var termWidth int
	for {
		select {
		case i := <-b.incrCh:
			n := current + i
			if n > total {
				current = total
				completed = true
				break
			}
			timeElapsed = time.Since(timeStarted)
			tpie = calcTimePerItemEstimate(tpie, blockStartTime, b.alpha, i)
			blockStartTime = time.Now()
			current = n
			if current == total && !completed {
				completed = true
			}
		case d := <-b.decoratorCh:
			switch d.kind {
			case decoratorAppend:
				appendFuncs = append(appendFuncs, d.f)
			case decoratorPrepend:
				prependFuncs = append(prependFuncs, d.f)
			}
		case respCh := <-b.currentReqCh:
			respCh <- current
		case r := <-b.redrawReqCh:
			termWidth = r.width
			stat := &Statistics{total, current, termWidth, timeElapsed, tpie}
			r.respCh <- b.draw(stat, appendFuncs, prependFuncs)
		case respCh := <-b.statusReqCh:
			respCh <- percentage(total, current, 100)
		case <-b.flushedCh:
			if completed && !b.isDone() {
				close(b.done)
				wg.Done()
			}
		case <-b.stopCh:
			close(b.done)
			if !completed {
				wg.Done()
			}
		}
	}
}

func (b *Bar) draw(stat *Statistics, appendFuncs, prependFuncs []DecoratorFunc) []byte {

	buf := make([]byte, 0, stat.TermWidth)

	barBlock := b.fillBar(stat.Total, stat.Current, b.width)

	// render append functions to the right of the bar
	var appendBlock []byte
	for _, f := range appendFuncs {
		appendBlock = append(appendBlock, []byte(f(stat))...)
	}

	// render prepend functions to the left of the bar
	var prependBlock []byte
	for _, f := range prependFuncs {
		prependBlock = append(prependBlock, []byte(f(stat))...)
	}

	prependCount := utf8.RuneCount(prependBlock)
	barCount := utf8.RuneCount(barBlock)
	appendCount := utf8.RuneCount(appendBlock)

	var leftSpace, rightSpace []byte
	space := []byte{' '}
	if !b.trimLeftSpace {
		prependCount++
		leftSpace = space
	}
	if !b.trimRightSpace {
		appendCount++
		rightSpace = space
	}

	totalCount := prependCount + barCount + appendCount
	if totalCount >= stat.TermWidth {
		newWidth := stat.TermWidth - prependCount - appendCount
		barBlock = b.fillBar(stat.Total, stat.Current, newWidth-1)
	}

	for _, block := range [...][]byte{prependBlock, leftSpace, barBlock, rightSpace, appendBlock} {
		buf = append(buf, block...)
	}

	return buf
}

func (b *Bar) fillBar(total, current, width int) []byte {
	if width < 2 {
		return []byte{b.leftEnd, b.rightEnd}
	}

	buf := make([]byte, width)
	completedWidth := percentage(total, current, width)

	for i := 1; i < completedWidth; i++ {
		buf[i] = b.fill
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

func (b *Bar) isDone() bool {
	select {
	case <-b.done:
		return true
	default:
		return false
	}
}

func (b *Bar) status() int {
	respCh := make(chan int)
	b.statusReqCh <- respCh
	return <-respCh
}

// SortableBarSlice satisfies sort interface
type SortableBarSlice []*Bar

func (p SortableBarSlice) Len() int { return len(p) }

func (p SortableBarSlice) Less(i, j int) bool { return p[i].status() < p[j].status() }

func (p SortableBarSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func calcTimePerItemEstimate(tpie time.Duration, blockStartTime time.Time, alpha float64, items int) time.Duration {
	lastBlockTime := time.Since(blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(items)
	return time.Duration((alpha * lastItemEstimate) + (1-alpha)*float64(tpie))
}

func percentage(total, current, ratio int) int {
	if total == 0 {
		return 0
	}
	return int(float64(ratio) * float64(current) / float64(total))
}
