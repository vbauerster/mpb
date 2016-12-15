# Multi Progress Bar for Go [![Build Status](https://travis-ci.org/vbauerster/mpb.svg?branch=master)](https://travis-ci.org/vbauerster/mpb)

Mutex free progress bar library, for console programs.

It is inspired by [uiprogress](https://github.com/gosuri/uiprogress) library,
but unlike the last one, implementation is mutex free, following Go's idiom:

> Don't communicate by sharing memory, share memory by communicating.

## Features

* __Multiple Bars__: mpb can render multiple progress bars that can be tracked concurrently
* __Dynamic Addition__:  Add additional progress bar at any time
* __Dynamic Removal__:  Remove rendering progress bar at any time
* __Dynamic Sorting__:  Sort bars by progression
* __Custom Decorator Functions__: Add custom functions around the bar along with helper functions
* __Predefined Decoratros__: Elapsed time, [Ewmaest](https://github.com/dgryski/trifles/tree/master/ewmaest) based ETA, Percentage, Bytes counter

## Usage

Following is the simplest use case:

```go
	// No need to initialize sync.WaitGroup, as it is initialized implicitly
	p := mpb.New() // Star mpb container
	for i := 0; i < 3; i++ {
		p.Wg.Add(1) // add wg counter
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name)).AppendPercentage()
		go func() {
			// you can p.AddBar() here, but ordering will be non deterministic
			for i := 0; i < 100; i++ {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}
	p.WaitAndStop() // Wait for goroutines to finish
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("finish")
```

This will produce following:

![example](example/gifs/simple.gif)
The source code: [example/simple/main.go](example/simple/main.go)

### Removing bar

![example](example/gifs/remove.gif)

The source code: [example/remove/main.go](example/remove/main.go)

### Sorting bars by progress

![example](example/gifs/sort.gif)

The source code: [example/sort/main.go](example/sort/main.go)

### Multiple io

![example](example/gifs/io-multiple.gif)

The source code: [example/io/multiple/main.go](example/io/multiple/main.go)

## Installation

```sh
$ go get -u github.com/vbauerster/mpb
```
