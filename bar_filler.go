package mpb

import (
	"io"
	"unicode/utf8"

	"github.com/vbauerster/mpb/v4/decor"
	"github.com/vbauerster/mpb/v4/internal"
)

const (
	rLeft = iota
	rFill
	rTip
	rEmpty
	rRight
	rRevTip
	rRefill
)

var defaultBarStyle = "[=>-]<+"

type barFiller struct {
	format      [][]byte
	refillCount int
	reverse     bool
}

func newDefaultBarFiller() Filler {
	bf := &barFiller{
		format: make([][]byte, utf8.RuneCountInString(defaultBarStyle)),
	}
	bf.setStyle(defaultBarStyle)
	return bf
}

func (s *barFiller) setStyle(style string) {
	if !utf8.ValidString(style) {
		return
	}
	src := make([][]byte, 0, utf8.RuneCountInString(style))
	for _, r := range style {
		src = append(src, []byte(string(r)))
	}
	copy(s.format, src)
}

func (s *barFiller) setReverse() {
	s.reverse = true
}

func (s *barFiller) Fill(w io.Writer, width int, stat *decor.Statistics) {

	// don't count rLeft and rRight [brackets]
	width -= 2
	if width < 2 {
		return
	}

	w.Write(s.format[rLeft])
	if width == 2 {
		w.Write(s.format[rRight])
		return
	}

	bb := make([][]byte, width)

	cwidth := int(internal.Percentage(stat.Total, stat.Current, int64(width)))

	for i := 0; i < cwidth; i++ {
		bb[i] = s.format[rFill]
	}

	if s.refillCount > 0 {
		var rwidth int
		if s.refillCount > cwidth {
			rwidth = cwidth
		} else {
			rwidth = int(internal.Percentage(stat.Total, int64(s.refillCount), int64(width)))
		}
		for i := 0; i < rwidth; i++ {
			bb[i] = s.format[rRefill]
		}
	}

	if cwidth < width && cwidth > 0 {
		if s.reverse {
			bb[cwidth-1] = s.format[rRevTip]
		} else {
			bb[cwidth-1] = s.format[rTip]
		}
	}

	for i := cwidth; i < width; i++ {
		bb[i] = s.format[rEmpty]
	}

	if s.reverse {
		for i, j := 0, len(bb)-1; i < j; i, j = i+1, j-1 {
			bb[i], bb[j] = bb[j], bb[i]
		}
	}

	for i := 0; i < width; i++ {
		w.Write(bb[i])
	}
	w.Write(s.format[rRight])
}

func (s *barFiller) SetRefill(count int) {
	s.refillCount = count
}
