package mpb

import (
	"io"

	"github.com/vbauerster/mpb/v8/decor"
)

// BarFillerBuilderFunc is function type adapter to convert compatible
// function into BarFillerBuilder interface.
type BarFillerBuilderFunc func() BarFiller

func (f BarFillerBuilderFunc) Build() BarFiller {
	return f()
}

// NopStyle provides BarFillerBuilder which builds NOP BarFiller.
func NopStyle() BarFillerBuilder {
	return BarFillerBuilderFunc(func() BarFiller {
		return BarFillerFunc(func(io.Writer, decor.Statistics) error {
			return nil
		})
	})
}
