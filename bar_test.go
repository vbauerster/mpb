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
)

func TestBarSetWidth(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	// overwrite default width 80
	customWidth := 60
	bar := p.AddBar(100).SetWidth(customWidth).
		TrimLeftSpace().TrimRightSpace()
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()

	gotWidth := len(buf.Bytes())
	if gotWidth != customWidth+1 { // +1 for new line
		t.Errorf("Expected width: %d, got: %d\n", customWidth, gotWidth)
	}
}

func TestBarSetInvalidWidth(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	bar := p.AddBar(100).SetWidth(1).
		TrimLeftSpace().TrimRightSpace()
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()

	wantWidth := 80
	gotWidth := len(buf.Bytes())
	if gotWidth != wantWidth+1 { // +1 for new line
		t.Errorf("Expected width: %d, got: %d\n", wantWidth, gotWidth)
	}
}

func TestBarFormat(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	p := mpb.New().WithCancel(cancel).SetOut(&buf)
	customFormat := "(#>_)"
	bar := p.AddBar(100).Format(customFormat).
		TrimLeftSpace().TrimRightSpace()

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(10 * time.Millisecond)
			bar.Incr(1)
		}
	}()

	time.Sleep(250 * time.Millisecond)
	close(cancel)
	p.Stop()

	// removing new line
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

func TestBarInvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	customWidth := 60
	p := mpb.New().SetWidth(customWidth).SetOut(&buf)
	customFormat := "(#>=_)"
	bar := p.AddBar(100).Format(customFormat).
		TrimLeftSpace().TrimRightSpace()

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	bytes := removeLastRune(buf.Bytes())
	got := string(bytes[len(bytes)-customWidth:])
	want := "[==========================================================]"
	if got != want {
		t.Errorf("Expected format: %s, got %s\n", want, got)
	}
}

func TestBarInProgress(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	p := mpb.New().WithCancel(cancel).SetOut(&buf)
	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace()

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

func TestGetSpinner(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	bar := p.AddBar(0).TrimLeftSpace().TrimRightSpace()

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	spinnerChars := []byte(`-\|/`)
	seen := make(map[byte]bool)
	for _, b := range buf.Bytes() {
		if !seen[b] {
			seen[b] = true
		}
	}
	for _, b := range spinnerChars {
		if !seen[b] {
			t.Errorf("Char %#U not found in bar's output\n", b)
		}
	}
}

func TestBarGetID(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	numBars := 3
	wg.Add(numBars)

	bars := make([]*mpb.Bar, numBars)
	for i := 0; i < numBars; i++ {
		bars[i] = p.AddBarWithID(i, 100)

		go func(bar *mpb.Bar) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Incr(1)
			}
		}(bars[i])
	}

	for wantID, bar := range bars {
		gotID := bar.GetID()
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
	p := mpb.New().SetWidth(width).SetOut(&buf)

	total := 100
	refill := 30
	delta := total - refill
	refillChar := '+'
	bar := p.AddBar(int64(total)).TrimLeftSpace().TrimRightSpace()

	bar.IncrWithReFill(refill, &mpb.Refill{Char: refillChar})

	for i := 0; i < delta; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	bytes := removeLastRune(buf.Bytes())

	gotBar := string(bytes[len(bytes)-width:])
	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(string(refillChar), refill-1),
		strings.Repeat("=", delta-1))
	if gotBar != wantBar {
		t.Errorf("Want bar: %s, got bar: %s\n", wantBar, gotBar)
	}
}

func TestBarPanics(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	wantPanic := "Upps!!!"
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("b#%02d:", i)
		bar := p.AddBarWithID(i, 100).
			PrependFunc(func(s *mpb.Statistics, yw chan<- int, mw <-chan int) string {
				if s.Id == 2 && s.Current >= 42 {
					panic(wantPanic)
				}
				return name
			})

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				bar.Incr(1)
			}
		}()
	}

	wg.Wait()
	p.Stop()

	out := strings.Split(buf.String(), "\n")
	gotPanic := out[len(out)-2]
	if gotPanic != wantPanic {
		t.Errorf("Want panic: %s, got panic: %s\n", wantPanic, gotPanic)
	}
}

func removeLastRune(bytes []byte) []byte {
	_, size := utf8.DecodeLastRune(bytes)
	return bytes[:len(bytes)-size]
}
