package mpb

import (
	"io"
	"strings"

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

func (s *barFiller) Fill(w io.Writer, width int, stat *decor.Statistics) {

	str := string(s.format[rLeft])

	// don't count rLeft and rRight [brackets]
	width -= 2

	if width <= 2 {
		io.WriteString(w, str+string(s.format[rRight]))
		return
	}

	progressWidth := internal.Percentage(stat.Total, stat.Current, int64(width))
	needTip := progressWidth < int64(width) && progressWidth > 0

	if needTip {
		progressWidth--
	}

	if s.refill != nil {
		refillCount := internal.Percentage(stat.Total, s.refill.limit, int64(width))
		rest := progressWidth - refillCount
		str += runeRepeat(s.refill.r, int(refillCount)) + runeRepeat(s.format[rFill], int(rest))
	} else {
		str += runeRepeat(s.format[rFill], int(progressWidth))
	}

	if needTip {
		str += string(s.format[rTip])
		progressWidth++
	}

	rest := int64(width) - progressWidth
	str += runeRepeat(s.format[rEmpty], int(rest)) + string(s.format[rRight])
	io.WriteString(w, str)
}

func runeRepeat(r rune, count int) string {
	return strings.Repeat(string(r), count)
}
