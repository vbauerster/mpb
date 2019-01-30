package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	. "github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
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

	p.Abort(bar, true)
	p.Wait()
}

func TestBarSetRefill(t *testing.T) {
	var buf bytes.Buffer

	width := 100
	p := New(WithOutput(&buf), WithWidth(width))

	total := 100
	till := 30
	refillRune := DefaultBarStyle[len(DefaultBarStyle)-1]

	bar := p.AddBar(int64(total), TrimSpace())

	bar.SetRefill(till)
	bar.IncrBy(till)

	for i := 0; i < total-till; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(string(refillRune), till-1),
		strings.Repeat("=", total-till-1),
	)

	got := string(getLastLine(buf.Bytes()))

	if got != wantBar {
		t.Errorf("Want bar: %q, got bar: %q\n", wantBar, got)
	}
}

func TestBarStyle(t *testing.T) {
	var buf bytes.Buffer
	customFormat := "╢▌▌░╟"
	p := New(WithOutput(&buf))
	total := 80
	bar := p.AddBar(int64(total), BarStyle(customFormat), TrimSpace())

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	runes := []rune(customFormat)
	wantBar := fmt.Sprintf("%s%s%s",
		string(runes[0]),
		strings.Repeat(string(runes[1]), total-2),
		string(runes[len(runes)-1]),
	)
	got := string(getLastLine(buf.Bytes()))

	if got != wantBar {
		t.Errorf("Want bar: %q:%d, got bar: %q:%d\n", wantBar, utf8.RuneCountInString(wantBar), got, utf8.RuneCountInString(got))
	}
}

func TestBarPanics(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithDebugOutput(&buf), WithOutput(ioutil.Discard))

	wantPanic := "Upps!!!"
	total := 100

	bar := p.AddBar(int64(total), PrependDecorators(panicDecorator(wantPanic)))

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

func panicDecorator(panicMsg string) decor.Decorator {
	d := &decorator{
		panicMsg: panicMsg,
	}
	d.Init()
	return d
}

type decorator struct {
	decor.WC
	panicMsg string
}

func (d *decorator) Decor(st *decor.Statistics) string {
	if st.Current >= 42 {
		panic(d.panicMsg)
	}
	return d.FormatMsg("")
}
