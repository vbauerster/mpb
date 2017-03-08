//+build go1.7

package mpb

import "context"

// WithContext cancellation via context
func (p *Progress) WithContext(ctx context.Context) *Progress {
	if ctx == nil {
		panic("nil context")
	}
	p2 := new(Progress)
	*p2 = *p
	p2.cancel = ctx.Done()
	return p2
}
