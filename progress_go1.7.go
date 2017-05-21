//+build go1.7

package mpb

import "context"

// WithContext cancellation via context.
// pancis, if called on stopped Progress instance, i.e after (*Progress).Stop()
// or nil context passed
func (p *Progress) WithContext(ctx context.Context) *Progress {
	if isClosed(p.done) {
		panic(ErrCallAfterStop)
	}
	if ctx == nil {
		panic("nil context")
	}
	conf := <-p.userConf
	conf.cancel = ctx.Done()
	p.userConf <- conf
	return p
}
