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

	var wg sync.WaitGroup
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBarWithID(i, int64(total)).PrependName(name, len(name), 0)

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
	case <-time.After(500 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}

func TestWithNilCancel(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			if msg, ok := p.(string); ok && msg == "nil cancel channel" {
				return
			}
			t.Errorf("Expected nil channel panic, got: %+v", p)
		}
	}()
	_ = mpb.New().WithCancel(nil)
}

func TestFormat(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	shutdown := make(chan struct{})
	customFormat := "╢▌▌░╟"
	p := mpb.New().Format(customFormat)
	p.WithCancel(cancel)
	p.ShutdownNotify(shutdown)
	p.SetOut(&buf)
	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace()

	go func() {
		for i := 0; i < 100; i++ {
			bar.Incr(1)
			time.Sleep(10 * time.Millisecond)
			if i == 42 {
				close(cancel)
			}
		}
	}()

	// p.Stop()
	// time.Sleep(300 * time.Millisecond)

	// gotBar := strings.TrimSpace(buf.String())
	gotBar := buf.String()
	seen := make(map[rune]bool)
	for _, r := range gotBar {
		if !seen[r] {
			seen[r] = true
		}
	}
	fmt.Println(gotBar)
	for r, _ := range seen {
		fmt.Printf("%#U\n", r)
	}
	// expectBar := "╢▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌╟"
	// if gotBar != expectBar {
	// 	t.Errorf("Expected for")
	// }
}
