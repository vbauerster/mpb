package mpb

import (
	"bytes"
	"io"

	"github.com/acarl005/stripansi"
	"github.com/mattn/go-runewidth"
	"github.com/vbauerster/mpb/v6/decor"
	"github.com/vbauerster/mpb/v6/internal"
)

const (
	iLbound = iota
	iRbound
	iFiller
	iRefiller
	iPadding
	components
)

// BarStyleComposer interface.
type BarStyleComposer interface {
	BarFillerBuilder
	Lbound(string) BarStyleComposer
	Rbound(string) BarStyleComposer
	Filler(string) BarStyleComposer
	Refiller(string) BarStyleComposer
	Padding(string) BarStyleComposer
	Tip(...string) BarStyleComposer
	Reverse() BarStyleComposer
}

type bFiller struct {
	components [components]*component
	tip        struct {
		count  uint
		frames []*component
	}
	flush func(dst io.Writer, src [][]byte, pad *component, padWidth int)
}

type component struct {
	width int
	bytes []byte
}

type barStyle struct {
	lbound   string
	rbound   string
	filler   string
	refiller string
	padding  string
	tip      []string
	rev      bool
}

// BarStyle constructs default bar style which can be altered via
// BarStyleComposer interface.
func BarStyle() BarStyleComposer {
	return &barStyle{
		lbound:   "[",
		rbound:   "]",
		filler:   "=",
		refiller: "+",
		padding:  "-",
		tip:      []string{">"},
	}
}

func (s *barStyle) Lbound(bound string) BarStyleComposer {
	s.lbound = bound
	return s
}

func (s *barStyle) Rbound(bound string) BarStyleComposer {
	s.rbound = bound
	return s
}

func (s *barStyle) Filler(filler string) BarStyleComposer {
	s.filler = filler
	return s
}

func (s *barStyle) Refiller(refiller string) BarStyleComposer {
	s.refiller = refiller
	return s
}

func (s *barStyle) Padding(padding string) BarStyleComposer {
	s.padding = padding
	return s
}

func (s *barStyle) Tip(tip ...string) BarStyleComposer {
	if len(tip) != 0 {
		s.tip = append(s.tip[:0], tip...)
	}
	return s
}

func (s *barStyle) Reverse() BarStyleComposer {
	s.rev = true
	return s
}

func (s *barStyle) Build() BarFiller {
	bf := &bFiller{
		flush: regFlush,
	}
	if s.rev {
		bf.flush = revFlush
	}
	bf.components[iLbound] = &component{
		width: runewidth.StringWidth(stripansi.Strip(s.lbound)),
		bytes: []byte(s.lbound),
	}
	bf.components[iRbound] = &component{
		width: runewidth.StringWidth(stripansi.Strip(s.rbound)),
		bytes: []byte(s.rbound),
	}
	bf.components[iFiller] = &component{
		width: runewidth.StringWidth(stripansi.Strip(s.filler)),
		bytes: []byte(s.filler),
	}
	bf.components[iRefiller] = &component{
		width: runewidth.StringWidth(stripansi.Strip(s.refiller)),
		bytes: []byte(s.refiller),
	}
	bf.components[iPadding] = &component{
		width: runewidth.StringWidth(stripansi.Strip(s.padding)),
		bytes: []byte(s.padding),
	}
	bf.tip.frames = make([]*component, len(s.tip))
	for i, t := range s.tip {
		bf.tip.frames[i] = &component{
			width: runewidth.StringWidth(stripansi.Strip(t)),
			bytes: []byte(t),
		}
	}
	return bf
}

func (s *bFiller) Fill(w io.Writer, width int, stat decor.Statistics) {
	width = internal.CheckRequestedWidth(width, stat.AvailableWidth)
	brackets := s.components[iLbound].width + s.components[iRbound].width
	if width < brackets {
		return
	}
	// don't count brackets as progress
	width -= brackets

	w.Write(s.components[iLbound].bytes)
	defer w.Write(s.components[iRbound].bytes)

	curWidth := int(internal.PercentageRound(stat.Total, stat.Current, width))
	padWidth := width - curWidth
	index, refill := 0, 0
	bb := make([][]byte, curWidth)

	if curWidth > 0 && curWidth != width {
		tipFrame := s.tip.frames[s.tip.count%uint(len(s.tip.frames))]
		bb[index] = tipFrame.bytes
		curWidth -= tipFrame.width
		s.tip.count++
		index++
	}

	if stat.Refill > 0 {
		refill = int(internal.PercentageRound(stat.Total, int64(stat.Refill), width))
		if refill > curWidth {
			refill = curWidth
		}
		curWidth -= refill
	}

	for curWidth > 0 {
		bb[index] = s.components[iFiller].bytes
		curWidth -= s.components[iFiller].width
		index++
	}

	for refill > 0 {
		bb[index] = s.components[iRefiller].bytes
		refill -= s.components[iRefiller].width
		index++
	}

	if curWidth+refill < 0 || s.components[iPadding].width > 1 {
		buf := new(bytes.Buffer)
		s.flush(buf, bb[:index], s.components[iPadding], padWidth)
		io.WriteString(w, runewidth.Truncate(buf.String(), width, "…"))
		return
	}

	s.flush(w, bb, s.components[iPadding], padWidth)
}

func regFlush(dst io.Writer, src [][]byte, pad *component, padWidth int) {
	for i := len(src) - 1; i >= 0; i-- {
		dst.Write(src[i])
	}
	for padWidth > 0 {
		dst.Write(pad.bytes)
		padWidth -= pad.width
	}
}

func revFlush(dst io.Writer, src [][]byte, pad *component, padWidth int) {
	for padWidth > 0 {
		dst.Write(pad.bytes)
		padWidth -= pad.width
	}
	for i := 0; i < len(src); i++ {
		dst.Write(src[i])
	}
}
