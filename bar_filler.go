package mpb

import (
	"bytes"
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
	rRefill
)

var defaultBarStyle = "[=>-]+"

type barFiller struct {
	format [][]byte
	rup    int
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
		style = defaultBarStyle
	}
	src := make([][]byte, 0, utf8.RuneCountInString(style))
	for _, r := range style {
		src = append(src, []byte(string(r)))
	}
	copy(s.format, src)
}

func (s *barFiller) Fill(w io.Writer, width int, stat *decor.Statistics) {

	b := s.format[rLeft]

	// don't count rLeft and rRight [brackets]
	width -= 2

	if width < 2 {
		return
	} else if width == 2 {
		w.Write(append(b, s.format[rRight]...))
		return
	}

	cwidth := internal.Percentage(stat.Total, stat.Current, int64(width))

	if s.rup > 0 {
		rwidth := internal.Percentage(stat.Total, int64(s.rup), int64(width))
		b = append(b, bytes.Repeat(s.format[rRefill], int(rwidth))...)
		rest := cwidth - rwidth
		b = append(b, bytes.Repeat(s.format[rFill], int(rest))...)
	} else {
		b = append(b, bytes.Repeat(s.format[rFill], int(cwidth))...)
	}

	if cwidth < int64(width) && cwidth > 0 {
		_, size := utf8.DecodeLastRune(b)
		b = append(b[:len(b)-size], s.format[rTip]...)
	}

	rest := int64(width) - cwidth
	b = append(b, bytes.Repeat(s.format[rEmpty], int(rest))...)
	w.Write(append(b, s.format[rRight]...))
}

func (s *barFiller) SetRefill(upto int) {
	s.rup = upto
}
