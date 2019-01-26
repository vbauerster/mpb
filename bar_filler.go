package mpb

import (
	"bytes"
	"io"

	"github.com/vbauerster/mpb/decor"
	"github.com/vbauerster/mpb/internal"
)

const (
	rLeft = iota
	rFill
	rTip
	rEmpty
	rRight
)

var defaultBarStyle = []rune("[=>-]")

type barFiller struct {
	format []rune
	refill *refill
}

func (s *barFiller) fill(w io.Writer, width int, stat *decor.Statistics) {

	w.Write([]byte(string(s.format[rLeft])))

	// don't count rLeft and rRight [brackets]
	width -= 2

	if width <= 2 {
		w.Write([]byte(string(s.format[rRight])))
		return
	}

	progressWidth := internal.Percentage(stat.Total, stat.Current, int64(width))
	needTip := progressWidth < int64(width) && progressWidth > 0

	if needTip {
		progressWidth--
	}

	if s.refill != nil {
		// append refill rune
		times := internal.Percentage(stat.Total, s.refill.limit, int64(width))
		w.Write(s.repeat(s.refill.r, int(times)))
		rest := progressWidth - times
		w.Write(s.repeat(s.format[rFill], int(rest)))
	} else {
		w.Write(s.repeat(s.format[rFill], int(progressWidth)))
	}

	if needTip {
		w.Write([]byte(string(s.format[rTip])))
		progressWidth++
	}

	rest := int64(width) - progressWidth
	w.Write(s.repeat(s.format[rEmpty], int(rest)))
	w.Write([]byte(string(s.format[rRight])))
}

func (s *barFiller) repeat(r rune, count int) []byte {
	return bytes.Repeat([]byte(string(r)), count)
}
