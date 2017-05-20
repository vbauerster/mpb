// Copyright (C) 2016-2017 Vladimir Bauer
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mpb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/cwriter"
)

// ErrCallAfterStop thrown by panic, if Progress methods like (*Progress).AddBar()
// are called after (*Progress).Stop() has been called
var ErrCallAfterStop = errors.New("method call on stopped Progress instance")

// default RefreshRate
var rr = 100 * time.Millisecond

type (
	// BeforeRender is a func, which gets called before render process
	BeforeRender func([]*Bar)
	barAction    uint

	bCommandData struct {
		action barAction
		bar    *Bar
		result chan bool
	}

	widthSync struct {
		listen []chan int
		result []chan int
	}

	// config changeable by user
	userConf struct {
		width        int
		format       string
		beforeRender BeforeRender
		cw           *cwriter.Writer
		ticker       *time.Ticker

		shutdownNotifier chan struct{}
		cancel           <-chan struct{}
	}
)

const (
	bAdd barAction = iota
	bRemove
)

const (
	// default width
	pwidth = 80
	// default format
	pformat = "[=>-]"
	// number of format runes for bar
	numFmtRunes = 5
)

// Progress represents the container that renders Progress bars
type Progress struct {
	// WaitGroup for internal rendering sync
	wg *sync.WaitGroup

	done          chan struct{}
	userConf      chan userConf
	bCommandCh    chan *bCommandData
	barCountReqCh chan chan int
	beforeStop    chan struct{}
}

// New creates new Progress instance, which will orchestrate bars rendering
// process. It acceepts context.Context, for cancellation.
// If you don't plan to cancel, it is safe to feed with nil
func New() *Progress {
	p := &Progress{
		wg:            new(sync.WaitGroup),
		done:          make(chan struct{}),
		userConf:      make(chan userConf),
		bCommandCh:    make(chan *bCommandData),
		barCountReqCh: make(chan chan int),
		beforeStop:    make(chan struct{}),
	}
	go p.server(userConf{
		width:  pwidth,
		format: pformat,
		cw:     cwriter.New(os.Stdout),
		ticker: time.NewTicker(rr),
	})
	return p
}

// WithCancel cancellation via channel.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
// or nil channel passed
func (p *Progress) WithCancel(ch <-chan struct{}) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if ch == nil {
		panic("nil cancel channel")
	}
	conf := <-p.userConf
	conf.cancel = ch
	p.userConf <- conf
	return p
}

// SetWidth overrides default (70) width of bar(s).
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) SetWidth(width int) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if width < 0 {
		return p
	}
	conf := <-p.userConf
	conf.width = width
	p.userConf <- conf
	return p
}

// SetOut sets underlying writer of progress. Default one is os.Stdout.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) SetOut(w io.Writer) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if w == nil {
		return p
	}
	conf := <-p.userConf
	conf.cw.Flush()
	conf.cw = cwriter.New(w)
	p.userConf <- conf
	return p
}

// RefreshRate overrides default (100ms) refresh rate value
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	conf := <-p.userConf
	conf.ticker.Stop()
	rr = d
	conf.ticker = time.NewTicker(rr)
	p.userConf <- conf
	return p
}

// BeforeRenderFunc accepts a func, which gets called before render process.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) BeforeRenderFunc(f BeforeRender) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	conf := <-p.userConf
	conf.beforeRender = f
	p.userConf <- conf
	return p
}

// AddBar creates a new progress bar and adds to the container.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) AddBar(total int64) *Bar {
	return p.AddBarWithID(0, total)
}

// AddBarWithID creates a new progress bar and adds to the container.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) AddBarWithID(id int, total int64) *Bar {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	conf := <-p.userConf
	result := make(chan bool)
	bar := newBar(id, total, p.wg, &conf)
	p.bCommandCh <- &bCommandData{bAdd, bar, result}
	if <-result {
		p.wg.Add(1)
	}
	return bar
}

// RemoveBar removes bar at any time.
// Pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) RemoveBar(b *Bar) bool {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	result := make(chan bool)
	p.bCommandCh <- &bCommandData{bRemove, b, result}
	return <-result
}

// BarCount returns bars count in the container.
// Pancis if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) BarCount() int {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	respCh := make(chan int, 1)
	p.barCountReqCh <- respCh
	return <-respCh
}

