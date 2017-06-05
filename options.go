package mpb

import (
	"io"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/cwriter"
)

type ProgressOption func(*pConf)

// WithWidth overrides default width 80
func WithWidth(w int) ProgressOption {
	return func(c *pConf) {
		if w > 2 {
			c.width = w
		}
	}
}

// WithFormat overrides default bar format "[=>-]"
func WithFormat(format string) ProgressOption {
	return func(c *pConf) {
		if utf8.RuneCountInString(format) == formatLen {
			c.format = format
		}
	}
}

// WithRefreshRate overrides default 100ms refresh rate
func WithRefreshRate(d time.Duration) ProgressOption {
	return func(c *pConf) {
		c.ticker.Stop()
		c.ticker = time.NewTicker(d)
		c.rr = d
	}
}

// WithBeforeRenderFunc provided BeforeRender func,
// will be called before each render cycle.
func WithBeforeRenderFunc(f BeforeRender) ProgressOption {
	return func(c *pConf) {
		c.beforeRender = f
	}
}

// WithCancel provide your cancel channel,
// which you plan to close at some point.
func WithCancel(ch <-chan struct{}) ProgressOption {
	return func(c *pConf) {
		c.cancel = ch
	}
}

// WithShutdownNotifier provided chanel will be closed, inside p.Stop() call
func WithShutdownNotifier(ch chan struct{}) ProgressOption {
	return func(c *pConf) {
		c.shutdownNotifier = ch
	}
}

// Output overrides default output os.Stdout
func Output(w io.Writer) ProgressOption {
	return func(c *pConf) {
		if w != nil {
			c.cw = cwriter.New(w)
		}
	}
}
