package decor

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	_   = iota
	KiB = 1 << (iota * 10)
	MiB
	GiB
	TiB
)

const (
	KB = 1000
	MB = KB * 1000
	GB = MB * 1000
	TB = GB * 1000
)

const (
	_ = iota
	UnitKiB
	UnitKB
)

type CounterKiB int64

func (c CounterKiB) Format(st fmt.State, verb rune) {
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

	var res, unit string
	switch {
	case c >= TiB:
		unit = "TiB"
		res = strconv.FormatFloat(float64(c)/TiB, 'f', prec, 64)
	case c >= GiB:
		unit = "GiB"
		res = strconv.FormatFloat(float64(c)/GiB, 'f', prec, 64)
	case c >= MiB:
		unit = "MiB"
		res = strconv.FormatFloat(float64(c)/MiB, 'f', prec, 64)
	case c >= KiB:
		unit = "KiB"
		res = strconv.FormatFloat(float64(c)/KiB, 'f', prec, 64)
	default:
		unit = "b"
		res = strconv.FormatInt(int64(c), 10)
	}

	if st.Flag(' ') {
		res += " "
	}
	res += unit

	if w, ok := st.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if st.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	io.WriteString(st, res)
}

type CounterKB int64

func (c CounterKB) Format(st fmt.State, verb rune) {
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

	var res, unit string
	switch {
	case c >= TB:
		unit = "TB"
		res = strconv.FormatFloat(float64(c)/TB, 'f', prec, 64)
	case c >= GB:
		unit = "GB"
		res = strconv.FormatFloat(float64(c)/GB, 'f', prec, 64)
	case c >= MB:
		unit = "MB"
		res = strconv.FormatFloat(float64(c)/MB, 'f', prec, 64)
	case c >= KB:
		unit = "kB"
		res = strconv.FormatFloat(float64(c)/KB, 'f', prec, 64)
	default:
		unit = "b"
		res = strconv.FormatInt(int64(c), 10)
	}

	if st.Flag(' ') {
		res += " "
	}
	res += unit

	if w, ok := st.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if st.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	io.WriteString(st, res)
}

// CountersNoUnit is a wrapper around Counters with no unit param.
func CountersNoUnit(pairFmt string, wcc ...WC) Decorator {
	return Counters(0, pairFmt, wcc...)
}

// CountersKibiByte is a wrapper around Counters with predefined unit
// UnitKiB (bytes/1024).
func CountersKibiByte(pairFmt string, wcc ...WC) Decorator {
	return Counters(UnitKiB, pairFmt, wcc...)
}

// CountersKiloByte is a wrapper around Counters with predefined unit
// UnitKB (bytes/1000).
func CountersKiloByte(pairFmt string, wcc ...WC) Decorator {
	return Counters(UnitKB, pairFmt, wcc...)
}

// Counters decorator with dynamic unit measure adjustment.
//
//	`unit` one of [0|UnitKiB|UnitKB] zero for no unit
//
//	`pairFmt` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`wcc` optional WC config
//
// pairFmt example if UnitKB is chosen:
//
//	"%.1f / %.1f" = "1.0MB / 12.0MB" or "% .1f / % .1f" = "1.0 MB / 12.0 MB"
//
func Counters(unit int, pairFmt string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	d := &countersDecorator{
		WC:      wc,
		unit:    unit,
		pairFmt: pairFmt,
	}
	return d
}

type countersDecorator struct {
	WC
	unit        int
	pairFmt     string
	completeMsg *string
}

func (d *countersDecorator) Decor(st *Statistics) string {
	if st.Completed && d.completeMsg != nil {
		return d.FormatMsg(*d.completeMsg)
	}

	var str string
	switch d.unit {
	case UnitKiB:
		str = fmt.Sprintf(d.pairFmt, CounterKiB(st.Current), CounterKiB(st.Total))
	case UnitKB:
		str = fmt.Sprintf(d.pairFmt, CounterKB(st.Current), CounterKB(st.Total))
	default:
		str = fmt.Sprintf(d.pairFmt, st.Current, st.Total)
	}

	return d.FormatMsg(str)
}

func (d *countersDecorator) OnCompleteMessage(msg string) {
	d.completeMsg = &msg
}
