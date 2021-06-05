package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func TestBarCompleted(t *testing.T) {
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(ioutil.Discard))
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
	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(ioutil.Discard))
	total := 100
	wantID := 11
	bar := p.AddBar(int64(total), mpb.BarID(wantID))

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

	bar.Abort(true)
	p.Wait()
}

func TestBarSetRefill(t *testing.T) {
	var buf bytes.Buffer

	p := mpb.New(mpb.WithOutput(&buf), mpb.WithWidth(100))

	total := 100
	till := 30
	refiller := "+"

	bar := p.Add(int64(total), mpb.NewBarFiller(mpb.BarStyle().Refiller(refiller)), mpb.BarFillerTrim())

	bar.SetRefill(int64(till))
	bar.IncrBy(till)

	for i := 0; i < total-till; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(refiller, till-1),
		strings.Repeat("=", total-till-1),
	)

	got := string(getLastLine(buf.Bytes()))

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q, got bar: %q\n", wantBar, got)
	}
}

func TestBarHas100PercentWithOnCompleteDecorator(t *testing.T) {
	var buf bytes.Buffer

	p := mpb.New(mpb.WithWidth(80), mpb.WithOutput(&buf))

	total := 50

	bar := p.AddBar(int64(total),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.Percentage(), "done",
			),
		),
	)

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	hundred := "100 %"
	if !bytes.Contains(buf.Bytes(), []byte(hundred)) {
		t.Errorf("Bar's buffer does not contain: %q\n", hundred)
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

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

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
	bar := p.Add(int64(total), mpb.NewBarFiller(bs), mpb.BarFillerTrim())

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	wantBar := fmt.Sprintf("%s%s%s%s",
		string(runes[0]),
		strings.Repeat(string(runes[1]), total-3),
		string(runes[2]),
		string(runes[4]),
	)
	got := string(getLastLine(buf.Bytes()))

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q:%d, got bar: %q:%d\n", wantBar, utf8.RuneCountInString(wantBar), got, utf8.RuneCountInString(got))
	}
}

func TestBarPanicBeforeComplete(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithDebugOutput(&buf),
		mpb.WithOutput(ioutil.Discard),
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

	for i := 0; i < total; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
	}

	p.Wait()

	if pCount != 1 {
		t.Errorf("Decor called after panic %d times\n", pCount-1)
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
		mpb.WithOutput(ioutil.Discard),
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

	for i := 0; i < total; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
	}

	p.Wait()

	if pCount > 2 {
		t.Error("Decor called after panic more than 2 times\n")
	}

	barStr := buf.String()
	if !strings.Contains(barStr, panicMsg) {
		t.Errorf("%q doesn't contain %q\n", barStr, panicMsg)
	}
}

func TestDecorStatisticsAvailableWidth(t *testing.T) {
	td1 := func(s decor.Statistics) string {
		if s.AvailableWidth != 80 {
			t.Errorf("expected AvailableWidth %d got %d\n", 80, s.AvailableWidth)
		}
		return fmt.Sprintf("\x1b[31;1;4m%s\x1b[0m", strings.Repeat("1", 20))
	}
	td2 := func(s decor.Statistics) string {
		if s.AvailableWidth != 60 {
			t.Errorf("expected AvailableWidth %d got %d\n", 60, s.AvailableWidth)
		}
		return ""
	}
	total := 100
	p := mpb.New(
		mpb.WithWidth(100),
		mpb.WithOutput(ioutil.Discard),
	)
	bar := p.AddBar(int64(total),
		mpb.BarFillerTrim(),
		mpb.PrependDecorators(
			decor.Name(strings.Repeat("0", 20)),
			decor.Any(td1),
		),
		mpb.AppendDecorators(
			decor.Any(td2),
		),
	)
	for i := 0; i < total; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
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
