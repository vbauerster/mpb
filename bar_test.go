package mpb_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func TestBarCompleted(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	if bar.Completed() {
		t.Fail()
	}

	bar.IncrBy(total)

	if !bar.Completed() {
		t.Error("bar isn't completed after increment")
	}

	p.Wait()
}

func TestBarAborted(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	if bar.Aborted() {
		t.Fail()
	}

	bar.Abort(false)

	if !bar.Aborted() {
		t.Error("bar isn't aborted after abort call")
	}

	p.Wait()
}

func TestBarEnableTriggerCompleteAndIncrementBefore(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0) // never complete bar

	for _, f := range []func(){
		func() { bar.SetTotal(40, false) },
		func() { bar.IncrBy(60) },
		func() { bar.SetTotal(80, false) },
		func() { bar.IncrBy(20) },
	} {
		f()
		if bar.Completed() {
			t.Fail()
		}
	}

	bar.EnableTriggerComplete()

	if !bar.Completed() {
		t.Fail()
	}

	p.Wait()
}

func TestBarEnableTriggerCompleteAndIncrementAfter(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0) // never complete bar

	for _, f := range []func(){
		func() { bar.SetTotal(40, false) },
		func() { bar.IncrBy(60) },
		func() { bar.SetTotal(80, false) },
		func() { bar.EnableTriggerComplete() },
	} {
		f()
		if bar.Completed() {
			t.Fail()
		}
	}

	bar.IncrBy(20)

	if !bar.Completed() {
		t.Fail()
	}

	p.Wait()
}

func TestBarID(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	total := 100
	wantID := 11
	bar := p.AddBar(int64(total), mpb.BarID(wantID))

	gotID := bar.ID()
	if gotID != wantID {
		t.Errorf("Expected bar id: %d, got %d\n", wantID, gotID)
	}

	bar.IncrBy(total)

	p.Wait()
}

func TestBarSetRefill(t *testing.T) {
	var buf bytes.Buffer

	p := mpb.New(mpb.WithOutput(&buf), mpb.WithWidth(100))

	total := 100
	till := 30
	refiller := "+"

	bar := p.New(int64(total), mpb.BarStyle().Refiller(refiller), mpb.BarFillerTrim())

	bar.SetRefill(int64(till))
	bar.IncrBy(till)
	bar.IncrBy(total - till)

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(refiller, till-1),
		strings.Repeat("=", total-till-1),
	)

	got := string(bytes.Split(buf.Bytes(), []byte("\n"))[0])

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q, got bar: %q\n", wantBar, got)
	}
}

func TestBarHas100PercentWithBarRemoveOnComplete(t *testing.T) {
	var buf bytes.Buffer

	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(&buf))

	total := 50

	bar := p.AddBar(int64(total),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(decor.Percentage()),
	)

	bar.IncrBy(total)

	p.Wait()

	hundred := "100 %"
	if !bytes.Contains(buf.Bytes(), []byte(hundred)) {
		t.Errorf("Bar's buffer does not contain: %q\n", hundred)
	}
}

func TestBarStyle(t *testing.T) {
	var buf bytes.Buffer
	customFormat := "╢▌▌░╟"
	runes := []rune(customFormat)
	total := 80
	p := mpb.New(mpb.WithWidth(total), mpb.WithOutput(&buf))
	bs := mpb.BarStyle()
	bs.Lbound(string(runes[0]))
	bs.Filler(string(runes[1]))
	bs.Tip(string(runes[2]))
	bs.Padding(string(runes[3]))
	bs.Rbound(string(runes[4]))
	bar := p.New(int64(total), bs, mpb.BarFillerTrim())

	bar.IncrBy(total)

	p.Wait()

	wantBar := fmt.Sprintf("%s%s%s%s",
		string(runes[0]),
		strings.Repeat(string(runes[1]), total-3),
		string(runes[2]),
		string(runes[4]),
	)
	got := string(bytes.Split(buf.Bytes(), []byte("\n"))[0])

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q:%d, got bar: %q:%d\n", wantBar, utf8.RuneCountInString(wantBar), got, utf8.RuneCountInString(got))
	}
}

func TestBarPanicBeforeComplete(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithDebugOutput(&buf),
		mpb.WithOutput(io.Discard),
	)

	total := 100
	panicMsg := "Upps!!!"
	var pCount uint32
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(panicDecorator(panicMsg,
			func(st decor.Statistics) bool {
				if st.Current >= 42 {
					atomic.AddUint32(&pCount, 1)
					return true
				}
				return false
			},
		)),
	)

	bar.IncrBy(total)

	p.Wait()

	if pCount != 1 {
		t.Errorf("Decorator called after panic %d times, expected 1\n", pCount)
	}

	barStr := buf.String()
	if !strings.Contains(barStr, panicMsg) {
		t.Errorf("%q doesn't contain %q\n", barStr, panicMsg)
	}
}

func TestBarPanicAfterComplete(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithDebugOutput(&buf),
		mpb.WithOutput(io.Discard),
	)

	total := 100
	panicMsg := "Upps!!!"
	var pCount uint32
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(panicDecorator(panicMsg,
			func(st decor.Statistics) bool {
				if st.Completed {
					atomic.AddUint32(&pCount, 1)
					return true
				}
				return false
			},
		)),
	)

	bar.IncrBy(total)

	p.Wait()

	if pCount != 1 {
		t.Errorf("Decorator called after panic %d times, expected 1\n", pCount)
	}

	barStr := buf.String()
	if !strings.Contains(barStr, panicMsg) {
		t.Errorf("%q doesn't contain %q\n", barStr, panicMsg)
	}
}

func TestDecorStatisticsAvailableWidth(t *testing.T) {
	var called [2]bool
	td1 := func(s decor.Statistics) string {
		if s.AvailableWidth != 80 {
			t.Errorf("expected AvailableWidth %d got %d\n", 80, s.AvailableWidth)
		}
		called[0] = true
		return fmt.Sprintf("\x1b[31;1;4m%s\x1b[0m", strings.Repeat("0", 20))
	}
	td2 := func(s decor.Statistics) string {
		if s.AvailableWidth != 40 {
			t.Errorf("expected AvailableWidth %d got %d\n", 40, s.AvailableWidth)
		}
		called[1] = true
		return ""
	}
	ctx, cancel := context.WithCancel(context.Background())
	refresh := make(chan interface{})
	p := mpb.NewWithContext(ctx, mpb.WithWidth(100),
		mpb.WithManualRefresh(refresh),
		mpb.WithOutput(io.Discard),
	)
	_ = p.AddBar(0,
		mpb.BarFillerTrim(),
		mpb.PrependDecorators(
			decor.Name(strings.Repeat("0", 20)),
			decor.Any(td1),
		),
		mpb.AppendDecorators(
			decor.Name(strings.Repeat("0", 20)),
			decor.Any(td2),
		),
	)
	refresh <- time.Now()
	cancel()
	p.Wait()
	for i, ok := range called {
		if !ok {
			t.Errorf("Decorator %d isn't called", i+1)
		}
	}
}

func panicDecorator(panicMsg string, cond func(decor.Statistics) bool) decor.Decorator {
	return decor.Any(func(st decor.Statistics) string {
		if cond(st) {
			panic(panicMsg)
		}
		return ""
	})
}
