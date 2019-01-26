package mpb

import (
	"bytes"
	"testing"
)

func TestDraw(t *testing.T) {
	// key is termWidth
	testSuite := map[int][]struct {
		name           string
		total, current int64
		barWidth       int
		barRefill      *refill
		trimSpace      bool
		want           string
	}{
		100: {
			{
				name:     "t,c,bw{100,100,0}",
				total:    100,
				current:  0,
				barWidth: 100,
				want:     " [------------------------------------------------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,100,0}:trimSpace",
				total:     100,
				current:   0,
				barWidth:  100,
				trimSpace: true,
				want:      "[--------------------------------------------------------------------------------------------------]",
			},
			{
				name:     "t,c,bw{100,1,100}",
				total:    100,
				current:  1,
				barWidth: 100,
				want:     " [>-----------------------------------------------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,1,100}:trimSpace",
				total:     100,
				current:   1,
				barWidth:  100,
				trimSpace: true,
				want:      "[>-------------------------------------------------------------------------------------------------]",
			},
			{
				name:     "t,c,bw{100,40,100}",
				total:    100,
				current:  40,
				barWidth: 100,
				want:     " [=====================================>----------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,40,100}:refill{'+', 32}",
				total:     100,
				current:   40,
				barWidth:  100,
				barRefill: &refill{'+', 32},
				want:      " [+++++++++++++++++++++++++++++++======>----------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,40,100}:refill{'+', 32}:trimSpace",
				total:     100,
				current:   40,
				barWidth:  100,
				barRefill: &refill{'+', 32},
				trimSpace: true,
				want:      "[+++++++++++++++++++++++++++++++=======>-----------------------------------------------------------]",
			},
			{
				name:     "t,c,bw{100,99,100}",
				total:    100,
				current:  99,
				barWidth: 100,
				want:     " [==============================================================================================>-] ",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [================================================================================================] ",
			},
		},
		2: {
			{
				name:     "t,c,bw{0,0,100}",
				barWidth: 100,
				want:     " [] ",
			},
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     " [] ",
			},
		},
		4: {
			{
				name:     "t,c,bw{100,20,100}",
				total:    100,
				current:  20,
				barWidth: 100,
				want:     " [] ",
			},
			{
				name:     "t,c,bw{100,98,100}",
				total:    100,
				current:  98,
				barWidth: 100,
				want:     " [] ",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [] ",
			},
		},
		8: {
			{
				name:     "t,c,bw{100,20,100}",
				total:    100,
				current:  20,
				barWidth: 100,
				want:     " [>---] ",
			},
			{
				name:     "t,c,bw{100,98,100}",
				total:    100,
				current:  98,
				barWidth: 100,
				want:     " [====] ",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [====] ",
			},
		},
		20: {
			{
				name:     "t,c,bw{100,20,100}",
				total:    100,
				current:  20,
				barWidth: 100,
				want:     " [==>-------------] ",
			},
			{
				name:     "t,c,bw{100,60,100}",
				total:    100,
				current:  60,
				barWidth: 100,
				want:     " [=========>------] ",
			},
			{
				name:     "t,c,bw{100,98,100}",
				total:    100,
				current:  98,
				barWidth: 100,
				want:     " [================] ",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [================] ",
			},
		},
		50: {
			{
				name:     "t,c,bw{100,20,100}",
				total:    100,
				current:  20,
				barWidth: 100,
				want:     " [========>-------------------------------------] ",
			},
			{
				name:     "t,c,bw{100,60,100}",
				total:    100,
				current:  60,
				barWidth: 100,
				want:     " [===========================>------------------] ",
			},
			{
				name:     "t,c,bw{100,98,100}",
				total:    100,
				current:  98,
				barWidth: 100,
				want:     " [============================================>-] ",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [==============================================] ",
			},
		},
	}

	var tmpBuf bytes.Buffer
	for termWidth, cases := range testSuite {
		for _, tc := range cases {
			s := newTestState()
			s.width = tc.barWidth
			s.total = tc.total
			s.current = tc.current
			s.trimSpace = tc.trimSpace
			if tc.barRefill != nil {
				s.filler.(*barFiller).refill = tc.barRefill
			}
			tmpBuf.Reset()
			tmpBuf.ReadFrom(s.draw(termWidth))
			by := tmpBuf.Bytes()
			by = by[:len(by)-1]
			got := string(by)
			if got != tc.want {
				t.Errorf("termWidth %d; %s: want: %q %d, got: %q %d\n", termWidth, tc.name, tc.want, len(tc.want), got, len(got))
			}
		}
	}
}

func newTestState() *bState {
	s := &bState{
		filler: &barFiller{format: defaultBarStyle},
		bufP:   new(bytes.Buffer),
		bufB:   new(bytes.Buffer),
		bufA:   new(bytes.Buffer),
	}
	return s
}
