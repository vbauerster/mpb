package decor

import "fmt"

const (
	_          = iota
	bytesInKiB = 1 << (iota * 10)
	bytesInMiB
	bytesInGiB
	bytesInTiB
)

const (
	bytesInKb = 1000
	bytesInMB = bytesInKb * 1000
	bytesInGB = bytesInMB * 1000
	bytesInTB = bytesInGB * 1000
)

const (
	// Kibibyte = 1024 b
	Unit_KiB = iota
	// Kilobyte = 1000 b
	Unit_kB
)

type Units uint

func Format(i int) *formatter {
	return &formatter{n: i}
}

type formatter struct {
	n     int
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

func formatKiB(i int) (result string) {
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

func formatKB(i int) (result string) {
	switch {
	case i >= bytesInTB:
		result = fmt.Sprintf("%.1fTB", float64(i)/bytesInTB)
	case i >= bytesInGB:
		result = fmt.Sprintf("%.1fGB", float64(i)/bytesInGB)
	case i >= bytesInMB:
		result = fmt.Sprintf("%.1fMB", float64(i)/bytesInMB)
	case i >= bytesInKb:
		result = fmt.Sprintf("%.1fkB", float64(i)/bytesInKb)
	default:
		result = fmt.Sprintf("%db", i)
	}
	return
}
