package uiprogress

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gosuri/uilive"
)

// Out is the default writer to render progress bars to
var Out = os.Stdout

// RefreshInterval in the default time duration to wait for refreshing the output
var RefreshInterval = time.Millisecond * 10

// Progress represents the container that renders progress bars
type Progress struct {
	// Out is the writer to render progress bars to
	Out io.Writer

	// Width is the width of the progress bars
	// Width int

	// Bars is the collection of progress bars
	// Bars []*Bar

	// RefreshInterval in the time duration to wait for refreshing the output
	// RefreshInterval time.Duration

	lw *uilive.Writer
	// stopChan chan struct{}
	// mtx      *sync.RWMutex
	bars   chan *Bar
	ticker *time.Ticker
}

// New returns a new progress bar with defaults
func New() *Progress {
	p := &Progress{
		Out:    Out,
		lw:     uilive.New(),
		bars:   make(chan *Bar),
		ticker: time.NewTicker(RefreshInterval),
	}
	go p.server()
	return p
}

// AddBar creates a new progress bar and adds to the container
func (p *Progress) AddBar(total int) *Bar {
	bar := NewBar(total)
	// bar.Width = p.Width
	p.bars <- bar
	return bar
}

// Listen listens for updates and renders the progress bars
func (p *Progress) server() {
	bars := make([]*Bar, 0)
	p.lw.Out = p.Out
loop:
	for {
		select {
		case bar, ok := <-p.bars:
			if !ok {
				break loop
			}
			bars = append(bars, bar)
		case <-p.ticker.C:
			for _, bar := range bars {
				fmt.Fprintln(p.lw, bar.String())
			}
			p.lw.Flush()
		}
	}
}
