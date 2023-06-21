package mpb

import (
	"io"

	"github.com/mattn/go-runewidth"
	"github.com/vbauerster/mpb/v8/decor"
	"github.com/vbauerster/mpb/v8/internal"
)

const (
	iLbound = iota
	iRbound
	iFiller
	iRefiller
	iPadding
	iTip
	components
)

var defaultBarStyle = [...]string{"[", "]", "=", "+", "-", ">"}

// BarStyleComposer interface.
type BarStyleComposer interface {
	BarFillerBuilder
	Lbound(string) BarStyleComposer
	LboundMeta(func(...interface{}) string) BarStyleComposer
	Rbound(string) BarStyleComposer
	RboundMeta(func(...interface{}) string) BarStyleComposer
	Filler(string) BarStyleComposer
	FillerMeta(func(...interface{}) string) BarStyleComposer
	Refiller(string) BarStyleComposer
	RefillerMeta(func(...interface{}) string) BarStyleComposer
	Padding(string) BarStyleComposer
	PaddingMeta(func(...interface{}) string) BarStyleComposer
	Tip(frames ...string) BarStyleComposer
	TipMeta(func(...interface{}) string) BarStyleComposer
	TipOnComplete() BarStyleComposer
	Reverse() BarStyleComposer
}

type bFiller struct {
	components    [components]component
	meta          [components]func(io.Writer, ...interface{}) error
	rev           bool
	tipOnComplete bool
	tip           struct {
		frames []component
		count  uint
	}
}

type component struct {
	width int
	bytes []byte
}

type barStyle struct {
	style         [components]string
	metaFuncs     [components]func(io.Writer, ...interface{}) error
	tipFrames     []string
	tipOnComplete bool
	rev           bool
}

// BarStyle constructs default bar style which can be altered via
// BarStyleComposer interface.
func BarStyle() BarStyleComposer {
	bs := &barStyle{
		style:     defaultBarStyle,
		tipFrames: []string{defaultBarStyle[iTip]},
	}
	defaultMeta := func(w io.Writer, a ...interface{}) (err error) {
		for i := 0; i < len(a) && err == nil; i++ {
			_, err = w.Write(a[i].([]byte))
		}
		return err
	}
	for i := range bs.metaFuncs {
		bs.metaFuncs[i] = defaultMeta
	}
	return bs
}

func (s *barStyle) Lbound(bound string) BarStyleComposer {
	s.style[iLbound] = bound
	return s
}

func (s *barStyle) LboundMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iLbound] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) Rbound(bound string) BarStyleComposer {
	s.style[iRbound] = bound
	return s
}

func (s *barStyle) RboundMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iRbound] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) Filler(filler string) BarStyleComposer {
	s.style[iFiller] = filler
	return s
}

func (s *barStyle) FillerMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iFiller] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) Refiller(refiller string) BarStyleComposer {
	s.style[iRefiller] = refiller
	return s
}

func (s *barStyle) RefillerMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iRefiller] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) Padding(padding string) BarStyleComposer {
	s.style[iPadding] = padding
	return s
}

func (s *barStyle) PaddingMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iPadding] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) Tip(frames ...string) BarStyleComposer {
	if len(frames) != 0 {
		s.tipFrames = frames
	}
	return s
}

func (s *barStyle) TipMeta(fn func(...interface{}) string) BarStyleComposer {
	s.metaFuncs[iTip] = makeMetaFunc(fn)
	return s
}

func (s *barStyle) TipOnComplete() BarStyleComposer {
	s.tipOnComplete = true
	return s
}

func (s *barStyle) Reverse() BarStyleComposer {
	s.rev = true
	return s
}

