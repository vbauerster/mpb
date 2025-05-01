package mpb

import (
	"bytes"
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestDrawDefault(t *testing.T) {
	t.Parallel()
	// key is termWidth
	testSuite := map[int][]struct {
		filler   BarFiller
		name     string
		total    int64
		current  int64
		refill   int64
		barWidth int
		trim     bool
		want     string
	}{
		0: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		1: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		2: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[]",
			},
		},
		3: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[-]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    "  ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    "  ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=]",
			},
		},
		4: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>-]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==]",
			},
		},
		5: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [-] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>--]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [>] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[==>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [=] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[===]",
			},
		},
		6: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>-] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>---]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [=>] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[===>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [==] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[====]",
			},
		},
		7: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>--] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>---]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==>] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[====>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=====]",
			},
		},
		8: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>---] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>----]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [===>] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=====>]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [====] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[======]",
			},
		},
		80: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [========================>---------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=========================>----------------------------------------------------]",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,20,60}",
				total:    60,
				current:  20,
				barWidth: 60,
				want:     " [==================>---------------------------------------] ",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,20,60}trim",
				total:    60,
				current:  20,
				barWidth: 60,
				trim:     true,
				want:     "[==================>---------------------------------------]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==========================================================================>-] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[============================================================================>-]",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,59,60}",
				total:    60,
				current:  59,
				barWidth: 60,
				want:     " [========================================================>-] ",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,59,60}trim",
				total:    60,
				current:  59,
				barWidth: 60,
				trim:     true,
				want:     "[========================================================>-]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [============================================================================] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==============================================================================]",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,60,60}",
				total:    60,
				current:  60,
				barWidth: 60,
				want:     " [==========================================================] ",
			},
			{
				filler:   BarStyle().Build(),
				name:     "t,c,bw{60,60,60}trim",
				total:    60,
				current:  60,
				barWidth: 60,
				trim:     true,
				want:     "[==========================================================]",
			},
		},
		99: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,1}",
				total:   100,
				current: 1,
				want:    " [>----------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,1}trim",
				total:   100,
				current: 1,
				trim:    true,
				want:    "[>------------------------------------------------------------------------------------------------]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,40,33}",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [+++++++++++++++++++++++++++++++======>---------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,40,33}trim",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++======>----------------------------------------------------------]",
			},
			{
				filler:  BarStyle().Tip("<").Reverse().Build(),
				name:    "t,c,r{100,40,33},rev",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [---------------------------------------------------------<======+++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Tip("<").Reverse().Build(),
				name:    "t,c,r{100,40,33}trim,rev",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[----------------------------------------------------------<======++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [=============================================================================================>-] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[===============================================================================================>-]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [===============================================================================================] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[=================================================================================================]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=]",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,99}rev",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [=++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,99}trim,rev",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[=++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,100}rev",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
		},
		100: {
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,0}",
				total:   100,
				current: 0,
				want:    " [------------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,0}trim",
				total:   100,
				current: 0,
				trim:    true,
				want:    "[--------------------------------------------------------------------------------------------------]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,1}",
				total:   100,
				current: 1,
				want:    " [>-----------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Tip("").Build(),
				name:    "t,c{100,1}empty_tip",
				total:   100,
				current: 1,
				want:    " [=-----------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,1}trim",
				total:   100,
				current: 1,
				trim:    true,
				want:    "[>-------------------------------------------------------------------------------------------------]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [==============================================================================================>-] ",
			},
			{
				filler:  BarStyle().Tip("").Build(),
				name:    "t,c{100,99}empty_tip",
				total:   100,
				current: 99,
				want:    " [===============================================================================================-] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[================================================================================================>-]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [================================================================================================] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[==================================================================================================]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,99}rev",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,99}trim,rev",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,100}rev",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Reverse().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,40,33}",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [++++++++++++++++++++++++++++++++=====>----------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Build(),
				name:    "t,c,r{100,40,33}trim",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++======>-----------------------------------------------------------]",
			},
			{
				filler:  BarStyle().Tip("<").Reverse().Build(),
				name:    "t,c,r{100,40,33},rev",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [----------------------------------------------------------<=====++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Tip("<").Reverse().Build(),
				name:    "t,c,r{100,40,33}trim,rev",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[-----------------------------------------------------------<======++++++++++++++++++++++++++++++++]",
			},
		},
	}

	for tw, cases := range testSuite {
		t.Run(fmt.Sprintf("tw_%d", tw), func(t *testing.T) {
			for _, tc := range cases {
				// tc := tc // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					var tmpBuf bytes.Buffer
					ps := pState{reqWidth: tc.barWidth}
					s := ps.makeBarState(tc.total, tc.filler)
					s.current = tc.current
					s.trimSpace = tc.trim
					s.refill = tc.refill
					r, err := s.draw(s.newStatistics(tw))
					if err != nil {
						t.Fatalf("draw error: %s", err.Error())
					}
					_, err = tmpBuf.ReadFrom(r)
					if err != nil {
						t.Fatalf("read from r error: %s", err.Error())
					}
					var got string
					if by := tmpBuf.Bytes(); len(by) != 0 && by[len(by)-1] == '\n' {
						got = string(by[:len(by)-1])
					} else {
						got = string(by)
					}
					if !utf8.ValidString(got) {
						t.Fatalf("not valid utf8: %#v", got)
					}
					if got != tc.want {
						t.Errorf("want: %q %d, got: %q %d\n", tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
					}
				})
			}
		})
	}
}

