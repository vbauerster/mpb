//+build go1.7

// Copyright (C) 2016-2017 Vladimir Bauer
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
