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
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestAddBar(t *testing.T) {
	p := mpb.New(mpb.Output(ioutil.Discard))

	var wg sync.WaitGroup
	wg.Add(1)
	b := p.AddBar(80)
	go func() {
		for i := 0; i < 80; i++ {
			if i == 33 {
				wg.Done()
			}
			b.Increment()
			time.Sleep(randomDuration(80 * time.Millisecond))
		}
	}()

	wg.Wait()
	count := p.BarCount()
	if count != 1 {
		t.Errorf("BarCount want: %q, got: %q\n", 1, count)
	}

	b.Complete()
	p.Wait()
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
	p.Wait()
}

func TestWithCancel(t *testing.T) {
	cancel := make(chan struct{})
	shutdown := make(chan struct{})
	p := mpb.New(
		mpb.Output(ioutil.Discard),
		mpb.WithCancel(cancel),
		mpb.WithShutdownNotifier(shutdown),
	)

	numBars := 3
	bars := make([]*mpb.Bar, 0, numBars)
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(1000), mpb.BarID(i))
		bars = append(bars, bar)
		go func() {
			for !bar.Completed() {
				time.Sleep(randomDuration(40 * time.Millisecond))
				bar.Increment()
			}
		}()
	}

	time.AfterFunc(100*time.Millisecond, func() {
		close(cancel)
	})

	p.Wait()
	for _, bar := range bars {
		if bar.Current() >= bar.Total() {
			t.Errorf("bar %d: total = %d, current = %d\n", bar.ID(), bar.Total(), bar.Current())
		}
	}
	select {
	case <-shutdown:
	case <-time.After(100 * time.Millisecond):
		t.Error("Progress didn't stop")
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
	p.Wait()

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

	p.Wait()

	got := buf.String()
	want := fmt.Sprintf("[%s]", strings.Repeat("=", customWidth-2))
	if !strings.Contains(got, want) {
		t.Errorf("Expected format: %s, got %s\n", want, got)
	}
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
