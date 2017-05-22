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

	bytes := buf.Bytes()
	_, size := utf8.DecodeLastRune(bytes)
	bytes = bytes[:len(bytes)-size] // removing new line

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
