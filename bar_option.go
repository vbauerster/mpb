package mpb

import "github.com/vbauerster/mpb/decor"

// BarOption is a function option which changes the default behavior of a bar,
// if passed to p.AddBar(int64, ...BarOption)
type BarOption func(*state)

func AppendDecorators(appenders ...decor.DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.appendFuncs = append(bs.appendFuncs, appenders...)
	}
}

func PrependDecorators(prependers ...decor.DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.prependFuncs = append(bs.prependFuncs, prependers...)
	}
}

func BarTrimLeft() BarOption {
	return func(bs *state) {
		bs.trimLeftSpace = true
	}
}

func BarTrimRight() BarOption {
	return func(bs *state) {
		bs.trimRightSpace = true
	}
}

func BarTrim() BarOption {
	return func(bs *state) {
		bs.trimLeftSpace = true
		bs.trimRightSpace = true
	}
}

func BarID(id int) BarOption {
	return func(bs *state) {
		bs.id = id
	}
}

// BarEtaAlpha option is a way to adjust ETA behavior.
// You can play with it, if you're not satisfied with default behavior.
// Default value is 0.25.
func BarEtaAlpha(a float64) BarOption {
	return func(bs *state) {
		bs.etaAlpha = a
	}
}

func barWidth(w int) BarOption {
	return func(bs *state) {
		bs.width = w
	}
}

func barFormat(format string) BarOption {
	return func(bs *state) {
		bs.updateFormat(format)
	}
}
