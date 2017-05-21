package mpb

import "fmt"

const (
	_          = iota
	bytesInKiB = 1 << (iota * 10)
	bytesInMiB
	bytesInGiB
	bytesInTiB
)

type Units uint

const (
	_ = iota
	UnitBytes
)

func Format(i int64) *formatter {
	return &formatter{n: i}
}

type formatter struct {
	n     int64
	unit  Units
	width int
}

func (f *formatter) To(unit Units) *formatter {
	f.unit = unit
	return f
}

func (f *formatter) Width(width int) *formatter {
	f.width = width
	return f
}

func (f *formatter) String() string {
	switch f.unit {
	case UnitBytes:
		return formatBytes(f.n)
	default:
		return fmt.Sprintf(fmt.Sprintf("%%%dd", f.width), f.n)
	}
}

func formatBytes(i int64) (result string) {
	switch {
	case i >= bytesInTiB:
		result = fmt.Sprintf("%.1fTiB", float64(i)/bytesInTiB)
	case i >= bytesInGiB:
		result = fmt.Sprintf("%.1fGiB", float64(i)/bytesInGiB)
	case i >= bytesInMiB:
		result = fmt.Sprintf("%.1fMiB", float64(i)/bytesInMiB)
	case i >= bytesInKiB:
		result = fmt.Sprintf("%.1fKiB", float64(i)/bytesInKiB)
	default:
		result = fmt.Sprintf("%db", i)
	}
	return
}
