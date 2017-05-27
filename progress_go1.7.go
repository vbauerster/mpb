//+build go1.7

package mpb

import "context"

// WithContext cancellation via context.
// You have to call p.Stop() anyway, after cancel.
// Pancis, if nil context is passed
func (p *Progress) WithContext(ctx context.Context) *Progress {
	if ctx == nil {
		panic("nil context")
	}
	return updateConf(p, func(c *pConf) {
		c.cancel = ctx.Done()
	})
}
