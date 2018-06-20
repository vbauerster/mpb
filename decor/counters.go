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
	prec, ok := st.Precision()

	if verb == 'd' || !ok {
		prec = 0
	}
	if verb == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if verb == 's' {
		prec = 1
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
	prec, ok := st.Precision()

	if verb == 'd' || !ok {
		prec = 0
	}
	if verb == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if verb == 's' {
		prec = 1
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
func CountersNoUnit(pairFormat string, wcc ...WC) Decorator {
	return Counters(0, pairFormat, wcc...)
}

// CountersKibiByte is a wrapper around Counters with predefined unit UnitKiB (bytes/1024).
func CountersKibiByte(pairFormat string, wcc ...WC) Decorator {
	return Counters(UnitKiB, pairFormat, wcc...)
}

// CountersKiloByte is a wrapper around Counters with predefined unit UnitKB (bytes/1000).
func CountersKiloByte(pairFormat string, wcc ...WC) Decorator {
	return Counters(UnitKB, pairFormat, wcc...)
}

// Counters decorator with dynamic unit measure adjustment.
//
//	`unit` one of [0|UnitKiB|UnitKB] zero for no unit
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`wcc` optional WC config
//
// pairFormat example if UnitKB is chosen:
//
//	"%.1f / %.1f" = "1.0MB / 12.0MB" or "% .1f / % .1f" = "1.0 MB / 12.0 MB"
func Counters(unit int, pairFormat string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		switch unit {
		case UnitKiB:
			str = fmt.Sprintf(pairFormat, CounterKiB(s.Current), CounterKiB(s.Total))
		case UnitKB:
			str = fmt.Sprintf(pairFormat, CounterKB(s.Current), CounterKB(s.Total))
		default:
			str = fmt.Sprintf(pairFormat, s.Current, s.Total)
		}
		return wc.FormatMsg(str, widthAccumulator, widthDistributor)
	})
}
