package mpb

import (
	"reflect"
	"testing"
)

func TestFillBar(t *testing.T) {
	s := newTestState(80, 60)
	tests := []struct {
		termWidth int
		barWidth  int
		want      []byte
	}{
		{
			termWidth: 1,
			barWidth:  60,
			want:      []byte{},
		},
		{
			termWidth: 2,
			barWidth:  60,
			want:      []byte("[]"),
		},
		{
			termWidth: 4,
			barWidth:  60,
			want:      []byte("[=>]"),
		},
		{
			termWidth: 6,
			barWidth:  60,
			want:      []byte("[==>-]"),
		},
		{
			termWidth: 8,
			barWidth:  60,
			want:      []byte("[====>-]"),
		},
		{
			termWidth: 80,
			barWidth:  60,
			want:      []byte("[===========================================>--------------]"),
		},
		{
			termWidth: 80,
			barWidth:  62,
			want:      []byte("[============================================>---------------]"),
		},
	}

	for _, test := range tests {
		s.barWidth = test.barWidth
		got := s.draw(test.termWidth)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

func newTestState(total, current int64) *state {
	return &state{
		format:         formatRunes{'[', '=', '>', '-', ']'},
		total:          total,
		current:        current,
		trimLeftSpace:  true,
		trimRightSpace: true,
	}
}
