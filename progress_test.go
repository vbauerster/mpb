package mpb

import "testing"

func TestAddBar(t *testing.T) {
	p := New().SetWidth(60)
	count := p.BarsCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	p.Wg.Add(1)
	bar := p.AddBar(10)
	count = p.BarsCount()
	if count != 1 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	for i := 0; i < 10; i++ {
		bar.Incr(1)
	}
	p.WaitAndStop()
}

func TestRemoveBar(t *testing.T) {
	p := New()
	p.Wg.Add(1)
	b := p.AddBar(10)

	if !p.RemoveBar(b) {
		t.Error("RemoveBar failure")
	}

	count := p.BarsCount()
	if count != 0 {
		t.Errorf("Count want: %q, got: %q\n", 0, count)
	}
	p.WaitAndStop()
}