func (s *barStyle) Build() BarFiller {
	bf := &bFiller{
		meta:          s.metaFuncs,
		rev:           s.rev,
		tipOnComplete: s.tipOnComplete,
	}
	bf.components[iLbound] = component{
		width: runewidth.StringWidth(s.style[iLbound]),
		bytes: []byte(s.style[iLbound]),
	}
	bf.components[iRbound] = component{
		width: runewidth.StringWidth(s.style[iRbound]),
		bytes: []byte(s.style[iRbound]),
	}
	bf.components[iFiller] = component{
		width: runewidth.StringWidth(s.style[iFiller]),
		bytes: []byte(s.style[iFiller]),
	}
	bf.components[iRefiller] = component{
		width: runewidth.StringWidth(s.style[iRefiller]),
		bytes: []byte(s.style[iRefiller]),
	}
	bf.components[iPadding] = component{
		width: runewidth.StringWidth(s.style[iPadding]),
		bytes: []byte(s.style[iPadding]),
	}
	bf.tip.frames = make([]component, len(s.tipFrames))
	for i, t := range s.tipFrames {
		bf.tip.frames[i] = component{
			width: runewidth.StringWidth(t),
			bytes: []byte(t),
		}
	}
	return bf
}

func (s *bFiller) Fill(w io.Writer, stat decor.Statistics) (err error) {
	width := internal.CheckRequestedWidth(stat.RequestedWidth, stat.AvailableWidth)
	// don't count brackets as progress
	width -= (s.components[iLbound].width + s.components[iRbound].width)
	if width < 0 {
		return nil
	}

	err = s.meta[iLbound](w, s.components[iLbound].bytes)
	if err != nil {
		return err
	}

	if width == 0 {
		return s.meta[iRbound](w, s.components[iRbound].bytes)
	}

	var tip component
	var refilling, filling, padding []byte
	var fillCount, refWidth int
	curWidth := int(internal.PercentageRound(stat.Total, stat.Current, uint(width)))

	if curWidth != 0 && !stat.Completed || s.tipOnComplete {
		tip = s.tip.frames[s.tip.count%uint(len(s.tip.frames))]
		s.tip.count++
		fillCount += tip.width
	}

	if stat.Refill != 0 {
		refWidth = int(internal.PercentageRound(stat.Total, stat.Refill, uint(width)))
		curWidth -= refWidth
		refWidth += curWidth
	}

	for curWidth-fillCount >= s.components[iFiller].width {
		filling = append(filling, s.components[iFiller].bytes...)
		fillCount += s.components[iFiller].width
	}

	for refWidth-fillCount >= s.components[iRefiller].width {
		refilling = append(refilling, s.components[iRefiller].bytes...)
		fillCount += s.components[iRefiller].width
	}

	for width-fillCount >= s.components[iPadding].width {
		padding = append(padding, s.components[iPadding].bytes...)
		fillCount += s.components[iPadding].width
	}

	if width-fillCount != 0 {
		padding = append(padding, "…"...)
	}

	err = flush(w, s.rev,
		flushSection{s.meta[iRefiller], refilling},
		flushSection{s.meta[iFiller], filling},
		flushSection{s.meta[iTip], tip.bytes},
		flushSection{s.meta[iPadding], padding},
	)
	if err != nil {
		return err
	}
	return s.meta[iRbound](w, s.components[iRbound].bytes)
}

type flushSection struct {
	meta  func(io.Writer, ...interface{}) error
	bytes []byte
}

func flush(w io.Writer, rev bool, sections ...flushSection) error {
	if rev {
		for i := len(sections) - 1; i >= 0; i-- {
			s := sections[i]
			err := s.meta(w, s.bytes)
			if err != nil {
				return err
			}
		}
	} else {
		for _, s := range sections {
			err := s.meta(w, s.bytes)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeMetaFunc(fn func(...interface{}) string) func(io.Writer, ...interface{}) error {
	return func(w io.Writer, a ...interface{}) (err error) {
		for i := 0; i < len(a) && err == nil; i++ {
			_, err = io.WriteString(w, fn(string(a[i].([]byte))))
		}
		return err
	}
}
