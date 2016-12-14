package mpb

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"
)

type decoratorFuncType uint

const (
	decoratorAppend decoratorFuncType = iota
	decoratorPrepend
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(s *Statistics) string

type decorator struct {
	kind decoratorFuncType
	f    DecoratorFunc
}

// Bar represents a progress Bar
type Bar struct {
	total   int
	width   int
	alpha   float64
	stopped bool

	fill     byte
	empty    byte
	tip      byte
	leftEnd  byte
	rightEnd byte

	incrCh        chan int
	redrawReqCh   chan chan []byte
	progressReqCh chan chan int
	decoratorCh   chan *decorator
	flushedCh     chan struct{}
	stopCh        chan struct{}
	done          chan struct{}
}

type Statistics struct {
	Total, Current                   int
	TimeElapsed, TimePerItemEstimate time.Duration
}

func (s *Statistics) eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

func newBar(total, width int, wg *sync.WaitGroup) *Bar {
	b := &Bar{
		fill:          '=',
		empty:         '-',
		tip:           '>',
		leftEnd:       '[',
		rightEnd:      ']',
		alpha:         0.25,
		total:         total,
		width:         width,
		incrCh:        make(chan int),
		redrawReqCh:   make(chan chan []byte),
		progressReqCh: make(chan chan int),
		decoratorCh:   make(chan *decorator),
		flushedCh:     make(chan struct{}),
		stopCh:        make(chan struct{}),
		done:          make(chan struct{}),
	}
	go b.server(wg)
	return b
}

// SetWidth sets width of the bar
func (b *Bar) SetWidth(n int) *Bar {
	if n <= 0 {
		return b
	}
	b.width = n
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

// String returns the string representation of the bar
func (b *Bar) String() string {
	respCh := make(chan []byte)
	b.redrawReqCh <- respCh
	return string(<-respCh)
}

func (b *Bar) Incr(n int) {
	if !b.IsCompleted() {
		b.incrCh <- n
	}
}

func (b *Bar) Stop() {
	if !b.stopped {
		b.stopCh <- struct{}{}
		b.stopped = true
	}
}

func (b *Bar) IsCompleted() bool {
	select {
	case <-b.done:
		return true
	default:
		return false
	}
}

func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorPrepend, f}
	return b
}

func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
	b.decoratorCh <- &decorator{decoratorAppend, f}
	return b
}

func (b *Bar) PrependName(name string, padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, name)
	})
	return b
}

func (b *Bar) PrependETA(padding int) *Bar {
	layout := "ETA%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, time.Duration(s.eta().Seconds())*time.Second)
	})
	return b
}

func (b *Bar) AppendETA() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		return fmt.Sprintf("ETA %s", time.Duration(s.eta().Seconds())*time.Second)
	})
	return b
}

func (b *Bar) PrependElapsed(padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, time.Duration(s.TimeElapsed.Seconds())*time.Second)
	})
	return b
}

func (b *Bar) AppendElapsed() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		return fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
	})
	return b
}

func (b *Bar) AppendPercentage() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		completed := int(100 * float64(s.Current) / float64(s.Total))
		return fmt.Sprintf("%3d %%", completed)
	})
	return b
}

func (b *Bar) PrependPercentage(padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "d %%"
	b.PrependFunc(func(s *Statistics) string {
		completed := int(100 * float64(s.Current) / float64(s.Total))
		return fmt.Sprintf(layout, completed)
	})
	return b
}

func (b *Bar) server(wg *sync.WaitGroup) {
	timeStarted := time.Now()
	blockStartTime := timeStarted
	buf := make([]byte, b.width, b.width+24)
	var tpie time.Duration
	var timeElapsed time.Duration
	var appendFuncs []DecoratorFunc
	var prependFuncs []DecoratorFunc
	var done bool
	var current int
	for {
		select {
		case i := <-b.incrCh:
			n := current + i
			// fmt.Fprintf(os.Stderr, "n = %+v\n", n)
			if n > b.total {
				current = b.total
				done = true
				break
			}
			timeElapsed = time.Since(timeStarted)
			tpie = calcTimePerItemEstimate(tpie, blockStartTime, b.alpha, i)
			blockStartTime = time.Now()
			current = n
			if current == b.total && !done {
				done = true
			}
		case d := <-b.decoratorCh:
			switch d.kind {
			case decoratorAppend:
				appendFuncs = append(appendFuncs, d.f)
			case decoratorPrepend:
				prependFuncs = append(prependFuncs, d.f)
			}
		case respCh := <-b.redrawReqCh:
			stat := &Statistics{b.total, current, timeElapsed, tpie}
			respCh <- b.draw(stat, buf, appendFuncs, prependFuncs)
		case respCh := <-b.progressReqCh:
			respCh <- int(100 * float64(current) / float64(b.total))
		case <-b.flushedCh:
			if done && !b.IsCompleted() {
				// fmt.Fprintln(os.Stderr, "flushedCh: wg.Done")
				close(b.done)
				wg.Done()
			}
		case <-b.stopCh:
			// fmt.Fprintln(os.Stderr, "received stop signal")
			if !done {
				// fmt.Fprintln(os.Stderr, "closing done chan: done = false")
				close(b.done)
				wg.Done()
			}
			return
		}
	}
}

func (b *Bar) draw(stat *Statistics, buf []byte, appendFuncs, prependFuncs []DecoratorFunc) []byte {
	completedWidth := stat.Current * b.width / b.total

	for i := 0; i < completedWidth; i++ {
		buf[i] = b.fill
	}
	for i := completedWidth; i < b.width; i++ {
		buf[i] = b.empty
	}
	// set tip bit
	if completedWidth > 0 && completedWidth < b.width {
		buf[completedWidth-1] = b.tip
	}

	// set left and right ends bits
	buf[0], buf[len(buf)-1] = b.leftEnd, b.rightEnd

	// render append functions to the right of the bar
	for _, f := range appendFuncs {
		buf = append(buf, ' ')
		buf = append(buf, []byte(f(stat))...)
	}

	// render prepend functions to the left of the bar
	for _, f := range prependFuncs {
		args := []byte(f(stat))
		args = append(args, ' ')
		buf = append(args, buf...)
	}
	return buf
}

func calcTimePerItemEstimate(tpie time.Duration, blockStartTime time.Time, alpha float64, items int) time.Duration {
	lastBlockTime := time.Since(blockStartTime)
	lastItemEstimate := float64(lastBlockTime) / float64(items)
	return time.Duration((alpha * lastItemEstimate) + (1-alpha)*float64(tpie))
}

func (b *Bar) progress() int {
	respCh := make(chan int)
	b.progressReqCh <- respCh
	return <-respCh
}

type SortableBarSlice []*Bar

func (p SortableBarSlice) Len() int { return len(p) }

func (p SortableBarSlice) Less(i, j int) bool { return p[i].progress() < p[j].progress() }

func (p SortableBarSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
