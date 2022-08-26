package mpb

import (
	"io"

	"github.com/vbauerster/mpb/v7/decor"
)

// BarFiller interface.
// Bar (without decorators) renders itself by calling BarFiller's Fill method.
//
//	If not set, it defaults to terminal width.
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

// NewBarFiller constructs a BarFiller from provided BarFillerBuilder.
// Deprecated. Prefer using `*Progress.New(...)` directly.
func NewBarFiller(b BarFillerBuilder) BarFiller {
	return b.Build()
}
