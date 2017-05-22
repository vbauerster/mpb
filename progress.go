package mpb

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/cwriter"
)

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

	done       chan struct{}
	userConfCh chan userConf
	bCommandCh chan *bCommandData
	barCountCh chan int
	stopReqCh  chan struct{}

	// follawing is used after (*Progress.done) is closed
	conf userConf
}

// New creates new Progress instance, which will orchestrate bars rendering
// process. It acceepts context.Context, for cancellation.
// If you don't plan to cancel, it is safe to feed with nil
func New() *Progress {
	p := &Progress{
		wg:         new(sync.WaitGroup),
		done:       make(chan struct{}),
		userConfCh: make(chan userConf),
		bCommandCh: make(chan *bCommandData),
		barCountCh: make(chan int),
		stopReqCh:  make(chan struct{}),
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
// Pancis, if nil channel is passed.
func (p *Progress) WithCancel(ch <-chan struct{}) *Progress {
	if ch == nil {
		panic("nil cancel channel")
	}
	p.updateConf(func(c *userConf) {
		c.cancel = ch
	})
	return p
}

// SetWidth overrides default (80) width of bar(s).
func (p *Progress) SetWidth(width int) *Progress {
	if width < 2 {
		return p
	}
	p.updateConf(func(c *userConf) {
		c.width = width
	})
	return p
}

// SetOut sets underlying writer of progress. Default one is os.Stdout.
func (p *Progress) SetOut(w io.Writer) *Progress {
	if w == nil {
		return p
	}
	p.updateConf(func(c *userConf) {
		c.cw.Flush()
		c.cw = cwriter.New(w)
	})
	return p
}

// RefreshRate overrides default (100ms) refresh rate value
func (p *Progress) RefreshRate(d time.Duration) *Progress {
	p.updateConf(func(c *userConf) {
		c.ticker.Stop()
		c.ticker = time.NewTicker(d)
		rr = d
	})
	return p
}

// BeforeRenderFunc accepts a func, which gets called before render process.
func (p *Progress) BeforeRenderFunc(f BeforeRender) *Progress {
	p.updateConf(func(c *userConf) {
		c.beforeRender = f
	})
	return p
}

// AddBar creates a new progress bar and adds to the container.
func (p *Progress) AddBar(total int64) *Bar {
	return p.AddBarWithID(0, total)
}

// AddBarWithID creates a new progress bar and adds to the container.
func (p *Progress) AddBarWithID(id int, total int64) *Bar {
	conf := p.getConf()
	bar := newBar(id, total, p.wg, &conf)
	p.bCommandCh <- &bCommandData{
		action: bAdd,
		bar:    bar,
	}
	return bar
}

// RemoveBar removes bar at any time.
func (p *Progress) RemoveBar(b *Bar) bool {
	result := make(chan bool)
	select {
	case p.bCommandCh <- &bCommandData{bRemove, b, result}:
		return <-result
	case <-p.done:
		return false
	}
}

// BarCount returns bars count in the container.
func (p *Progress) BarCount() int {
	select {
	case count := <-p.barCountCh:
		return count
	case <-p.done:
		return 0
	}
}

// ShutdownNotify means to be notified when main rendering goroutine quits, usualy after p.Stop() call.
func (p *Progress) ShutdownNotify(ch chan struct{}) *Progress {
	p.updateConf(func(c *userConf) {
		c.shutdownNotifier = ch
	})
	return p
}

// Format sets custom format for underlying bar(s), default one is "[=>-]".
func (p *Progress) Format(format string) *Progress {
	if utf8.RuneCountInString(format) != numFmtRunes {
		return p
	}
	p.updateConf(func(c *userConf) {
		c.format = format
	})
	return p
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
		p.wg.Wait() // wait for all bars to quit
		p.stopReqCh <- struct{}{}
		<-p.done // wait for p.server to quit
	}
}

func (p *Progress) getConf() userConf {
	select {
	case conf := <-p.userConfCh:
		return conf
	case <-p.done:
		return p.conf
	}
}

func (p *Progress) updateConf(cb func(*userConf)) {
	c := p.getConf()
	cb(&c)
	select {
	case p.userConfCh <- c:
	case <-p.done:
		return
	}
}

// server monitors underlying channels and renders any progress bars
func (p *Progress) server(conf userConf) {

	defer func() {
		p.conf = conf
		conf.ticker.Stop()
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
		}
		close(p.done)
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
		case p.userConfCh <- conf:
		case conf = <-p.userConfCh:
		case data := <-p.bCommandCh:
			switch data.action {
			case bAdd:
				bars = append(bars, data.bar)
				p.wg.Add(1)
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
		case p.barCountCh <- len(bars):
		case <-conf.ticker.C:
			var notick bool
			select {
			// stop ticking if cancel requested
			case <-conf.cancel:
				conf.ticker.Stop()
				notick = true
			default:
			}

			numBars := len(bars)
			if notick || numBars == 0 {
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
		case <-p.stopReqCh:
			for _, b := range bars {
				if b.GetStatistics().Total <= 0 {
					b.Completed()
				}
			}
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

func max(slice []int) int {
	max := slice[0]

	for i := 1; i < len(slice); i++ {
		if slice[i] > max {
			max = slice[i]
		}
	}

	return max
}
