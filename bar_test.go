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

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
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
	refillRune, _ := utf8.DecodeLastRuneInString(mpb.BarDefaultStyle)

	bar := p.AddBar(int64(total), mpb.BarFillerTrim())

	bar.SetRefill(int64(till))
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
	total := 80
	p := mpb.New(mpb.WithWidth(total), mpb.WithOutput(&buf))
	bar := p.Add(int64(total), mpb.NewBarFiller(customFormat), mpb.BarFillerTrim())

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

func panicDecorator(panicMsg string, cond func(decor.Statistics) bool) decor.Decorator {
	return decor.Any(func(st decor.Statistics) string {
		if cond(st) {
			panic(panicMsg)
		}
		return ""
	})
}
