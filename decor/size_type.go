package decor

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

const (
	_ib   SizeB1024 = iota
	_iKiB SizeB1024 = 1 << (iota * 10)
	_iMiB
	_iGiB
	_iTiB
)

//go:generate stringer -type=SizeB1024 -trimprefix=_i
type SizeB1024 int64

func (self SizeB1024) Format(st fmt.State, verb rune) {
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
	var unit SizeB1024
	switch {
	case self < _iKiB:
		unit = _ib
		b.WriteString(strconv.FormatFloat(float64(self), 'f', prec, 64))
	case self < _iMiB:
		unit = _iKiB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_iKiB), 'f', prec, 64))
	case self < _iGiB:
		unit = _iMiB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_iMiB), 'f', prec, 64))
	case self < _iTiB:
		unit = _iGiB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_iGiB), 'f', prec, 64))
	case self <= math.MaxInt64:
		unit = _iTiB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_iTiB), 'f', prec, 64))
	}

	if st.Flag(' ') {
		b.WriteString(" ")
	}
	b.WriteString(unit.String())

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

const (
	_b  SizeB1000 = 0
	_KB SizeB1000 = 1000
	_MB SizeB1000 = _KB * 1000
	_GB SizeB1000 = _MB * 1000
	_TB SizeB1000 = _GB * 1000
)

//go:generate stringer -type=SizeB1000 -trimprefix=_
type SizeB1000 int64

func (self SizeB1000) Format(st fmt.State, verb rune) {
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
	var unit SizeB1000
	switch {
	case self < _KB:
		unit = _b
		b.WriteString(strconv.FormatFloat(float64(self), 'f', prec, 64))
	case self < _MB:
		unit = _KB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_KB), 'f', prec, 64))
	case self < _GB:
		unit = _MB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_MB), 'f', prec, 64))
	case self < _TB:
		unit = _GB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_GB), 'f', prec, 64))
	case self <= math.MaxInt64:
		unit = _TB
		b.WriteString(strconv.FormatFloat(float64(self)/float64(_TB), 'f', prec, 64))
	}

	if st.Flag(' ') {
		b.WriteString(" ")
	}
	b.WriteString(unit.String())

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
