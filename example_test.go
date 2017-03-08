package mpb_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
)

func Example() {
	// Star mpb's rendering goroutine.
	// If you don't plan to cancel, feed with nil
	// otherwise provide context.Context, see cancel example
	p := mpb.New(nil)
	// Set custom width for every bar, which mpb will contain
	// The default one in 70
	p.SetWidth(80)
	// Set custom format for every bar, the default one is "[=>-]"
	p.Format("╢▌▌░╟")
	// Set custom refresh rate, the default one is 100 ms
	p.RefreshRate(120 * time.Millisecond)

	// Add a bar. You're not limited to just one bar, add many if you need.
	bar := p.AddBar(100).
		PrependName("Single Bar:", 0, 0).
		AppendPercentage(5, 0)

	for i := 0; i < 100; i++ {
		bar.Incr(1) // increment progress bar
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	// Don't forget to stop mpb's rendering goroutine
	p.Stop()

	// You cannot add bars after p.Stop() has been called
	// p.AddBar(100) // will panic
}

func ExampleBar_InProgress() {
	p := mpb.New(nil)
	bar := p.AddBar(100).AppendPercentage(5, 0)

	for bar.InProgress() {
		bar.Incr(1)
		time.Sleep(time.Millisecond * 20)
	}
}

func ExampleBar_PrependFunc() {
	decor := func(s *mpb.Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprintf("%d/%d", s.Current, s.Total)
		return fmt.Sprintf("%8s", str)
	}

	totalItem := 100
	p := mpb.New(nil)
	bar := p.AddBar(int64(totalItem)).PrependFunc(decor)

	for i := 0; i < totalItem; i++ {
		bar.Incr(1) // increment progress bar
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
}

func ExampleBar_AppendFunc() {
	decor := func(s *mpb.Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprintf("%d/%d", s.Current, s.Total)
		return fmt.Sprintf("%8s", str)
	}

	totalItem := 100
	p := mpb.New(nil)
	bar := p.AddBar(int64(totalItem)).AppendFunc(decor)

	for i := 0; i < totalItem; i++ {
		bar.Incr(1) // increment progress bar
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
}