func TestDrawTipOnComplete(t *testing.T) {
	t.Parallel()
	// key is termWidth
	testSuite := map[int][]struct {
		filler   BarFiller
		name     string
		total    int64
		current  int64
		refill   int64
		barWidth int
		trim     bool
		want     string
	}{
		3: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    "  ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[>]",
			},
		},
		4: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=>]",
			},
		},
		5: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[==>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==>]",
			},
		},
		6: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [=>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[===>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [=>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[===>]",
			},
		},
		7: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[====>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [==>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[====>]",
			},
		},
		8: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [===>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=====>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=====>]",
			},
		},
		80: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==========================================================================>-] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[============================================================================>-]",
			},
			{
				filler:   BarStyle().TipOnComplete().Build(),
				name:     "t,c,bw{60,59,60}",
				total:    60,
				current:  59,
				barWidth: 60,
				want:     " [========================================================>-] ",
			},
			{
				filler:   BarStyle().TipOnComplete().Build(),
				name:     "t,c,bw{60,59,60}trim",
				total:    60,
				current:  59,
				barWidth: 60,
				trim:     true,
				want:     "[========================================================>-]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===========================================================================>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=============================================================================>]",
			},
			{
				filler:   BarStyle().TipOnComplete().Build(),
				name:     "t,c,bw{60,60,60}",
				total:    60,
				current:  60,
				barWidth: 60,
				want:     " [=========================================================>] ",
			},
			{
				filler:   BarStyle().TipOnComplete().Build(),
				name:     "t,c,bw{60,60,60}trim",
				total:    60,
				current:  60,
				barWidth: 60,
				trim:     true,
				want:     "[=========================================================>]",
			},
		},
		99: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [=============================================================================================>-] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[===============================================================================================>-]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [==============================================================================================>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[================================================================================================>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
			{
				filler:  BarStyle().Tip("<").TipOnComplete().Reverse().Build(),
				name:    `t,c,r{100,100,99}.Tip("<").TipOnComplete().Reverse()`,
				total:   100,
				current: 100,
				refill:  99,
				want:    " [<++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Tip("<").TipOnComplete().Reverse().Build(),
				name:    `t,c,r{100,100,99}.Tip("<").TipOnComplete().Reverse()trim`,
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[<++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
			{
				filler:  BarStyle().Tip("<").TipOnComplete().Reverse().Build(),
				name:    `t,c,r{100,100,100}.Tip("<").TipOnComplete().Reverse()`,
				total:   100,
				current: 100,
				refill:  100,
				want:    " [<++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				filler:  BarStyle().Tip("<").TipOnComplete().Reverse().Build(),
				name:    `t,c,r{100,100,100}.Tip("<").TipOnComplete().Reverse()trim`,
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[<++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
		},
		100: {
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [==============================================================================================>-] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[================================================================================================>-]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [===============================================================================================>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[=================================================================================================>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				filler:  BarStyle().TipOnComplete().Build(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
		},
	}

	for tw, cases := range testSuite {
		t.Run(fmt.Sprintf("tw_%d", tw), func(t *testing.T) {
			for _, tc := range cases {
				// tc := tc // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					var tmpBuf bytes.Buffer
					ps := pState{reqWidth: tc.barWidth}
					s := ps.makeBarState(tc.total, tc.filler)
					s.current = tc.current
					s.trimSpace = tc.trim
					s.refill = tc.refill
					r, err := s.draw(s.newStatistics(tw))
					if err != nil {
						t.Fatalf("draw error: %s", err.Error())
					}
					_, err = tmpBuf.ReadFrom(r)
					if err != nil {
						t.Fatalf("read from r error: %s", err.Error())
					}
					var got string
					if by := tmpBuf.Bytes(); len(by) != 0 && by[len(by)-1] == '\n' {
						got = string(by[:len(by)-1])
					} else {
						got = string(by)
					}
					if !utf8.ValidString(got) {
						t.Fatalf("not valid utf8: %#v", got)
					}
					if got != tc.want {
						t.Errorf("want: %q %d, got: %q %d\n", tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
					}
				})
			}
		})
	}
}

func TestDrawDoubleWidth(t *testing.T) {
	t.Parallel()
	// key is termWidth
	testSuite := map[int][]struct {
		filler   BarFiller
		name     string
		total    int64
		current  int64
		refill   int64
		barWidth int
		trim     bool
		want     string
	}{
		99: {
			{
				filler:  BarStyle().Lbound("の").Rbound("の").Build(),
				name:    `t,c{100,1}.Lbound("の").Rbound("の")`,
				total:   100,
				current: 1,
				want:    " の>--------------------------------------------------------------------------------------------の ",
			},
			{
				filler:  BarStyle().Lbound("の").Rbound("の").Build(),
				name:    `t,c{100,1}.Lbound("の").Rbound("の")`,
				total:   100,
				current: 2,
				want:    " の=>-------------------------------------------------------------------------------------------の ",
			},
			{
				filler:  BarStyle().Tip("だ").Build(),
				name:    `t,c{100,1}Tip("だ")`,
				total:   100,
				current: 1,
				want:    " [だ---------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Tip("だ").Build(),
				name:    `t,c{100,2}Tip("だ")`,
				total:   100,
				current: 2,
				want:    " [だ---------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Tip("だ").Build(),
				name:    `t,c{100,3}Tip("だ")`,
				total:   100,
				current: 3,
				want:    " [=だ--------------------------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().Tip("だ").Build(),
				name:    `t,c{100,99}Tip("だ")`,
				total:   100,
				current: 99,
				want:    " [============================================================================================だ-] ",
			},
			{
				filler:  BarStyle().Tip("だ").Build(),
				name:    `t,c{100,100}Tip("だ")`,
				total:   100,
				current: 100,
				want:    " [===============================================================================================] ",
			},
			{
				filler:  BarStyle().Tip("だ").TipOnComplete().Build(),
				name:    `t,c{100,100}.Tip("だ").TipOnComplete()`,
				total:   100,
				current: 100,
				want:    " [=============================================================================================だ] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").Padding("つ").Build(),
				name:    `t,c{100,1}Filler("の").Tip("だ").Padding("つ")`,
				total:   100,
				current: 1,
				want:    " [だつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつ…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").Padding("つ").Build(),
				name:    `t,c{100,2}Filler("の").Tip("だ").Padding("つ")`,
				total:   100,
				current: 2,
				want:    " [だつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつ…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").Padding("つ").Build(),
				name:    `t,c{100,99}Filler("の").Tip("だ").Padding("つ")`,
				total:   100,
				current: 99,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").Padding("つ").Build(),
				name:    `t,c{100,100}.Filler("の").Tip("だ").Padding("つ")`,
				total:   100,
				current: 100,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののの…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").Padding("つ").Reverse().Build(),
				name:    `t,c{100,100}Filler("の").Tip("だ").Padding("つ").Reverse()`,
				total:   100,
				current: 100,
				want:    " […ののののののののののののののののののののののののののののののののののののののののののののののの] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").TipOnComplete().Padding("つ").Build(),
				name:    `t,c{100,99}Filler("の").Tip("だ").TipOnComplete().Padding("つ")`,
				total:   100,
				current: 99,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").TipOnComplete().Padding("つ").Build(),
				name:    `t,c{100,100}.Filler("の").Tip("だ").TipOnComplete().Padding("つ")`,
				total:   100,
				current: 100,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				filler:  BarStyle().Filler("の").Tip("だ").TipOnComplete().Padding("つ").Reverse().Build(),
				name:    `t,c{100,100}.Filler("の").Tip("だ").TipOnComplete().Padding("つ").Reverse()`,
				total:   100,
				current: 100,
				want:    " […だのののののののののののののののののののののののののののののののののののののののののののののの] ",
			},
			{
				filler:  BarStyle().Refiller("の").Build(),
				name:    `t,c,r{100,100,99}Refiller("の")`,
				total:   100,
				current: 100,
				refill:  99,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののの=] ",
			},
			{
				filler:  BarStyle().Refiller("の").Build(),
				name:    `t,c,r{100,100,99}Refiller("の")trim`,
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[のののののののののののののののののののののののののののののののののののののののののののののののの=]",
			},
		},
	}

	for tw, cases := range testSuite {
		t.Run(fmt.Sprintf("tw_%d", tw), func(t *testing.T) {
			for _, tc := range cases {
				// tc := tc // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					var tmpBuf bytes.Buffer
					ps := pState{reqWidth: tc.barWidth}
					s := ps.makeBarState(tc.total, tc.filler)
					s.current = tc.current
					s.trimSpace = tc.trim
					s.refill = tc.refill
					r, err := s.draw(s.newStatistics(tw))
					if err != nil {
						t.Fatalf("draw error: %s", err.Error())
					}
					_, err = tmpBuf.ReadFrom(r)
					if err != nil {
						t.Fatalf("read from r error: %s", err.Error())
					}
					var got string
					if by := tmpBuf.Bytes(); len(by) != 0 && by[len(by)-1] == '\n' {
						got = string(by[:len(by)-1])
					} else {
						got = string(by)
					}
					if !utf8.ValidString(got) {
						t.Fatalf("not valid utf8: %#v", got)
					}
					if got != tc.want {
						t.Errorf("want: %q %d, got: %q %d\n", tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
					}
				})
			}
		})
	}
}

