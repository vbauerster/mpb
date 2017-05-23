package mpb_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/vbauerster/mpb"
)

func TestPrependName(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	name := "TestBar"
	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependName(name, 0, 0)
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}

	p.Stop()

	want := name + "["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependNameDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)
	name := "TestBar"
	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependName(name, len(name)+1, mpb.DidentRight)
	for i := 0; i < 100; i++ {
		bar.Incr(1)
	}

	p.Stop()

	want := name + " ["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
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

	barOut := buf.String()
	want := fmt.Sprintf("%[1]db / %[1]db[", total)
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependCountersDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	reader := strings.NewReader(content)

	total := int64(len(content))
	bar := p.AddBar(total).TrimLeftSpace().TrimRightSpace().
		PrependCounters("%3s / %3s", mpb.UnitBytes, 12, mpb.DidentRight)
	preader := bar.ProxyReader(reader)

	_, err := io.Copy(ioutil.Discard, preader)
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Stop()

	barOut := buf.String()
	want := fmt.Sprintf("%[1]db / %[1]db [", total)
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendPercentage(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendPercentage(6, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "] 100 %"
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendPercentageDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendPercentage(6, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "]100 % "
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependPercentage(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependPercentage(6, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := " 100 %["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependPercentageDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependPercentage(6, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "100 % ["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependElapsed(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependElapsed(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "1s["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependElapsedDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependElapsed(3, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "1s ["
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendElapsed(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendElapsed(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "]1s"
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendElapsedDindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendElapsed(3, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := "]1s "
	barOut := buf.String()
	if !strings.Contains(barOut, want) {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependETA(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependETA(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := `0s?\[`
	barOut := buf.String()

	matched, err := regexp.MatchString(want, barOut)
	if err != nil {
		t.Logf("Regex %q err: %+v\n", want, err)
		t.FailNow()
	}

	if !matched {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestPrependETADindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		PrependETA(3, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := `0s?\s+\[`
	barOut := buf.String()

	matched, err := regexp.MatchString(want, barOut)
	if err != nil {
		t.Logf("Regex %q err: %+v\n", want, err)
		t.FailNow()
	}

	if !matched {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendETA(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendETA(0, 0)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := `\]0s?`
	barOut := buf.String()

	matched, err := regexp.MatchString(want, barOut)
	if err != nil {
		t.Logf("Regex %q err: %+v\n", want, err)
		t.FailNow()
	}

	if !matched {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}

func TestAppendETADindentRight(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New().SetOut(&buf)

	bar := p.AddBar(100).TrimLeftSpace().TrimRightSpace().
		AppendETA(3, mpb.DidentRight)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Incr(1)
	}

	p.Stop()

	want := `\]0s? `
	barOut := buf.String()

	matched, err := regexp.MatchString(want, barOut)
	if err != nil {
		t.Logf("Regex %q err: %+v\n", want, err)
		t.FailNow()
	}

	if !matched {
		t.Errorf("%q not found in bar: %s\n", want, barOut)
	}
}
