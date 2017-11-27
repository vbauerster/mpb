package mpb_test

import (
	"bytes"
	"fmt"
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

		gotWidth := len(buf.Bytes())
		if gotWidth != tc.expected {
			t.Errorf("%s: Expected width: %d, got: %d\n", k, tc.expected, gotWidth)
		}
	}
}

func TestBarFormat(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	customFormat := "(#>_)"
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithCancel(cancel),
		mpb.WithFormat(customFormat),
	)
	bar := p.AddBar(100, mpb.BarTrim())

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(10 * time.Millisecond)
			bar.Increment()
		}
	}()

	time.Sleep(250 * time.Millisecond)
	close(cancel)
	p.Stop()

	barAsStr := strings.Trim(buf.String(), "\n")

	seen := make(map[rune]bool)
	for _, r := range barAsStr {
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

func TestBarInvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	customWidth := 60
	customFormat := "(#>=_)"
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithWidth(customWidth),
		mpb.WithFormat(customFormat),
	)
	bar := p.AddBar(100, mpb.BarTrim())

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	got := buf.String()
	want := fmt.Sprintf("[%s]", strings.Repeat("=", customWidth-2))
	if !strings.Contains(got, want) {
		t.Errorf("Expected format: %s, got %s\n", want, got)
	}
}

func TestBarInProgress(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithCancel(cancel),
	)
	bar := p.AddBar(100, mpb.BarTrim())

	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		for bar.InProgress() {
			time.Sleep(10 * time.Millisecond)
			bar.Incr(1)
		}
	}()

	time.Sleep(250 * time.Millisecond)
	close(cancel)
	p.Stop()

	select {
	case <-stopped:
	case <-time.After(300 * time.Millisecond):
		t.Error("bar.InProgress returns true after cancel")
	}
}

func TestBarGetID(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := mpb.New(mpb.Output(&buf))

	numBars := 3
	wg.Add(numBars)

	bars := make([]*mpb.Bar, numBars)
	for i := 0; i < numBars; i++ {
		bars[i] = p.AddBar(100, mpb.BarID(i))

		go func(bar *mpb.Bar) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Incr(1)
			}
		}(bars[i])
	}

	for wantID, bar := range bars {
		gotID := bar.ID()
		if gotID != wantID {
			t.Errorf("Expected bar id: %d, got %d\n", wantID, gotID)
		}
	}

	wg.Wait()
	p.Stop()
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
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	bytes := removeLastRune(buf.Bytes())

	gotBar := string(bytes[len(bytes)-width:])
	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(string(refillChar), till-1),
		strings.Repeat("=", total-till-1))
	if gotBar != wantBar {
		t.Errorf("Want bar: %s, got bar: %s\n", wantBar, gotBar)
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

	out := bytes.Split(removeLastRune(buf.Bytes()), []byte("\n"))
	gotPanic := out[len(out)-1]
	wantPanic = fmt.Sprintf("b#%02d panic: %v", 2, wantPanic)

	if string(gotPanic) != wantPanic {
		t.Errorf("Want: %q, got: %q\n", wantPanic, gotPanic)
	}
}

func removeLastRune(bytes []byte) []byte {
	_, size := utf8.DecodeLastRune(bytes)
	return bytes[:len(bytes)-size]
}