func TestDrawMeta(t *testing.T) {
	t.Parallel()
	meta := func(s string) string { return "{" + s + "}" }
	nopMeta := func(s string) string { return s }
	// key is termWidth
	testSuite := map[int][]struct {
		filler   BarFiller
		name     string
		total    int64
		current  int64
		refill   int64
		barWidth int
		want     string
	}{
		80: {
			{
				filler:  BarStyle().LboundMeta(meta).RboundMeta(meta).Build(),
				name:    "LboundMeta RboundMeta",
				total:   100,
				current: 0,
				want:    " {[}----------------------------------------------------------------------------{]} ",
			},
			{
				filler:  BarStyle().TipMeta(meta).Build(),
				name:    "TipMeta 1",
				total:   100,
				current: 1,
				want:    " [{>}---------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().TipMeta(meta).Tip("").Build(),
				name:    "TipMeta EmptyTip 1",
				total:   100,
				current: 1,
				want:    " [={}---------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().FillerMeta(meta).Build(),
				name:    "FillerMeta 1",
				total:   100,
				current: 1,
				want:    " [{}>---------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().FillerMeta(meta).Build(),
				name:    "FillerMeta 2",
				total:   100,
				current: 2,
				want:    " [{=}>--------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().FillerMeta(meta).Build(),
				name:    "FillerMeta 4",
				total:   100,
				current: 4,
				want:    " [{==}>-------------------------------------------------------------------------] ",
			},
			{
				filler:  BarStyle().PaddingMeta(meta).Build(),
				name:    "PaddingMeta",
				total:   100,
				current: 0,
				want:    " [{----------------------------------------------------------------------------}] ",
			},
			{
				filler:  BarStyle().RefillerMeta(meta).Build(),
				name:    "RefillerMeta",
				total:   100,
				current: 80,
				refill:  50,
				want:    " [{++++++++++++++++++++++++++++++++++++++}======================>---------------] ",
			},
			{
				filler:  BarStyle().RefillerMeta(meta).FillerMeta(meta).PaddingMeta(meta).TipMeta(meta).Build(),
				name:    "RefillerMeta FillerMeta PaddingMeta TipMeta",
				total:   100,
				current: 80,
				refill:  50,
				want:    " [{++++++++++++++++++++++++++++++++++++++}{======================}{>}{---------------}] ",
			},
			{
				filler:  BarStyle().RefillerMeta(meta).FillerMeta(meta).PaddingMeta(meta).TipMeta(meta).LboundMeta(meta).RboundMeta(meta).Build(),
				name:    "RefillerMeta FillerMeta PaddingMeta TipMeta LboundMeta RboundMeta",
				total:   100,
				current: 80,
				refill:  50,
				want:    " {[}{++++++++++++++++++++++++++++++++++++++}{======================}{>}{---------------}{]} ",
			},
			{
				filler:  BarStyle().RefillerMeta(nopMeta).FillerMeta(nopMeta).PaddingMeta(nopMeta).TipMeta(nopMeta).LboundMeta(nopMeta).RboundMeta(nopMeta).Build(),
				name:    "RefillerMeta FillerMeta PaddingMeta TipMeta LboundMeta RboundMeta nopMeta",
				total:   100,
				current: 80,
				refill:  50,
				want:    " [++++++++++++++++++++++++++++++++++++++======================>---------------] ",
			},
			{
				filler:  BarStyle().RefillerMeta(nil).FillerMeta(nil).PaddingMeta(nil).TipMeta(nil).LboundMeta(nil).RboundMeta(nil).Build(),
				name:    "RefillerMeta FillerMeta PaddingMeta TipMeta LboundMeta RboundMeta nilMeta",
				total:   100,
				current: 80,
				refill:  50,
				want:    " [++++++++++++++++++++++++++++++++++++++======================>---------------] ",
			},
		},
	}

	for tw, cases := range testSuite {
		t.Run(fmt.Sprintf("tw_%d", tw), func(t *testing.T) {
			for _, tc := range cases {
				// tc := tc // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					var tmpBuf bytes.Buffer
					ps := pState{reqWidth: tc.barWidth}
					s := ps.makeBarState(tc.total, tc.filler)
					s.current = tc.current
					s.refill = tc.refill
					r, err := s.draw(s.newStatistics(tw))
					if err != nil {
						t.Fatalf("draw error: %s", err.Error())
					}
					_, err = tmpBuf.ReadFrom(r)
					if err != nil {
						t.Fatalf("read from r error: %s", err.Error())
					}
					var got string
					if by := tmpBuf.Bytes(); len(by) != 0 && by[len(by)-1] == '\n' {
						got = string(by[:len(by)-1])
					} else {
						got = string(by)
					}
					if !utf8.ValidString(got) {
						t.Fatalf("not valid utf8: %#v", got)
					}
					if got != tc.want {
						t.Errorf("want: %q %d, got: %q %d\n", tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
					}
				})
			}
		})
	}
}
