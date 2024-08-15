package mpb_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func TestBarCompleted(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	if bar.Completed() {
		t.Error("expected bar not to complete")
	}

	bar.IncrBy(total)

	if !bar.Completed() {
		t.Error("expected bar to complete")
	}

	p.Wait()
}

func TestBarAborted(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	if bar.Aborted() {
		t.Error("expected bar not to be aborted")
	}

	bar.Abort(false)

	if !bar.Aborted() {
		t.Error("expected bar to be aborted")
	}

	p.Wait()
}

func TestBarSetTotal(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0)

	bar.SetTotal(0, false)
	if bar.Completed() {
		t.Error("expected bar not to complete")
	}

	bar.SetTotal(0, true)
	if !bar.Completed() {
		t.Error("expected bar to complete")
	}

	p.Wait()
}

func TestBarEnableTriggerCompleteZeroBar(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0) // never complete bar

	if bar.Completed() {
		t.Error("expected bar not to complete")
	}

	// Calling bar.SetTotal(0, true) has same effect
	// but this one is more concise and intuitive
	bar.EnableTriggerComplete()

	if !bar.Completed() {
		t.Error("expected bar to complete")
	}

	p.Wait()
}

func TestBarEnableTriggerCompleteAndIncrementBefore(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0) // never complete bar

	targetTotal := int64(80)

	for _, f := range []func(){
		func() { bar.SetTotal(40, false) },
		func() { bar.IncrBy(60) },
		func() { bar.SetTotal(targetTotal, false) },
		func() { bar.IncrBy(20) },
	} {
		f()
		if bar.Completed() {
			t.Error("expected bar not to complete")
		}
	}

	bar.EnableTriggerComplete()

	if !bar.Completed() {
		t.Error("expected bar to complete")
	}

	if current := bar.Current(); current != targetTotal {
		t.Errorf("Expected current: %d, got: %d", targetTotal, current)
	}

	p.Wait()
}

func TestBarEnableTriggerCompleteAndIncrementAfter(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(io.Discard))
	bar := p.AddBar(0) // never complete bar

	targetTotal := int64(80)

	for _, f := range []func(){
		func() { bar.SetTotal(40, false) },
		func() { bar.IncrBy(60) },
		func() { bar.SetTotal(targetTotal, false) },
		func() { bar.EnableTriggerComplete() }, // disables any next SetTotal
		func() { bar.SetTotal(100, true) },     // nop
	} {
		f()
		if bar.Completed() {
			t.Error("expected bar not to complete")
		}
	}

	bar.IncrBy(20)

	if !bar.Completed() {
		t.Error("expected bar to complete")
	}

	if current := bar.Current(); current != targetTotal {
		t.Errorf("Expected current: %d, got: %d", targetTotal, current)
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
		t.Errorf("Expected bar id: %d, got %d", wantID, gotID)
	}

	bar.IncrBy(total)

	p.Wait()
}

func TestBarSetRefill(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(
		mpb.WithWidth(100),
		mpb.WithOutput(&buf),
		mpb.WithAutoRefresh(),
	)

	total := 100
	till := 30
	refiller := "+"

	bar := p.New(int64(total), mpb.BarStyle().Refiller(refiller), mpb.BarFillerTrim())

	bar.IncrBy(till)
	bar.SetRefill(int64(till))
	bar.IncrBy(total - till)

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(refiller, till-1),
		strings.Repeat("=", total-till-1),
	)

	got := string(bytes.Split(buf.Bytes(), []byte("\n"))[0])

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q, got bar: %q", wantBar, got)
	}
}

func TestBarHas100PercentWithBarRemoveOnComplete(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithOutput(&buf),
		mpb.WithAutoRefresh(),
	)

	total := 50

	bar := p.AddBar(int64(total),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(decor.Percentage()),
	)

	bar.IncrBy(total)

	p.Wait()

	hundred := "100 %"
	if !bytes.Contains(buf.Bytes(), []byte(hundred)) {
		t.Errorf("Bar's buffer does not contain: %q", hundred)
	}
}

func TestBarStyle(t *testing.T) {
	var buf bytes.Buffer
	customFormat := "╢▌▌░╟"
	runes := []rune(customFormat)
	total := 80
	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithOutput(&buf),
		mpb.WithAutoRefresh(),
	)
	bs := mpb.BarStyle()
	bs = bs.Lbound(string(runes[0]))
	bs = bs.Filler(string(runes[1]))
	bs = bs.Tip(string(runes[2]))
	bs = bs.Padding(string(runes[3]))
	bs = bs.Rbound(string(runes[4]))
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
		t.Errorf("Want bar: %q:%d, got bar: %q:%d", wantBar, utf8.RuneCountInString(wantBar), got, utf8.RuneCountInString(got))
	}
}

func TestDecorStatisticsAvailableWidth(t *testing.T) {
	ch := make(chan int, 2)
	td1 := func(s decor.Statistics) string {
		ch <- s.AvailableWidth
		return strings.Repeat("0", 20)
	}
	td2 := func(s decor.Statistics) string {
		ch <- s.AvailableWidth
		return ""
	}
	ctx, cancel := context.WithCancel(context.Background())
	refresh := make(chan interface{})
	p := mpb.NewWithContext(ctx,
		mpb.WithWidth(100),
		mpb.WithManualRefresh(refresh),
		mpb.WithOutput(io.Discard),
	)
	_ = p.AddBar(0,
		mpb.BarFillerTrim(),
		mpb.PrependDecorators(
			decor.Name(strings.Repeat("0", 20)),
			decor.Meta(
				decor.Any(td1),
				func(s string) string {
					return "\x1b[31;1m" + s + "\x1b[0m"
				},
			),
		),
		mpb.AppendDecorators(
			decor.Name(strings.Repeat("0", 20)),
			decor.Any(td2),
		),
	)
	refresh <- time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	p.Wait()

	if availableWidth := <-ch; availableWidth != 80 {
		t.Errorf("expected AvailableWidth %d got %d", 80, availableWidth)
	}

	if availableWidth := <-ch; availableWidth != 40 {
		t.Errorf("expected AvailableWidth %d got %d", 40, availableWidth)
	}
}

func TestBarQueueAfterBar(t *testing.T) {
	shutdown := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(io.Discard),
		mpb.WithAutoRefresh(),
		mpb.WithShutdownNotifier(shutdown),
	)
	a := p.AddBar(100)
	b := p.AddBar(100, mpb.BarQueueAfter(a))
	identity := map[*mpb.Bar]string{
		a: "a",
		b: "b",
	}

	a.IncrBy(100)
	a.Wait()
	cancel()

	bars := (<-shutdown).([]*mpb.Bar)
	if l := len(bars); l != 1 {
		t.Errorf("Expected len of bars: %d, got: %d", 1, l)
	}

	p.Wait()
	if bars[0] != b {
		t.Errorf("Expected bars[0] == b, got: %s", identity[bars[0]])
	}
}
