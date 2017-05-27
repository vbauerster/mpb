package mpb

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/cwriter"
)

type (
	// BeforeRender is a func, which gets called before render process
	BeforeRender func([]*Bar)

	widthSync struct {
		listen []chan int
		result []chan int
	}

	// progress config, all fieals are adjustable by user
	pConf struct {
		bars []*Bar

		width        int
		format       string
		rr           time.Duration
		cw           *cwriter.Writer
		ticker       *time.Ticker
		beforeRender BeforeRender

		shutdownNotifier chan struct{}
		cancel           <-chan struct{}
	}
)

const (
	// default RefreshRate
	prr = 100 * time.Millisecond
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

	done      chan struct{}
	ops       chan func(*pConf)
	stopReqCh chan struct{}

	// following is used after (*Progress.done) is closed
	conf pConf
}

// New creates new Progress instance, which will orchestrate bars rendering
// process. It acceepts context.Context, for cancellation.
// If you don't plan to cancel, it is safe to feed with nil
func New() *Progress {
	p := &Progress{
		wg:        new(sync.WaitGroup),
		done:      make(chan struct{}),
		ops:       make(chan func(*pConf)),
		stopReqCh: make(chan struct{}),
	}
	go p.server(pConf{
		bars:   make([]*Bar, 0, 3),
		width:  pwidth,
		format: pformat,
		cw:     cwriter.New(os.Stdout),
		rr:     prr,
		ticker: time.NewTicker(prr),
	})
	return p
}

// WithCancel cancellation via channel.
// You have to call p.Stop() anyway, after cancel.
// Pancis, if nil channel is passed.
func (p *Progress) WithCancel(ch <-chan struct{}) *Progress {
	if ch == nil {
		panic("nil cancel channel")
	}
	return updateConf(p, func(c *pConf) {
		c.cancel = ch
	})
}

// SetWidth overrides default (80) width of bar(s).
func (p *Progress) SetWidth(width int) *Progress {
	if width < 2 {
		return p
	}
	return updateConf(p, func(c *pConf) {
		c.width = width
	})
}

// SetOut sets underlying writer of progress. Default one is os.Stdout.
func (p *Progress) SetOut(w io.Writer) *Progress {
	if w == nil {
		return p
	}
	return updateConf(p, func(c *pConf) {
		c.cw.Flush()
		c.cw = cwriter.New(w)
	})
}

// RefreshRate overrides default (100ms) refresh rate value
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	return updateConf(p, func(c *pConf) {
		c.ticker.Stop()
		c.ticker = time.NewTicker(d)
		c.rr = d
	})
}

// BeforeRenderFunc accepts a func, which gets called before render process.
func (p *Progress) BeforeRenderFunc(f BeforeRender) *Progress {
	return updateConf(p, func(c *pConf) {
		c.beforeRender = f
	})
}

// AddBar creates a new progress bar and adds to the container.
func (p *Progress) AddBar(total int64) *Bar {
	return p.AddBarWithID(0, total)
}

// AddBarWithID creates a new progress bar and adds to the container.
func (p *Progress) AddBarWithID(id int, total int64) *Bar {
	result := make(chan *Bar, 1)
	op := func(c *pConf) {
		bar := newBar(id, total, c.width, c.format, p.wg, c.cancel)
		c.bars = append(c.bars, bar)
		p.wg.Add(1)
		result <- bar
	}
	select {
	case p.ops <- op:
		return <-result
	case <-p.done:
		return nil
	}
}

// RemoveBar removes bar at any time.
func (p *Progress) RemoveBar(b *Bar) bool {
	result := make(chan bool, 1)
	op := func(c *pConf) {
		var ok bool
		for i, bar := range c.bars {
			if bar == b {
				c.bars = append(c.bars[:i], c.bars[i+1:]...)
				bar.remove()
				ok = true
				break
			}
		}
		result <- ok
	}
	select {
	case p.ops <- op:
		return <-result
	case <-p.done:
		return false
	}
}

// BarCount returns bars count
func (p *Progress) BarCount() int {
	result := make(chan int, 1)
	op := func(c *pConf) {
		result <- len(c.bars)
	}
	select {
	case p.ops <- op:
		return <-result
	case <-p.done:
		return 0
	}
}

// ShutdownNotify means to be notified when main rendering goroutine quits, usualy after p.Stop() call.
func (p *Progress) ShutdownNotify(ch chan struct{}) *Progress {
	return updateConf(p, func(c *pConf) {
		c.shutdownNotifier = ch
	})
}

// Format sets custom format for underlying bar(s), default one is "[=>-]".
func (p *Progress) Format(format string) *Progress {
	if utf8.RuneCountInString(format) != numFmtRunes {
		return p
	}
	return updateConf(p, func(c *pConf) {
		c.format = format
	})
}

// Stop shutdowns Progress' goroutine.
// Should be called only after each bar's work done, i.e. bar has reached its
// 100 %. It is NOT for cancelation. Use WithContext or WithCancel for
// cancelation purposes.
func (p *Progress) Stop() {
	select {
	case <-p.done:
		return
	default:
		// complete Total unknown bars
		p.ops <- func(c *pConf) {
			for _, b := range c.bars {
				s := b.getState()
				if !s.completed && !s.aborted {
					b.Complete()
				}
			}
		}
		// wait for all bars to quit
		p.wg.Wait()
		// stop request
		p.stopReqCh <- struct{}{}
		// wait for p.server to quit
		<-p.done
	}
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(conf pConf) {

	defer func() {
		p.conf = conf
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
		}
		close(p.done)
	}()

	recoverFn := func(ch chan []byte) {
		if p := recover(); p != nil {
			ch <- []byte(fmt.Sprintln(p))
		}
		close(ch)
	}

	for {
		select {
		case op := <-p.ops:
			op(&conf)
		case <-conf.ticker.C:
			numBars := len(conf.bars)
			if numBars == 0 {
				break
			}

			if conf.beforeRender != nil {
				conf.beforeRender(conf.bars)
			}

			quitWidthSyncCh := make(chan struct{})
			time.AfterFunc(conf.rr, func() {
				close(quitWidthSyncCh)
			})

			b0 := conf.bars[0]
			prependWs := newWidthSync(quitWidthSyncCh, numBars, b0.NumOfPrependers())
			appendWs := newWidthSync(quitWidthSyncCh, numBars, b0.NumOfAppenders())

			width, _, _ := cwriter.GetTermSize()

			sequence := make([]<-chan []byte, numBars)
			for i, b := range conf.bars {
				sequence[i] = b.render(recoverFn, width, prependWs, appendWs)
			}

			ch := fanIn(sequence...)

			for buf := range ch {
				conf.cw.Write(buf)
			}

			conf.cw.Flush()

			for _, b := range conf.bars {
				b.flushed()
			}
		case <-conf.cancel:
			conf.ticker.Stop()
			conf.cancel = nil
		case <-p.stopReqCh:
			conf.ticker.Stop()
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

func updateConf(p *Progress, op func(*pConf)) *Progress {
	select {
	case p.ops <- op:
		return p
	case <-p.done:
		return nil
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
