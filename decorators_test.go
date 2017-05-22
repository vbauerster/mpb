package mpb_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
)

func TestPrependName(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	wantName := "TestBar"
	bar := p.AddBar(100).PrependName(wantName, 0, 0)
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}
	p.Stop()
	if !strings.Contains(buf.String(), wantName) {
		t.Errorf("%q not found in bar\n", wantName)
	}
}

func TestPrependCounters(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	reader := strings.NewReader(content)

	total := int64(len(content))
	bar := p.AddBar(total).TrimLeftSpace().TrimRightSpace().
		PrependCounters("%3s / %3s", mpb.UnitBytes, 0, 0)
	preader := bar.ProxyReader(reader)

	_, err := io.Copy(ioutil.Discard, preader)
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Stop()

	out := buf.String()
	out = out[:strings.IndexRune(out, '[')]
	want := fmt.Sprintf("%[1]db / %[1]db", total)
	if out != want {
		t.Errorf("Expected: %s, got %s\n", want, out)
	}
}

func TestAppendPercentage(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendPercentage(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "100 %"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("%q not found in bar\n", want)
	}
}

func TestPrependPercentage(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependPercentage(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "100 %"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("%q not found in bar\n", want)
	}
}
