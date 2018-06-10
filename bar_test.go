package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
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
	total := 80
	wantID := 11
	bar := p.AddBar(int64(total), BarID(wantID))

	go func() {
		for i := 0; i < total; i++ {
			time.Sleep(50 * time.Millisecond)
			bar.Increment()
		}
	}()

	gotID := bar.ID()
	if gotID != wantID {
		t.Errorf("Expected bar id: %d, got %d\n", wantID, gotID)
	}

	p.Abort(bar)
	p.Wait()
}

func TestBarIncrWithReFill(t *testing.T) {
	var buf bytes.Buffer

	width := 100
	p := New(WithOutput(&buf), WithWidth(width))

	total := 100
	till := 30
	refillChar := '+'

	bar := p.AddBar(int64(total), BarTrim())

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
	var buf bytes.Buffer
	p := New(WithDebugOutput(&buf), WithOutput(ioutil.Discard))

	wantPanic := "Upps!!!"
	total := 100

	bar := p.AddBar(int64(total), PrependDecorators(
		func(s *decor.Statistics, _ chan<- int, _ <-chan int) string {
			if s.Current >= 42 {
				panic(wantPanic)
			}
			return "test"
		},
	))

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(10 * time.Millisecond)
			bar.Increment()
		}
	}()

	p.Wait()

	wantPanic = fmt.Sprintf("panic: %s", wantPanic)
	debugStr := buf.String()
	if !strings.Contains(debugStr, wantPanic) {
		t.Errorf("%q doesn't contain %q\n", debugStr, wantPanic)
	}
}
