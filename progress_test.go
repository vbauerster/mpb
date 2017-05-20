package mpb_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb"
)

func TestDefaultWidth(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	bar := p.AddBar(100)
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()
	runeCount := utf8.RuneCountInString(strings.TrimSpace(buf.String()))
	defWidth := 80
	if runeCount != defWidth {
		defWidth = 78 // when testing with ./...
		if runeCount != defWidth {
			t.Errorf("Expected default width: %d, got: %d\n", defWidth, runeCount)
		}
	}
}

func TestCustomWidth(t *testing.T) {
	customWidth := 60
	var buf bytes.Buffer
	p := mpb.New().SetWidth(customWidth).SetOut(&buf)
	bar := p.AddBar(100)
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()
	runeCount := utf8.RuneCountInString(strings.TrimSpace(buf.String()))
	if runeCount != customWidth {
		customWidth = 58 // when testing with ./...
		if runeCount != customWidth {
			t.Errorf("Expected default width: %d, got: %d\n", customWidth, runeCount)
		}
	}
}

func TestAddBar(t *testing.T) {
	var wg sync.WaitGroup
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	count := p.BarCount()
	if count != 0 {
		t.Errorf("BarCount want: %q, got: %q\n", 0, count)
	}

	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name), 0)

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
	p := mpb.New().WithCancel(cancel).ShutdownNotify(shutdown)
	numBars := 3

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(100).PrependName(name, len(name), 0)

		go func() {
			for i := 0; i < 10000; i++ {
				bar.Incr(1)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}()
	}

	close(cancel)

	select {
	case <-shutdown:
	case <-time.After(500 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}
