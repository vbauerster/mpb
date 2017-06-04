package mpb

type BarOption func(*state)

func AppendDecorators(appenders ...DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.appendFuncs = append(bs.appendFuncs, appenders...)
	}
}

func PrependDecorators(prependers ...DecoratorFunc) BarOption {
	return func(bs *state) {
		bs.prependFuncs = append(bs.prependFuncs, prependers...)
	}
}

func BarTrimLeft(bs *state) {
	bs.trimLeftSpace = true
}

func BarTrimRight(bs *state) {
	bs.trimRightSpace = true
}

func BarTrim(bs *state) {
	bs.trimLeftSpace = true
	bs.trimRightSpace = true
}

func BarID(id int) BarOption {
	return func(bs *state) {
		bs.id = id
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
