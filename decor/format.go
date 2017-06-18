package decor

import "fmt"

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

type Units uint

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
	case Unit_KiB:
		return formatKiB(f.n)
	case Unit_kB:
		return formatKB(f.n)
	default:
		return fmt.Sprintf(fmt.Sprintf("%%%dd", f.width), f.n)
	}
}

func formatKiB(i int64) (result string) {
	switch {
	case i >= TiB:
		result = fmt.Sprintf("%.1fTiB", float64(i)/TiB)
	case i >= GiB:
		result = fmt.Sprintf("%.1fGiB", float64(i)/GiB)
	case i >= MiB:
		result = fmt.Sprintf("%.1fMiB", float64(i)/MiB)
	case i >= KiB:
		result = fmt.Sprintf("%.1fKiB", float64(i)/KiB)
	default:
		result = fmt.Sprintf("%db", i)
	}
	return
}

func formatKB(i int64) (result string) {
	switch {
	case i >= TB:
		result = fmt.Sprintf("%.1fTB", float64(i)/TB)
	case i >= GB:
		result = fmt.Sprintf("%.1fGB", float64(i)/GB)
	case i >= MB:
		result = fmt.Sprintf("%.1fMB", float64(i)/MB)
	case i >= KB:
		result = fmt.Sprintf("%.1fkB", float64(i)/KB)
	default:
		result = fmt.Sprintf("%db", i)
	}
	return
}
