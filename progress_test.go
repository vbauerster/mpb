package mpb

import "testing"

func TestAddBar(t *testing.T) {
	p := New()
	count := p.BarsCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	p.AddBar(10)
	count = p.BarsCount()
	if count != 1 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
}

func TestRemoveBar(t *testing.T) {
	p := New()
	b := p.AddBar(10)

	if !p.RemoveBar(b) {
		t.Error("RemoveBar failure")
	}

	count := p.BarsCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
}
