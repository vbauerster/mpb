package mpb

import "github.com/vbauerster/mpb/decor"

// BarOption is a function option which changes the default behavior of a bar,
// if passed to p.AddBar(int64, ...BarOption)
type BarOption func(*state)

// AppendDecorators let you inject decorators to the bar's right side
func AppendDecorators(appenders ...decor.DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.appendFuncs = append(bs.appendFuncs, appenders...)
	}
}

// PrependDecorators let you inject decorators to the bar's left side
func PrependDecorators(prependers ...decor.DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.prependFuncs = append(bs.prependFuncs, prependers...)
	}
}

// BarTrimLeft trims left side space of the bar
func BarTrimLeft() BarOption {
	return func(bs *state) {
		bs.trimLeftSpace = true
	}
}

// BarTrimRight trims right space of the bar
func BarTrimRight() BarOption {
	return func(bs *state) {
		bs.trimRightSpace = true
	}
}

// BarTirm trims both left and right spaces of the bar
func BarTrim() BarOption {
	return func(bs *state) {
		bs.trimLeftSpace = true
		bs.trimRightSpace = true
	}
}

// BarID overwrites internal bar id
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

// BarDropRatio sets drop ratio, default is 10. Effective when total is dynamic.
// If progress tip reaches total, but total is not final value yet, tip will be
// dropped by specified ratio.
func BarDropRatio(ratio int64) BarOption {
	return func(bs *state) {
		bs.dropRatio = ratio
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
