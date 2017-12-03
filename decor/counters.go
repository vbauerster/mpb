package decor

import (
	"fmt"
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
	// Unit_KiB Kibibyte = 1024 b
	Unit_KiB
	// Unit_kB Kilobyte = 1000 b
	Unit_kB
)

type Unit uint

type CounterKiB int64

func (c CounterKiB) Format(f fmt.State, r rune) {
	prec, ok := f.Precision()

	if r == 'd' || !ok {
		prec = 0
	}
	if r == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if r == 's' {
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

	if f.Flag(int(' ')) {
		res += " "
	}
	res += unit

	if w, ok := f.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if f.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	f.Write([]byte(res))
}

type CounterKB int64

func (c CounterKB) Format(f fmt.State, r rune) {
	prec, ok := f.Precision()

	if r == 'd' || !ok {
		prec = 0
	}
	if r == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if r == 's' {
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

	if f.Flag(int(' ')) {
		res += " "
	}
	res += unit

	if w, ok := f.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if f.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	f.Write([]byte(res))
}
