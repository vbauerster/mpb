package mpb

import (
	"reflect"
	"testing"
)

func TestFillBar(t *testing.T) {
	tests := []struct {
		termWidth int
		barWidth  int
		total     int64
		current   int64
		barRefill *refill
		want      []byte
	}{
		{
			termWidth: 1,
			barWidth:  100,
			want:      []byte{},
		},
		{
			termWidth: 2,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      []byte("[]"),
		},
		{
			termWidth: 20,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      []byte("[===>--------------]"),
		},
		{
			termWidth: 50,
			barWidth:  100,
			total:     100,
			current:   20,
			want:      []byte("[=========>--------------------------------------]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   0,
			want:      []byte("[--------------------------------------------------------------------------------------------------]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   1,
			want:      []byte("[>-------------------------------------------------------------------------------------------------]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   40,
			want:      []byte("[======================================>-----------------------------------------------------------]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   40,
			barRefill: &refill{'+', 32},
			want:      []byte("[+++++++++++++++++++++++++++++++=======>-----------------------------------------------------------]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   99,
			want:      []byte("[================================================================================================>-]"),
		},
		{
			termWidth: 100,
			barWidth:  100,
			total:     100,
			current:   100,
			want:      []byte("[==================================================================================================]"),
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
		got := draw(s, test.termWidth, prependWs, appendWs)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

func newTestState() *state {
	return &state{
		format:         barFmtRunes{'[', '=', '>', '-', ']'},
		trimLeftSpace:  true,
		trimRightSpace: true,
	}
}
