package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestAddBar(t *testing.T) {
	p := mpb.New()

	count := p.BarCount()
	if count != 0 {
		t.Errorf("BarCount want: %q, got: %q\n", 0, count)
	}

	bar := p.AddBar(100)

	count = p.BarCount()
	if count != 1 {
		t.Errorf("BarCount want: %q, got: %q\n", 1, count)
	}

	bar.Complete()
	p.Stop()
}

func TestRemoveBar(t *testing.T) {
	p := mpb.New()

	bar := p.AddBar(10)

	if !p.RemoveBar(bar) {
		t.Error("RemoveBar failure")
	}

	count := p.BarCount()
	if count != 0 {
		t.Errorf("BarCount want: %q, got: %q\n", 0, count)
	}

	bar.Complete()
	p.Stop()
}

func TestRemoveBars(t *testing.T) {
	p := mpb.New(mpb.Output(ioutil.Discard))

	var wg sync.WaitGroup
	bars := make([]*mpb.Bar, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		b := p.AddBar(80)
		bars[i] = b
		go func() {
			for i := 0; i < 80; i++ {
				if i == 33 {
					wg.Done()
				}
				b.Increment()
				time.Sleep(randomDuration(80 * time.Millisecond))
			}
		}()
	}

	wg.Wait()
	for i := 0; i < 3; i++ {
		i := i
		go func() {
			if ok := p.RemoveBar(bars[i]); !ok {
				t.Errorf("bar %d: remove failed\n", i)
			}
		}()
	}
	p.Stop()
}

func TestWithCancel(t *testing.T) {
	var wg sync.WaitGroup
	cancel := make(chan struct{})
	shutdown := make(chan struct{})
	p := mpb.New(
		mpb.Output(ioutil.Discard),
		mpb.WithCancel(cancel),
		mpb.WithShutdownNotifier(shutdown),
		mpb.WithWaitGroup(&wg),
	)

	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		bar := p.AddBar(int64(total), mpb.BarID(i),
			mpb.PrependDecorators(decor.StaticName(name, len(name), 0)))

		go func() {
			defer wg.Done()
			for i := 0; i < total; i++ {
				select {
				case <-cancel:
					return
				default:
				}
				time.Sleep(randomDuration(80 * time.Millisecond))
				bar.Increment()
			}
		}()
	}

	time.AfterFunc(300*time.Millisecond, func() {
		close(cancel)
	})

	p.Stop()

	select {
	case <-shutdown:
	case <-time.After(300 * time.Millisecond):
		t.Error("ProgressBar didn't stop")
	}
}

func TestCustomFormat(t *testing.T) {
	var buf bytes.Buffer
	cancel := make(chan struct{})
	customFormat := "╢▌▌░╟"
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithCancel(cancel),
		mpb.WithFormat(customFormat),
	)
	bar := p.AddBar(80, mpb.BarTrim())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < 80; i++ {
			if i == 33 {
				wg.Done()
			}
			time.Sleep(randomDuration(80 * time.Millisecond))
			bar.Increment()
		}
	}()

	wg.Wait()
	close(cancel)
	p.Stop()

	for _, r := range customFormat {
		if !bytes.ContainsRune(buf.Bytes(), r) {
			t.Errorf("Rune %#U not found in bar\n", r)
		}
	}
}

func TestInvalidFormatWidth(t *testing.T) {
	var buf bytes.Buffer
	customWidth := 60
	customFormat := "(#>=_)"
	p := mpb.New(
		mpb.Output(&buf),
		mpb.WithWidth(customWidth),
		mpb.WithFormat(customFormat),
	)
	bar := p.AddBar(100, mpb.BarTrim())

	for i := 0; i < 100; i++ {
		time.Sleep(randomDuration(40 * time.Millisecond))
		bar.Increment()
	}

	p.Stop()

	got := buf.String()
	want := fmt.Sprintf("[%s]", strings.Repeat("=", customWidth-2))
	if !strings.Contains(got, want) {
		t.Errorf("Expected format: %s, got %s\n", want, got)
	}
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
