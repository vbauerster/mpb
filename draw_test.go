package mpb

import (
	"bytes"
	"testing"
)

func TestFillBar(t *testing.T) {
	tests := []struct {
		termWidth int
		barWidth  int
		total     int64
		current   int64
		barRefill *refill
		want      string
	}{
		{
			termWidth: 2,
			barWidth:  100,
			want:      "[]",
		},
		{
			termWidth: 3,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      "[-]",
		},
		{
			termWidth: 5,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      "[>--]",
		},
		{
			termWidth: 6,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      "[>---]",
		},
		{
			termWidth: 20,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      "[===>--------------]",
		},
		{
			termWidth: 50,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      "[=========>--------------------------------------]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   0,
			want:      "[--------------------------------------------------------------------------------------------------]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   1,
			want:      "[>-------------------------------------------------------------------------------------------------]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   40,
			want:      "[======================================>-----------------------------------------------------------]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   40,
			barRefill: &refill{'+', 32},
			want:      "[+++++++++++++++++++++++++++++++=======>-----------------------------------------------------------]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   99,
			want:      "[================================================================================================>-]",
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   100,
			want:      "[==================================================================================================]",
		},
	}

	prependWs := newWidthSync(nil, 1, 0)
	appendWs := newWidthSync(nil, 1, 0)
	for _, test := range tests {
		s := newTestState()
		s.width = test.barWidth
		s.total = test.total
		s.current = test.current
		if test.barRefill != nil {
			s.refill = test.barRefill
		}
		s.draw(test.termWidth, prependWs, appendWs)
		got := s.bufB.String()
		if got != test.want {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

func newTestState() *state {
	s := &state{
		trimLeftSpace:  true,
		trimRightSpace: true,
		bufP:           new(bytes.Buffer),
		bufB:           new(bytes.Buffer),
		bufA:           new(bytes.Buffer),
	}
	s.updateFormat("[=>-]")
	return s
}
