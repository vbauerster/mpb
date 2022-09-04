package mpb

import (
	"io"

	"github.com/vbauerster/mpb/v8/decor"
)

// BarFiller interface.
// Bar (without decorators) renders itself by calling BarFiller's Fill method.
type BarFiller interface {
	Fill(w io.Writer, stat decor.Statistics)
}

// BarFillerBuilder interface.
// Default implementations are:
//
//	BarStyle()
//	SpinnerStyle()
//	NopStyle()
type BarFillerBuilder interface {
	Build() BarFiller
}

// BarFillerFunc is function type adapter to convert compatible function
// into BarFiller interface.
type BarFillerFunc func(w io.Writer, stat decor.Statistics)

func (f BarFillerFunc) Fill(w io.Writer, stat decor.Statistics) {
	f(w, stat)
}

// BarFillerBuilderFunc is function type adapter to convert compatible
// function into BarFillerBuilder interface.
type BarFillerBuilderFunc func() BarFiller

func (f BarFillerBuilderFunc) Build() BarFiller {
	return f()
}
