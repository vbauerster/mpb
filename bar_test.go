package mpb

import (
	"reflect"
	"testing"
)

func TestFillBar(t *testing.T) {
	b := newTestBar(80).SetEmpty('-').SetFill('=').SetTip('>').SetLeftEnd('[').SetRightEnd(']')
	tests := []struct {
		width int
		want  []byte
	}{
		{
			width: 1,
			want:  []byte{},
		},
		{
			width: 2,
			want:  []byte{'[', ']'},
		},
		{
			width: 3,
			want:  []byte{'[', '>', ']'},
		},
		{
			width: 4,
			want:  []byte{'[', '=', '>', ']'},
		},
	}

	for _, test := range tests {
		got := b.fillBar(80, 60, test.width)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

func newTestBar(width int) *Bar {
	b := &Bar{
		width: width,
	}
	return b
}
