package mpb

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestDraw(t *testing.T) {
	// key is termWidth
	testSuite := map[int][]struct {
		name           string
		total, current int64
		barWidth       int
		trimSpace      bool
		rup            int64
		want           string
	}{
		2: {
			{
				name:     "t,c,bw{60,20,80}",
				total:    60,
				current:  20,
				barWidth: 80,
				want:     "  ",
			},
			{
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{60,20,80,true}",
				total:     60,
				current:   20,
				barWidth:  80,
				trimSpace: true,
				want:      "[]",
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
				name:      "t,c,bw,trim{60,20,80,true}",
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
				want:     " [] ",
			},
			{
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{60,20,80,true}",
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
				name:      "t,c,bw,trim{100,100,0,true}",
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
				name:      "t,c,bw,trim{100,1,100,true}",
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
				name:      "t,c,bw,trim{100,33,100,true}",
				total:     100,
				current:   33,
				barWidth:  100,
				trimSpace: true,
				want:      "[===============================>------------------------------------------------------------------]",
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
				name:      "t,c,bw,rup,trim{100,33,100,33,true}",
				total:     100,
				current:   33,
				barWidth:  100,
				rup:       33,
				trimSpace: true,
				want:      "[+++++++++++++++++++++++++++++++>------------------------------------------------------------------]",
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
				name:      "t,c,bw,rup,trim{100,40,100,32,true}",
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
				name:      "t,c,bw,trim{100,99,100,true}",
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
				name:      "t,c,bw,trim{100,100,100,true}",
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
			s := newTestState()
			s.width = tc.barWidth
			s.total = tc.total
			s.current = tc.current
			s.trimSpace = tc.trimSpace
			if tc.rup > 0 {
				if f, ok := s.filler.(interface{ SetRefill(int64) }); ok {
					f.SetRefill(tc.rup)
				}
			}
			tmpBuf.Reset()
			tmpBuf.ReadFrom(s.draw(termWidth))
			by := tmpBuf.Bytes()
			by = by[:len(by)-1]

			if utf8.RuneCount(by) > termWidth {
				t.Errorf("termWidth:%d %q barWidth:%d overflow termWidth\n", termWidth, tc.name, utf8.RuneCount(by))
			}

			got := string(by)
			if got != tc.want {
				t.Errorf("termWidth:%d %q want: %q %d, got: %q %d\n", termWidth, tc.name, tc.want, len(tc.want), got, len(got))
			}
		}
	}
}

func newTestState() *bState {
	s := &bState{
		filler: newDefaultBarFiller(),
		bufP:   new(bytes.Buffer),
		bufB:   new(bytes.Buffer),
		bufA:   new(bytes.Buffer),
	}
	return s
}
