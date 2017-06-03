//+build go1.7

package mpb

import "context"

// WithContext Deprecated, use mpb.WithContext
func (p *Progress) WithContext(ctx context.Context) *Progress {
	if ctx == nil {
		panic("nil context")
	}
	return updateConf(p, func(c *pConf) {
		c.cancel = ctx.Done()
	})
}
