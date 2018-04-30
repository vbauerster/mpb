# Multi Progress Bar

[![GoDoc](https://godoc.org/github.com/vbauerster/mpb?status.svg)](https://godoc.org/github.com/vbauerster/mpb)
[![Build Status](https://travis-ci.org/vbauerster/mpb.svg?branch=master)](https://travis-ci.org/vbauerster/mpb)
[![Go Report Card](https://goreportcard.com/badge/github.com/vbauerster/mpb)](https://goreportcard.com/report/github.com/vbauerster/mpb)
[![codecov](https://codecov.io/gh/vbauerster/mpb/branch/master/graph/badge.svg)](https://codecov.io/gh/vbauerster/mpb)

**mpb** is a Go lib for rendering progress bars in terminal applications.

## Features

* __Multiple Bars__: Multiple progress bars are supported
* __Dynamic Total__: [Set total](https://github.com/vbauerster/mpb/issues/9#issuecomment-344448984) while bar is running
* __Dynamic Add/Remove__: Dynamically add or remove bars
* __Cancellation__: Cancel whole rendering process
* __Predefined Decorators__: Elapsed time, [Ewmaest](https://github.com/dgryski/trifles/tree/master/ewmaest) based ETA, Percentage, Bytes counter
* __Decorator's width sync__:  Synchronized decorator's width among multiple bars

## Installation

```sh
go get github.com/vbauerster/mpb
```

_Note:_ it is preferable to go get from github.com, rather than gopkg.in. See issue [#11](https://github.com/vbauerster/mpb/issues/11).

## Usage

#### [Rendering single bar](examples/singleBar/main.go)
```go
	p := mpb.New(
		// override default (80) width
		mpb.WithWidth(64),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	total := 100
	name := "Single Bar:"
	// adding a single bar
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(
			// Display our static name with one space on the right
			decor.StaticName(name, len(name)+1, decor.DidentRight),
			// ETA decorator with width reservation of 3 runes
			decor.ETA(3, 0),
		),
		mpb.AppendDecorators(
			// Percentage decorator with width reservation of 5 runes
			decor.Percentage(5, 0),
		),
	)

	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		// increment by 1 (there is bar.IncrBy(int) method, if needed)
		bar.Increment()
	}
	// wait for our bar to complete and flush
	p.Wait()
```

#### [Rendering multiple bars](examples/simple/main.go)
```go
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// Display our static name with one space on the right
				decor.StaticName(name, len(name)+1, decor.DidentRight),
				// DwidthSync bit enables same column width synchronization
				decor.Percentage(0, decor.DwidthSync),
			),
			mpb.AppendDecorators(
				// Replace our ETA decorator with "done!", on bar completion event
				decor.OnComplete(decor.ETA(3, 0), "done!", 0, 0),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				bar.Increment()
			}
		}()
	}
	// first wait for provided wg, then
	// wait for all bars to complete and flush
	p.Wait()
```

#### [Dynamic total](examples/dynTotal/main.go)

![dynamic total](examples/gifs/1LuTsBJUAm4yV6PpT5OJSmJYw.svg)

#### [Complex example](examples/complex/main.go)

![complex](examples/gifs/ln3qeyH5iXUoesLYfw7zHDn6u.svg)

#### [Bytes counters](examples/io/multiple/main.go)

![byte counters](examples/gifs/ZsrT3r0ecrFwoarnplQz4UeL4.svg)
