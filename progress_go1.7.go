//+build go1.7

package mpb

import "context"

// New Progress instance, it orchestrates the rendering of progress bars.
// It supports cancellation via Context.
func NewWithCtx(ctx context.Context) *Progress {
	p := New()
	go func() {
		<-ctx.Done()
		p.Stop()
	}()
	return p
}
