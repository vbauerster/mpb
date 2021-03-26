package complexfiller

import (
	"bytes"
	"io"
	"reflect"
	"regexp"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
	"github.com/vbauerster/mpb/v6/decor"
	"github.com/vbauerster/mpb/v6/internal"
)

const (
	rLeft = iota
	rFill
	rTip
	rSpace
	rRight
	rRevTip
	rRefill
)

// BarComplexDefaultStyle is a style for rendering a progress bar.
// It consists of a slice containing 7 ordered strings:
//
//	'1st string' stands for left boundary string
//
//	'2nd string' stands for fill string
//
//	'3rd string' stands for tip string
//
//	'4th string' stands for space string
//
//	'5th string' stands for right boundary string
//
//	'6th string' stands for reverse tip string
//
//	'7th stirng' stands for refill string
//
var BarDefaultStyle = []string{"[", "=", ">", "-", "]", "<", "+"}

type barFiller struct {
	format  [][]byte
	rwidth  []int
	tip     []byte
	refill  int64
	reverse bool
	flush   func(io.Writer, *space, [][]byte)
}

type space struct {
	space  []byte
	rwidth int
	count  int
}

// NewBarComplexFiller returns a BarComplexFiller implementation which renders a
// progress bar in regular direction. If style is empty string,
// BarDefaultStyle is applied. To be used with `*Progress.Add(...)
// *Bar` method.
func NewBarComplexFiller(style ...string) BarFiller {
	return newBarComplexFiller(false, style...)
}

// NewBarComplexFillerRev returns a BarComplexFiller implementation which renders a
// progress bar in reverse direction. If style is empty string,
// BarDefaultStyle is applied. To be used with `*Progress.Add(...)
// *Bar` method.
func NewBarComplexFillerRev(style ...string) BarFiller {
	return newBarComplexFiller(true, style...)
}

// NewBarComplexFillerPick pick between regular and reverse BarComplexFiller implementation
// based on rev param. To be used with `*Progress.Add(...) *Bar` method.
func NewBarComplexFillerPick(rev bool, style ...string) BarFiller {
	return newBarComplexFiller(rev, style...)
}

func newBarComplexFiller(rev bool, style ...string) BarFiller {
	bf := &barFiller{
		format:  make([][]byte, len(BarDefaultStyle)),
		rwidth:  make([]int, len(BarDefaultStyle)),
		reverse: rev,
	}
	if len(style) != 0 && !reflect.DeepEqual(style, BarDefaultStyle) {
		bf.parseComplex(style)
	} else {
		bf.parseComplex(BarDefaultStyle)
	}
	return bf
}

func (s *barFiller) parseComplex(style []string) {
	srcFormat := make([][]byte, 0, len(BarDefaultStyle))
	srcRwidth := make([]int, 0, len(BarDefaultStyle))
	for _, el := range style {
		if !utf8.ValidString(el) {
			panic("invalid bar style")
		}
		tmpFormat := []byte{}
		tmpRwidth := 0
		gr := uniseg.NewGraphemes(el)
		for gr.Next() {
			tmpFormat = append(tmpFormat, gr.Bytes()...)
			tmpRwidth += runewidth.StringWidth(gr.Str())
		}

		// Width fix for Bash colour codes, `NewGraphemes()` treats like normal characters (makes sense)
		r := regexp.MustCompile(`\[[0-9;]+m`)
		matches := r.FindAllString(el, -1)
		for _, el := range matches {
			for range el {
				tmpRwidth--
			}
		}

		srcFormat = append(srcFormat, tmpFormat)
		srcRwidth = append(srcRwidth, tmpRwidth)
	}
	copy(s.format, srcFormat)
	copy(s.rwidth, srcRwidth)
	if s.reverse {
		s.tip = s.format[rRevTip]
		s.flush = reverseFlush
	} else {
		s.tip = s.format[rTip]
		s.flush = regularFlush
	}
}

func (s *barFiller) Fill(w io.Writer, reqWidth int, stat decor.Statistics) {
	width := internal.CheckRequestedWidth(reqWidth, stat.AvailableWidth)
	brackets := s.rwidth[rLeft] + s.rwidth[rRight]
	if width < brackets {
		return
	}
	// don't count brackets as progress
	width -= brackets

	w.Write(s.format[rLeft])
	defer w.Write(s.format[rRight])

	cwidth := int(internal.PercentageRound(stat.Total, stat.Current, width))
	space := &space{
		space:  s.format[rSpace],
		rwidth: s.rwidth[rSpace],
		count:  width - cwidth,
	}

	index, refill := 0, 0
	bb := make([][]byte, cwidth)

	if cwidth > 0 && cwidth != width {
		bb[index] = s.tip
		cwidth -= s.rwidth[rTip]
		index++
	}

	if stat.Refill > 0 {
		refill = int(internal.PercentageRound(stat.Total, int64(stat.Refill), width))
		if refill > cwidth {
			refill = cwidth
		}
		cwidth -= refill
	}

	for cwidth > 0 {
		bb[index] = s.format[rFill]
		cwidth -= s.rwidth[rFill]
		index++
	}

	for refill > 0 {
		bb[index] = s.format[rRefill]
		refill -= s.rwidth[rRefill]
		index++
	}

	if cwidth+refill < 0 || space.rwidth > 1 {
		buf := new(bytes.Buffer)
		s.flush(buf, space, bb[:index])
		io.WriteString(w, runewidth.Truncate(buf.String(), width, "…"))
		return
	}

	s.flush(w, space, bb)
}

func regularFlush(w io.Writer, space *space, bb [][]byte) {
	for i := len(bb) - 1; i >= 0; i-- {
		w.Write(bb[i])
	}
	for space.count > 0 {
		w.Write(space.space)
		space.count -= space.rwidth
	}
}

func reverseFlush(w io.Writer, space *space, bb [][]byte) {
	for space.count > 0 {
		w.Write(space.space)
		space.count -= space.rwidth
	}
	for i := 0; i < len(bb); i++ {
		w.Write(bb[i])
	}
}
