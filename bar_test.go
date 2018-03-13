package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func TestWithWidth(t *testing.T) {
	cases := map[string]struct{ w, expected int }{
		"WithWidth-1": {-1, 81},
		"WithWidth0":  {0, 3},
		"WithWidth1":  {1, 3},
		"WithWidth2":  {2, 3},
		"WithWidth3":  {3, 4},
		"WithWidth60": {60, 61},
	}

	var buf bytes.Buffer
	for k, tc := range cases {
		buf.Reset()
		p := mpb.New(
			mpb.Output(&buf),
			mpb.WithWidth(tc.w),
		)
		bar := p.AddBar(10, mpb.BarTrim())

		for i := 0; i < 10; i++ {
			bar.Increment()
		}

		p.Stop()

		gotWidth := utf8.RuneCount(buf.Bytes())
		if gotWidth != tc.expected {
			t.Errorf("%s: Expected width: %d, got: %d\n", k, tc.expected, gotWidth)
		}
	}
}

func TestBarCompleted(t *testing.T) {
	p := mpb.New(
		mpb.Output(ioutil.Discard),
	)
	total := 80
	bar := p.AddBar(int64(total))

	var count int
	for !bar.Completed() {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
		count++
	}

	p.Stop()
	if count != total {
		t.Errorf("got count: %d, expected %d\n", count, total)
	}
}

func TestBarID(t *testing.T) {
	p := mpb.New(mpb.Output(ioutil.Discard))

	numBars := 3
	bars := make([]*mpb.Bar, numBars)
	for i := 0; i < numBars; i++ {
		bars[i] = p.AddBar(80, mpb.BarID(i))
		go func(bar *mpb.Bar) {
			for i := 0; i < 80; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Increment()
			}
		}(bars[i])
	}

	p.Stop()
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
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithWidth(width),
	)

	total := 100
	till := 30
	refillChar := '+'

	bar := p.AddBar(100, mpb.BarTrim())

	bar.ResumeFill(refillChar, int64(till))

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Stop()

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
	p := mpb.New(mpb.Output(&buf), mpb.WithWaitGroup(&wg))

	wantPanic := "Upps!!!"
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("b#%02d:", i)
		bar := p.AddBar(100, mpb.BarID(i), mpb.PrependDecorators(
			func(s *decor.Statistics, _ chan<- int, _ <-chan int) string {
				if s.ID == 2 && s.Current >= 42 {
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

	p.Stop()

	wantPanic = fmt.Sprintf("b#%02d panic: %v", 2, wantPanic)

	if !strings.Contains(buf.String(), wantPanic) {
		t.Errorf("Want: %q, got: %q\n", wantPanic, buf.String())
	}
}
