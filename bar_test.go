package mpb_test

import (
	"bytes"
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

func removeLastRune(bytes []byte) []byte {
	_, size := utf8.DecodeLastRune(bytes)
	return bytes[:len(bytes)-size]
}
