// Copyright (C) 2016-2018 Vladimir Bauer
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
 Package decor contains common decorators for "github.com/vbauerster/mpb" package.

 All decorators returned by this package have some closure state. It is ok to use
 decorators concurrently, unless you share the same decorator among multiple
 *mpb.Bar instances. To avoid data races, create new decorator per *mpb.Bar
 instance.

 Don't:

	 p := mpb.New()
	 eta := decor.ETA(0, 0)
	 p.AddBar(100, mpb.AppendDecorators(eta))
	 p.AddBar(100, mpb.AppendDecorators(eta))

 Do:

	p := mpb.New()
	p.AddBar(100, mpb.AppendDecorators(decor.ETA(0, 0)))
	p.AddBar(100, mpb.AppendDecorators(decor.ETA(0, 0)))
*/
package decor
