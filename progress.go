package uiprogress

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gosuri/uilive"
)

// progress represents the container that renders progress bars
type progress struct {
	// out is the writer to render progress bars to
	out io.Writer

	// Width is the width of the progress bars
	// Width int

	lw *uilive.Writer

	// new Bars can be added over this channel
	bars chan *Bar

	// new refresh interval to be send over this channel
	interval chan time.Duration
}

// New returns a new progress bar with defaults
func New() *progress {
	p := &progress{
		out:      os.Stdout,
		lw:       uilive.New(),
		bars:     make(chan *Bar),
		interval: make(chan time.Duration),
	}
	go p.server()
	return p
}

// RefreshInterval overrides default interval value 30 ms
func (p *progress) RefreshInterval(d time.Duration) *progress {
	p.interval <- d
	return p
}

// SetOut sets underlying writer of progress
// default is os.Stdout
func (p *progress) SetOut(w io.Writer) *progress {
	p.out = w
	return p
}

// AddBar creates a new progress bar and adds to the container
func (p *progress) AddBar(total int) *Bar {
	bar := NewBar(total)
	// bar.Width = p.Width
	p.bars <- bar
	return bar
}

// Bypass returns a writer which allows non-buffered data to be written to the underlying output
func (p *progress) Bypass() io.Writer {
	return p.lw.Bypass()
}

// Stop stops listening
func (p *progress) Stop() {
	close(p.bars)
}

// server listens for updates and renders the progress bars
func (p *progress) server() {
	t := time.NewTicker(30 * time.Millisecond)
	bars := make([]*Bar, 0)
	p.lw.Out = p.out
	for {
		select {
		case bar, ok := <-p.bars:
			if !ok {
				t.Stop()
				return
			}
			bars = append(bars, bar)
		case <-t.C:
			for _, bar := range bars {
				fmt.Fprintln(p.lw, bar.String())
			}
			p.lw.Flush()
		case d := <-p.interval:
			t.Stop()
			t = time.NewTicker(d)
		}
	}
}
