package mpb_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func TestDefaultWidth(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(mpb.Output(&buf))
	bar := p.AddBar(100, mpb.BarTrim())

	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()

	wantWidth := 80
	gotWidth := utf8.RuneCount(buf.Bytes())
	if gotWidth != wantWidth+1 { // + 1 for new line
		t.Errorf("Expected default width: %d, got: %d\n", wantWidth, gotWidth)
	}
}

func TestCustomWidth(t *testing.T) {
	wantWidth := 60
	var buf bytes.Buffer
	p := mpb.New(mpb.WithWidth(wantWidth), mpb.Output(&buf))
	bar := p.AddBar(100, mpb.BarTrim())

	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()

	gotWidth := utf8.RuneCount(buf.Bytes())
	if gotWidth != wantWidth+1 { // +1 for new line
		t.Errorf("Expected default width: %d, got: %d\n", wantWidth, gotWidth)
	}
}

func TestInvalidWidth(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(mpb.WithWidth(1), mpb.Output(&buf))
	bar := p.AddBar(100, mpb.BarTrim())

	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()

	wantWidth := 80
	gotWidth := utf8.RuneCount(buf.Bytes())
	if gotWidth != wantWidth+1 { // +1 for new line
		t.Errorf("Expected width: %d, got: %d\n", wantWidth, gotWidth)
	}
}

func TestAddBar(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := mpb.New(mpb.Output(&buf))

	count := p.BarCount()
	if count != 0 {
		t.Errorf("BarCount want: %q, got: %q\n", 0, count)
	}

	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100, mpb.PrependDecorators(decor.Name(name, len(name), 0)))

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				bar.Incr(1)
			}
		}()
	}

	count = p.BarCount()
	if count != numBars {
		t.Errorf("BarCount want: %q, got: %q\n", numBars, count)
	}
	wg.Wait()
	p.Stop()
}

func TestRemoveBar(t *testing.T) {
	p := mpb.New()

	b := p.AddBar(10)

	if !p.RemoveBar(b) {
		t.Error("RemoveBar failure")
	}

	count := p.BarCount()
	if count != 0 {
		t.Errorf("BarCount want: %q, got: %q\n", 0, count)
	}
	p.Stop()
}

func TestWithCancel(t *testing.T) {
	cancel := make(chan struct{})
	shutdown := make(chan struct{})
	p := mpb.New(mpb.WithCancel(cancel), mpb.WithShutdownNotifier(shutdown))

	var wg sync.WaitGroup
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total), mpb.BarID(i),
			mpb.PrependDecorators(decor.Name(name, len(name), 0)))

		go func() {
			defer func() {
				// fmt.Printf("%s done\n", name)
				wg.Done()
			}()
			for i := 0; i < total; i++ {
				select {
				case <-cancel:
					return
				default:
				}
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}

	time.AfterFunc(300*time.Millisecond, func() {
		close(cancel)
	})

	wg.Wait()
	p.Stop()

	select {
	case <-shutdown:
	case <-time.After(300 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}

func TestCustomFormat(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	customFormat := "╢▌▌░╟"
	p := mpb.New(mpb.Output(&buf), mpb.WithCancel(cancel), mpb.WithFormat(customFormat))
	bar := p.AddBar(100, mpb.BarTrim())

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(10 * time.Millisecond)
			bar.Incr(1)
		}
	}()

	time.Sleep(250 * time.Millisecond)
	close(cancel)
	p.Stop()

	bytes := removeLastRune(buf.Bytes())

	seen := make(map[rune]bool)
	for _, r := range string(bytes) {
		if !seen[r] {
			seen[r] = true
		}
	}
	for _, r := range customFormat {
		if !seen[r] {
			t.Errorf("Rune %#U not found in bar\n", r)
		}
	}
}
