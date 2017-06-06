//+build go1.7

package mpb

import "context"

func WithContext(ctx context.Context) ProgressOption {
	return func(c *pConf) {
		if ctx != nil {
			c.cancel = ctx.Done()
		}
	}
}
