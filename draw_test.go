package mpb

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestDrawDefault(t *testing.T) {
	// key is termWidth
	testSuite := map[int][]struct {
		style    BarStyleComposer
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
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		1: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		2: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[]",
			},
		},
		3: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[-]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    "  ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    "  ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=]",
			},
		},
		4: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>-]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==]",
			},
		},
		5: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [-] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>--]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [>] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[==>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [=] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[===]",
			},
		},
		6: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>-] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>---]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [=>] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[===>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [==] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[====]",
			},
		},
		7: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>--] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>---]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==>] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[====>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=====]",
			},
		},
		8: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>---] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>----]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [===>] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=====>]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [====] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[======]",
			},
		},
		80: {
			{
				style:   BarStyle(),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [========================>---------------------------------------------------] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=========================>----------------------------------------------------]",
			},
			{
				style:    BarStyle(),
				name:     "t,c,bw{60,20,60}",
				total:    60,
				current:  20,
				barWidth: 60,
				want:     " [==================>---------------------------------------] ",
			},
			{
				style:    BarStyle(),
				name:     "t,c,bw{60,20,60}trim",
				total:    60,
				current:  20,
				barWidth: 60,
				trim:     true,
				want:     "[==================>---------------------------------------]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==========================================================================>-] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[============================================================================>-]",
			},
			{
				style:    BarStyle(),
				name:     "t,c,bw{60,59,60}",
				total:    60,
				current:  59,
				barWidth: 60,
				want:     " [========================================================>-] ",
			},
			{
				style:    BarStyle(),
				name:     "t,c,bw{60,59,60}trim",
				total:    60,
				current:  59,
				barWidth: 60,
				trim:     true,
				want:     "[========================================================>-]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [============================================================================] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==============================================================================]",
			},
			{
				style:    BarStyle(),
				name:     "t,c,bw{60,60,60}",
				total:    60,
				current:  60,
				barWidth: 60,
				want:     " [==========================================================] ",
			},
			{
				style:    BarStyle(),
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
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,1}Tip("だ")`,
				total:   100,
				current: 1,
				want:    " [だつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつ…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,2}Tip("だ")`,
				total:   100,
				current: 2,
				want:    " [だつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつつ…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,99}Tip("だ")`,
				total:   100,
				current: 99,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,100}Tip("だ")`,
				total:   100,
				current: 100,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののの…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").Padding("つ").Rbound("]").Reverse(),
				name:    `t,c{100,100}Tip("だ")rev`,
				total:   100,
				current: 100,
				want:    " […ののののののののののののののののののののののののののののののののののののののののののののののの] ",
			},
		},
		100: {
			{
				style:   BarStyle(),
				name:    "t,c{100,0}",
				total:   100,
				current: 0,
				want:    " [------------------------------------------------------------------------------------------------] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,0}trim",
				total:   100,
				current: 0,
				trim:    true,
				want:    "[--------------------------------------------------------------------------------------------------]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,1}",
				total:   100,
				current: 1,
				want:    " [>-----------------------------------------------------------------------------------------------] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,1}trim",
				total:   100,
				current: 1,
				trim:    true,
				want:    "[>-------------------------------------------------------------------------------------------------]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [==============================================================================================>-] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[================================================================================================>-]",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [================================================================================================] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[==================================================================================================]",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=]",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				style:   BarStyle().Tip("", "<").Reverse(),
				name:    "t,c,r{100,100,99}rev",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().Tip("", "<").Reverse(),
				name:    "t,c,r{100,100,99}trim,rev",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				style:   BarStyle().Tip("", "<").Reverse(),
				name:    "t,c,r{100,100,100}rev",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().Tip("", "<").Reverse(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,40,33}",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [++++++++++++++++++++++++++++++++=====>----------------------------------------------------------] ",
			},
			{
				style:   BarStyle(),
				name:    "t,c,r{100,40,33}trim",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++======>-----------------------------------------------------------]",
			},
			{
				style:   BarStyle().Tip("<").Reverse(),
				name:    "t,c,r{100,40,33},rev",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [----------------------------------------------------------<=====++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().Tip("<").Reverse(),
				name:    "t,c,r{100,40,33}trim,rev",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[-----------------------------------------------------------<======++++++++++++++++++++++++++++++++]",
			},
		},
	}

	var tmpBuf bytes.Buffer
	for tw, cases := range testSuite {
		for _, tc := range cases {
			s := newTestState(NewBarFiller(tc.style))
			s.reqWidth = tc.barWidth
			s.total = tc.total
			s.current = tc.current
			s.trimSpace = tc.trim
			s.refill = tc.refill
			tmpBuf.Reset()
			tmpBuf.ReadFrom(s.draw(newStatistics(tw, s)))
			by := tmpBuf.Bytes()

			got := string(by[:len(by)-1])
			if !utf8.ValidString(got) {
				t.Fail()
			}
			if got != tc.want {
				t.Errorf("termWidth:%d %q want: %q %d, got: %q %d\n", tw, tc.name, tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
			}
		}
	}
}

func TestDrawTipOnComplete(t *testing.T) {
	// key is termWidth
	testSuite := map[int][]struct {
		style    BarStyleComposer
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
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		1: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "",
			},
		},
		2: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[]",
			},
		},
		3: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    "  ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[-]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    "  ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    "  ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[>]",
			},
		},
		4: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>-]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=>]",
			},
		},
		5: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [-] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>--]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[==>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[==>]",
			},
		},
		6: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>-] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[>---]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [=>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[===>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [=>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[===>]",
			},
		},
		7: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>--] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>---]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[====>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [==>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[====>]",
			},
		},
		8: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [>---] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=>----]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [===>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[=====>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=====>]",
			},
		},
		80: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}",
				total:   60,
				current: 20,
				want:    " [========================>---------------------------------------------------] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,20}trim",
				total:   60,
				current: 20,
				trim:    true,
				want:    "[=========================>----------------------------------------------------]",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
				name:     "t,c,bw{60,20,60}",
				total:    60,
				current:  20,
				barWidth: 60,
				want:     " [==================>---------------------------------------] ",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
				name:     "t,c,bw{60,20,60}trim",
				total:    60,
				current:  20,
				barWidth: 60,
				trim:     true,
				want:     "[==================>---------------------------------------]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}",
				total:   60,
				current: 59,
				want:    " [==========================================================================>-] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,59}trim",
				total:   60,
				current: 59,
				trim:    true,
				want:    "[============================================================================>-]",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
				name:     "t,c,bw{60,59,60}",
				total:    60,
				current:  59,
				barWidth: 60,
				want:     " [========================================================>-] ",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
				name:     "t,c,bw{60,59,60}trim",
				total:    60,
				current:  59,
				barWidth: 60,
				trim:     true,
				want:     "[========================================================>-]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}",
				total:   60,
				current: 60,
				want:    " [===========================================================================>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{60,60}trim",
				total:   60,
				current: 60,
				trim:    true,
				want:    "[=============================================================================>]",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
				name:     "t,c,bw{60,60,60}",
				total:    60,
				current:  60,
				barWidth: 60,
				want:     " [=========================================================>] ",
			},
			{
				style:    BarStyle().TipOnComplete(">"),
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
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").TipOnComplete("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,99}Tip("だ").TipOnComplete("だ")`,
				total:   100,
				current: 99,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").TipOnComplete("だ").Padding("つ").Rbound("]"),
				name:    `t,c{100,100}Tip("だ").TipOnComplete("だ")`,
				total:   100,
				current: 100,
				want:    " [ののののののののののののののののののののののののののののののののののののののののののののののだ…] ",
			},
			{
				style:   BarStyle().Lbound("[").Filler("の").Tip("だ").TipOnComplete("だ").Padding("つ").Rbound("]").Reverse(),
				name:    `t,c{100,100}Tip("だ").TipOnComplete("だ")rev`,
				total:   100,
				current: 100,
				want:    " […だのののののののののののののののののののののののののののののののののののののののののののののの] ",
			},
		},
		100: {
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,0}",
				total:   100,
				current: 0,
				want:    " [------------------------------------------------------------------------------------------------] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,0}trim",
				total:   100,
				current: 0,
				trim:    true,
				want:    "[--------------------------------------------------------------------------------------------------]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,1}",
				total:   100,
				current: 1,
				want:    " [>-----------------------------------------------------------------------------------------------] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,1}trim",
				total:   100,
				current: 1,
				trim:    true,
				want:    "[>-------------------------------------------------------------------------------------------------]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,99}",
				total:   100,
				current: 99,
				want:    " [==============================================================================================>-] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,99}trim",
				total:   100,
				current: 99,
				trim:    true,
				want:    "[================================================================================================>-]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,100}",
				total:   100,
				current: 100,
				want:    " [===============================================================================================>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c{100,100}trim",
				total:   100,
				current: 100,
				trim:    true,
				want:    "[=================================================================================================>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,100,99}",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,100,99}trim",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,100,100}",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++>]",
			},
			{
				style:   BarStyle().TipOnComplete("").Reverse(),
				name:    "t,c,r{100,100,99}rev",
				total:   100,
				current: 100,
				refill:  99,
				want:    " [=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().TipOnComplete("").Reverse(),
				name:    "t,c,r{100,100,99}trim,rev",
				total:   100,
				current: 100,
				refill:  99,
				trim:    true,
				want:    "[=+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				style:   BarStyle().TipOnComplete("").Reverse(),
				name:    "t,c,r{100,100,100}rev",
				total:   100,
				current: 100,
				refill:  100,
				want:    " [++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().TipOnComplete("").Reverse(),
				name:    "t,c,r{100,100,100}trim",
				total:   100,
				current: 100,
				refill:  100,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++]",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,40,33}",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [++++++++++++++++++++++++++++++++=====>----------------------------------------------------------] ",
			},
			{
				style:   BarStyle().TipOnComplete(">"),
				name:    "t,c,r{100,40,33}trim",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[++++++++++++++++++++++++++++++++======>-----------------------------------------------------------]",
			},
			{
				style:   BarStyle().Tip("<").TipOnComplete("<").Reverse(),
				name:    "t,c,r{100,40,33},rev",
				total:   100,
				current: 40,
				refill:  33,
				want:    " [----------------------------------------------------------<=====++++++++++++++++++++++++++++++++] ",
			},
			{
				style:   BarStyle().Tip("<").TipOnComplete("<").Reverse(),
				name:    "t,c,r{100,40,33}trim,rev",
				total:   100,
				current: 40,
				refill:  33,
				trim:    true,
				want:    "[-----------------------------------------------------------<======++++++++++++++++++++++++++++++++]",
			},
		},
	}

	var tmpBuf bytes.Buffer
	for tw, cases := range testSuite {
		for _, tc := range cases {
			s := newTestState(NewBarFiller(tc.style))
			s.reqWidth = tc.barWidth
			s.total = tc.total
			s.current = tc.current
			s.trimSpace = tc.trim
			s.refill = tc.refill
			tmpBuf.Reset()
			tmpBuf.ReadFrom(s.draw(newStatistics(tw, s)))
			by := tmpBuf.Bytes()

			got := string(by[:len(by)-1])
			if !utf8.ValidString(got) {
				t.Fail()
			}
			if got != tc.want {
				t.Errorf("termWidth:%d %q want: %q %d, got: %q %d\n", tw, tc.name, tc.want, utf8.RuneCountInString(tc.want), got, utf8.RuneCountInString(got))
			}
		}
	}
}

func newTestState(filler BarFiller) *bState {
	s := &bState{
		filler: filler,
		bufP:   new(bytes.Buffer),
		bufB:   new(bytes.Buffer),
		bufA:   new(bytes.Buffer),
	}
	return s
}
