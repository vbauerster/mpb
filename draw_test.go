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
		trimSpace      bool
		reverse        bool
		rup            int64
		want           string
	}{
		0: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "",
			},
		},
		1: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "",
			},
		},
		2: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "",
			},
		},
		3: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "",
			},
		},
		4: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[>-]",
			},
		},
		5: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[>--]",
			},
		},
		6: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     " [>-] ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[>---]",
			},
		},
		7: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     " [>--] ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[=>---]",
			},
		},
		8: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     " [>---] ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[=>----]",
			},
		},
		80: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     " [========================>---------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{60,20,80}trim",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[=========================>----------------------------------------------------]",
			},
		},
		100: {
			{
				name:     "t,c,bw{100,100,0}",
				total:    100,
				current:  0,
				barWidth: 100,
				want:     " [------------------------------------------------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,100,0}trim",
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
				name:      "t,c,bw{100,1,100}trim",
				total:     100,
				current:   1,
				barWidth:  100,
				trimSpace: true,
				want:      "[>-------------------------------------------------------------------------------------------------]",
			},
			{
				name:     "t,c,bw{100,33,100}",
				total:    100,
				current:  33,
				barWidth: 100,
				want:     " [===============================>----------------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw{100,33,100}trim",
				total:     100,
				current:   33,
				barWidth:  100,
				trimSpace: true,
				want:      "[===============================>------------------------------------------------------------------]",
			},
			{
				name:      "t,c,bw,rev{100,33,100}trim",
				total:     100,
				current:   33,
				barWidth:  100,
				trimSpace: true,
				reverse:   true,
				want:      "[------------------------------------------------------------------<===============================]",
			},
			{
				name:     "t,c,bw,rup{100,33,100,33}",
				total:    100,
				current:  33,
				barWidth: 100,
				rup:      33,
				want:     " [+++++++++++++++++++++++++++++++>----------------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw,rup{100,33,100,33}trim",
				total:     100,
				current:   33,
				barWidth:  100,
				rup:       33,
				trimSpace: true,
				want:      "[+++++++++++++++++++++++++++++++>------------------------------------------------------------------]",
			},
			{
				name:      "t,c,bw,rup,rev{100,33,100,33}trim",
				total:     100,
				current:   33,
				barWidth:  100,
				rup:       33,
				trimSpace: true,
				reverse:   true,
				want:      "[------------------------------------------------------------------<+++++++++++++++++++++++++++++++]",
			},
			{
				name:     "t,c,bw,rup{100,40,100,32}",
				total:    100,
				current:  40,
				barWidth: 100,
				rup:      33,
				want:     " [++++++++++++++++++++++++++++++++=====>----------------------------------------------------------] ",
			},
			{
				name:      "t,c,bw,rup{100,40,100,32}trim",
				total:     100,
				current:   40,
				barWidth:  100,
				rup:       33,
				trimSpace: true,
				want:      "[++++++++++++++++++++++++++++++++======>-----------------------------------------------------------]",
			},
			{
				name:     "t,c,bw{100,99,100}",
				total:    100,
				current:  99,
				barWidth: 100,
				want:     " [==============================================================================================>-] ",
			},
			{
				name:      "t,c,bw{100,99,100}trim",
				total:     100,
				current:   99,
				barWidth:  100,
				trimSpace: true,
				want:      "[================================================================================================>-]",
			},
			{
				name:     "t,c,bw{100,100,100}",
				total:    100,
				current:  100,
				barWidth: 100,
				want:     " [================================================================================================] ",
			},
			{
				name:      "t,c,bw{100,100,100}trim",
				total:     100,
				current:   100,
				barWidth:  100,
				trimSpace: true,
				want:      "[==================================================================================================]",
			},
		},
	}

	var tmpBuf bytes.Buffer
	for termWidth, cases := range testSuite {
		for _, tc := range cases {
			s := newTestState(tc.reverse)
			s.reqWidth = tc.barWidth
			s.total = tc.total
			s.current = tc.current
			s.trimSpace = tc.trimSpace
			if tc.rup > 0 {
				if f, ok := s.filler.(interface{ SetRefill(int64) }); ok {
					f.SetRefill(tc.rup)
				}
			}
			tmpBuf.Reset()
			tmpBuf.ReadFrom(s.draw(newStatistics(termWidth, s)))
			by := tmpBuf.Bytes()
			by = by[:len(by)-1]

			got := string(by)
			if got != tc.want {
				t.Errorf("termWidth:%d %q want: %q %d, got: %q %d\n", termWidth, tc.name, tc.want, len(tc.want), got, len(got))
			}
		}
	}
}

func newTestState(reverse bool) *bState {
	s := &bState{
		filler: NewBarFiller(DefaultBarStyle, reverse),
		bufP:   new(bytes.Buffer),
		bufB:   new(bytes.Buffer),
		bufA:   new(bytes.Buffer),
	}
	return s
}
