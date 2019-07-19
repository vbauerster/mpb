package decor

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/vbauerster/mpb/v4/internal"
)

type PercentageType float64

func (s PercentageType) Format(st fmt.State, verb rune) {
	var prec int
	switch verb {
	case 'd':
	case 's':
		prec = -1
	default:
		if p, ok := st.Precision(); ok {
			prec = p
		} else {
			prec = 6
		}
	}

	var b strings.Builder
	b.WriteString(strconv.FormatFloat(float64(s), 'f', prec, 64))

	if st.Flag(' ') {
		b.WriteString(" ")
	}
	b.WriteString("%")

	if w, ok := st.Width(); ok {
		if l := b.Len(); l < w {
			pad := strings.Repeat(" ", w-l)
			if st.Flag('-') {
				b.WriteString(pad)
			} else {
				tmp := b.String()
				b.Reset()
				b.WriteString(pad)
				b.WriteString(tmp)
			}
		}
	}

	io.WriteString(st, b.String())
}

// Percentage returns percentage decorator. It's a wrapper of NewPercentage.
func Percentage(wcc ...WC) Decorator {
	return NewPercentage("% d", wcc...)
}

// NewPercentage percentage decorator with custom fmt string.
//
// fmt examples:
//
//	fmt="%.1f"  output: "1.0%"
//	fmt="% .1f" output: "1.0 %"
//	fmt="%d"    output: "1%"
//	fmt="% d"   output: "1 %"
//
func NewPercentage(fmt string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	if fmt == "" {
		fmt = "% d"
	}
	d := &percentageDecorator{
		WC:  wc,
		fmt: fmt,
	}
	return d
}

type percentageDecorator struct {
	WC
	fmt         string
	completeMsg *string
}

func (d *percentageDecorator) Decor(st *Statistics) string {
	if st.Completed && d.completeMsg != nil {
		return d.FormatMsg(*d.completeMsg)
	}
	p := internal.Percentage(st.Total, st.Current, 100)
	return d.FormatMsg(fmt.Sprintf(d.fmt, PercentageType(p)))
}

func (d *percentageDecorator) OnCompleteMessage(msg string) {
	d.completeMsg = &msg
}