// ShutdownNotify means to be notified when main rendering goroutine quits, usualy after p.Stop() call.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) ShutdownNotify(ch chan struct{}) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	conf := <-p.userConf
	conf.shutdownNotifier = ch
	p.userConf <- conf
	return p
}

// Format sets custom format for underlying bar(s), default one is "[=>-]".
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
func (p *Progress) Format(format string) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if utf8.RuneCountInString(format) != numFmtRunes {
		return p
	}
	conf := <-p.userConf
	conf.format = format
	p.userConf <- conf
	return p
}

// Stop shutdowns Progress' goroutine.
// Should be called only after each bar's work done, i.e. bar has reached its
// 100 %. It is NOT for cancelation. Use WithContext or WithCancel for
// cancelation purposes.
func (p *Progress) Stop() {
	if isClosed(p.done) {
		return
	}
	p.beforeStop <- struct{}{}
	p.wg.Wait()
	close(p.bCommandCh)
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(conf userConf) {

	defer func() {
		conf.ticker.Stop()
		close(p.done)
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
		}
	}()

	recoverFn := func(ch chan []byte) {
		if p := recover(); p != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			os.Stderr.Write(buf[:n])
			ch <- []byte(fmt.Sprintln(p))
		}
		close(ch)
	}

	bars := make([]*Bar, 0, 3)

	for {
		select {
		case p.userConf <- conf:
		case conf = <-p.userConf:
		case data, ok := <-p.bCommandCh:
			if !ok {
				return
			}
			switch data.action {
			case bAdd:
				bars = append(bars, data.bar)
				data.result <- true
			case bRemove:
				var ok bool
				for i, b := range bars {
					if b == data.bar {
						bars = append(bars[:i], bars[i+1:]...)
						ok = true
						b.remove()
						break
					}
				}
				data.result <- ok
			}
		case respCh := <-p.barCountReqCh:
			respCh <- len(bars)
		case <-conf.ticker.C:
			numBars := len(bars)

			if numBars == 0 {
				break
			}

			if conf.beforeRender != nil {
				conf.beforeRender(bars)
			}

			quitWidthSyncCh := make(chan struct{})
			time.AfterFunc(rr, func() {
				close(quitWidthSyncCh)
			})

			b0 := bars[0]
			prependWs := newWidthSync(quitWidthSyncCh, numBars, b0.NumOfPrependers())
			appendWs := newWidthSync(quitWidthSyncCh, numBars, b0.NumOfAppenders())

			width, _, _ := cwriter.GetTermSize()

			sequence := make([]<-chan []byte, numBars)
			for i, b := range bars {
				sequence[i] = b.render(recoverFn, width, prependWs, appendWs)
			}

			ch := fanIn(sequence...)

			for buf := range ch {
				conf.cw.Write(buf)
			}

			conf.cw.Flush()

			for _, b := range bars {
				b.flushed()
			}
		case <-p.beforeStop:
			for _, b := range bars {
				if b.GetStatistics().Total <= 0 {
					b.Completed()
				}
			}
		case <-conf.cancel:
			return
		}
	}
}

func newWidthSync(quit <-chan struct{}, numBars, numColumn int) *widthSync {
	ws := &widthSync{
		listen: make([]chan int, numColumn),
		result: make([]chan int, numColumn),
	}
	for i := 0; i < numColumn; i++ {
		ws.listen[i] = make(chan int, numBars)
		ws.result[i] = make(chan int, numBars)
	}
	for i := 0; i < numColumn; i++ {
		go func(listenCh <-chan int, resultCh chan<- int) {
			defer close(resultCh)
			widths := make([]int, 0, numBars)
		loop:
			for {
				select {
				case w := <-listenCh:
					widths = append(widths, w)
					if len(widths) == numBars {
						break loop
					}
				case <-quit:
					if len(widths) == 0 {
						return
					}
					break loop
				}
			}
			result := max(widths)
			for i := 0; i < len(widths); i++ {
				resultCh <- result
			}
		}(ws.listen[i], ws.result[i])
	}
	return ws
}

func fanIn(inputs ...<-chan []byte) <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)
		for _, input := range inputs {
			ch <- <-input
		}
	}()

	return ch
}

// isClosed check if ch closed
// caution see: http://www.tapirgames.com/blog/golang-channel-closing
func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func max(slice []int) int {
	max := slice[0]

	for i := 1; i < len(slice); i++ {
		if slice[i] > max {
			max = slice[i]
		}
	}

	return max
}
