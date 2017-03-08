package mpb

import (
	"bytes"
	"testing"
)

func TestAddBar(t *testing.T) {
	var buf bytes.Buffer
	p := New().SetWidth(60).SetOut(&buf)
	count := p.BarCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	bar := p.AddBar(10)
	count = p.BarCount()
	if count != 1 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	for i := 0; i < 10; i++ {
		bar.Incr(1)
	}
	p.Stop()
}

func TestRemoveBar(t *testing.T) {
	p := New()
	b := p.AddBar(10)

	if !p.RemoveBar(b) {
		t.Error("RemoveBar failure")
	}

	count := p.BarCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	p.Stop()
}
