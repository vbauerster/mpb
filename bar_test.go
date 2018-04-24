package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func TestBarCompleted(t *testing.T) {
	p := New(WithOutput(ioutil.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	var count int
	for !bar.Completed() {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
		count++
	}

	p.Wait()
	if count != total {
		t.Errorf("got count: %d, expected %d\n", count, total)
	}
}

func TestBarID(t *testing.T) {
	p := New(WithOutput(ioutil.Discard))

	numBars := 3
	bars := make([]*Bar, numBars)
	for i := 0; i < numBars; i++ {
		bars[i] = p.AddBar(80, BarID(i))
		go func(bar *Bar) {
			for i := 0; i < 80; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Increment()
			}
		}(bars[i])
	}

	p.Wait()
	for wantID, bar := range bars {
		gotID := bar.ID()
		if gotID != wantID {
			t.Errorf("Expected bar id: %d, got %d\n", wantID, gotID)
		}
	}
}

func TestBarIncrWithReFill(t *testing.T) {
	var buf bytes.Buffer

	width := 100
	p := New(WithOutput(&buf), WithWidth(width))

	total := 100
	till := 30
	refillChar := '+'

	bar := p.AddBar(100, BarTrim())

	bar.ResumeFill(refillChar, int64(till))

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(string(refillChar), till-1),
		strings.Repeat("=", total-till-1))

	if !strings.Contains(buf.String(), wantBar) {
		t.Errorf("Want bar: %s, got bar: %s\n", wantBar, buf.String())
	}
}

func TestBarPanics(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := New(WithDebugOutput(&buf), WithOutput(nil), WithWaitGroup(&wg))

	wantPanic := "Upps!!!"
	numBars := 1
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("b#%02d:", i)
		bar := p.AddBar(100, PrependDecorators(
			func(s *decor.Statistics, _ chan<- int, _ <-chan int) string {
				if s.Current >= 42 {
					panic(wantPanic)
				}
				return name
			},
		))

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Increment()
			}
		}()
	}

	p.Wait()

	wantPanic = fmt.Sprintf("panic: %s", wantPanic)

	debugStr := buf.String()
	if !strings.Contains(debugStr, wantPanic) {
		t.Errorf("%q doesn't contain %q\n", debugStr, wantPanic)
	}
}
