# Multi Progress Bar for Go [![Build Status](https://travis-ci.org/vbauerster/mpb.svg?branch=master)](https://travis-ci.org/vbauerster/mpb)

Mutex free progress bar library, for console programs.

It is inspired by [uiprogress](https://github.com/gosuri/uiprogress) library,
but unlike the last one, implementation is mutex free, following Go's idiom:

> Don't communicate by sharing memory, share memory by communicating.

## Features

* __Multiple Bars__: mpb can render multiple progress bars that can be tracked concurrently
* __Cancellable__: cancel rendering goroutine at any time
* __Dynamic Addition__:  Add additional progress bar at any time
* __Dynamic Removal__:  Remove rendering progress bar at any time
* __Dynamic Sorting__:  Sort bars by progression
* __Dynamic Resize__:  Resize bars on terminal width change
* __Custom Decorator Functions__: Add custom functions around the bar along with helper functions
* __Predefined Decoratros__: Elapsed time, [Ewmaest](https://github.com/dgryski/trifles/tree/master/ewmaest) based ETA, Percentage, Bytes counter

## Usage

Following is the simplest use case:

```go
	name := "Single bar:"
	// Star mpb's rendering goroutine.
	// If you don't plan to cancel, feed with nil
	// otherwise provide context.Context, see cancel example
	p := mpb.New(nil)
	bar := p.AddBar(100).PrependName(name, 0).AppendPercentage()

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()
```
The source code: [example/singleBar/main.go](example/singleBar/main.go)

However **mpb** was designed with concurrency in mind, each new bar renders in its
own goroutine. Therefore adding multiple bars is easy and safe:

```go
	var wg sync.WaitGroup
	p := mpb.New(nil)
	for i := 0; i < 3; i++ {
		wg.Add(1) // add wg delta
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name)).AppendPercentage()
		go func() {
			defer wg.Done()
			// you can p.AddBar() here, but ordering will be non deterministic
			for i := 0; i < 100; i++ {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}
	wg.Wait() // Wait for goroutines to finish
	p.Stop()  // Stop mpb's rendering goroutine
	// p.AddBar(1) // panic: you cannot reuse p, create new one!
	fmt.Println("finish")
```

This will produce following:

![example](example/gifs/simple.gif)

The source code: [example/simple/main.go](example/simple/main.go)

### Cancel

![example](example/gifs/cancel.gif)

The source code: [example/cancel/main.go](example/cancel/main.go)

### Removing bar

![example](example/gifs/remove.gif)

The source code: [example/remove/main.go](example/remove/main.go)

### Sorting bars by progress

![example](example/gifs/sort.gif)

The source code: [example/sort/main.go](example/sort/main.go)

### Resizing bars on terminal width change

![example](example/gifs/resize.gif)

The source code: [example/prependETA/main.go](example/prependETA/main.go)

### Multiple io

![example](example/gifs/io-multiple.gif)

The source code: [example/io/multiple/main.go](example/io/multiple/main.go)

## Installation

```sh
$ go get -u github.com/vbauerster/mpb
```

## License

This lib is under [WTFPL license](http://www.wtfpl.net)
