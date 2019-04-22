package mpb

import (
	"github.com/vbauerster/mpb/v4/decor"
)

// BarOption is a function option which changes the default behavior of a bar.
type BarOption func(*bState)

// AppendDecorators let you inject decorators to the bar's right side.
func AppendDecorators(appenders ...decor.Decorator) BarOption {
	return func(s *bState) {
		for _, decorator := range appenders {
			s.appendAmountReceiver(decorator)
			s.appendShutdownListener(decorator)
			if md, ok := decorator.(*decor.MergeDecorator); ok {
				s.appendAmountReceiver(md.Decorator)
				s.appendShutdownListener(md.Decorator)
				s.aDecorators = append(s.aDecorators, md.PlaceHolders()...)
			}
			s.aDecorators = append(s.aDecorators, decorator)
		}
	}
}

// PrependDecorators let you inject decorators to the bar's left side.
func PrependDecorators(prependers ...decor.Decorator) BarOption {
	return func(s *bState) {
		for _, decorator := range prependers {
			s.appendAmountReceiver(decorator)
			s.appendShutdownListener(decorator)
			if md, ok := decorator.(*decor.MergeDecorator); ok {
				s.appendAmountReceiver(md.Decorator)
				s.appendShutdownListener(md.Decorator)
				s.pDecorators = append(s.pDecorators, md.PlaceHolders()...)
			}
			s.pDecorators = append(s.pDecorators, decorator)
		}
	}
}

// BarID sets bar id.
func BarID(id int) BarOption {
	return func(s *bState) {
		s.id = id
	}
}

// BarWidth sets bar width independent of the container.
func BarWidth(width int) BarOption {
	return func(s *bState) {
		s.width = width
	}
}

// BarRemoveOnComplete removes whole bar line on complete event. Any
// decorators attached to the bar, having OnComplete action will not
// have a chance to run its OnComplete action, because of this option.
func BarRemoveOnComplete() BarOption {
	return func(s *bState) {
		s.dropOnComplete = true
	}
}

// BarReplaceOnComplete is deprecated. Refer to BarParkTo option.
func BarReplaceOnComplete(runningBar *Bar) BarOption {
	return BarParkTo(runningBar)
}

// BarParkTo parks constructed bar into the runningBar. In other words,
// constructed bar will start only after runningBar has been completed.
// Parked bar inherits priority of the runningBar, if no BarPriority
// option is set. Parked bar will replace runningBar eventually.
func BarParkTo(runningBar *Bar) BarOption {
	if runningBar == nil {
		return nil
	}
	runningBar.dropOnComplete()
	return func(s *bState) {
		s.runningBar = runningBar
	}
}

// BarClearOnComplete clears bar part of bar line on complete event.
func BarClearOnComplete() BarOption {
	return func(s *bState) {
		s.noBufBOnComplete = true
	}
}

// BarPriority sets bar's priority. Zero is highest priority, i.e. bar
// will be on top. If `BarReplaceOnComplete` option is supplied, this
// option is ignored.
func BarPriority(priority int) BarOption {
	return func(s *bState) {
		s.priority = priority
	}
}

// BarExtender is an option to extend bar to the next new line, with
// arbitrary output.
func BarExtender(extender Filler) BarOption {
	return func(s *bState) {
		s.extender = extender
	}
}

// TrimSpace trims bar's edge spaces.
func TrimSpace() BarOption {
	return func(s *bState) {
		s.trimSpace = true
	}
}

// BarStyle sets custom bar style, default one is "[=>-]<+".
//
//	'[' left bracket rune
//
//	'=' fill rune
//
//	'>' tip rune
//
//	'-' empty rune
//
//	']' right bracket rune
//
//	'<' reverse tip rune, used when BarReverse option is set
//
//	'+' refill rune, used when *Bar.SetRefill(int64) is called
//
// It's ok to provide first five runes only, for example BarStyle("╢▌▌░╟").
// To omit left and right bracket runes, either set style as " =>- "
// or use BarNoBrackets option.
func BarStyle(style string) BarOption {
	chk := func(filler Filler) (interface{}, bool) {
		if style == "" {
			return nil, false
		}
		t, ok := filler.(*barFiller)
		return t, ok
	}
	cb := func(t interface{}) {
		t.(*barFiller).setStyle(style)
	}
	return MakeFillerTypeSpecificBarOption(chk, cb)
}

// BarNoBrackets omits left and right edge runes of the bar. Edges are
// brackets in default bar style, hence the name of the option.
func BarNoBrackets() BarOption {
	chk := func(filler Filler) (interface{}, bool) {
		t, ok := filler.(*barFiller)
		return t, ok
	}
	cb := func(t interface{}) {
		t.(*barFiller).noBrackets = true
	}
	return MakeFillerTypeSpecificBarOption(chk, cb)
}

// BarReverse reverse mode, bar will progress from right to left.
func BarReverse() BarOption {
	chk := func(filler Filler) (interface{}, bool) {
		t, ok := filler.(*barFiller)
		return t, ok
	}
	cb := func(t interface{}) {
		t.(*barFiller).reverse = true
	}
	return MakeFillerTypeSpecificBarOption(chk, cb)
}

// SpinnerStyle sets custom spinner style.
// Effective when Filler type is spinner.
func SpinnerStyle(frames []string) BarOption {
	chk := func(filler Filler) (interface{}, bool) {
		if len(frames) == 0 {
			return nil, false
		}
		t, ok := filler.(*spinnerFiller)
		return t, ok
	}
	cb := func(t interface{}) {
		t.(*spinnerFiller).frames = frames
	}
	return MakeFillerTypeSpecificBarOption(chk, cb)
}

// MakeFillerTypeSpecificBarOption makes BarOption specific to Filler's
// actual type. If you implement your own Filler, so most probably
// you'll need this. See BarStyle or SpinnerStyle for example.
func MakeFillerTypeSpecificBarOption(
	typeChecker func(Filler) (interface{}, bool),
	cb func(interface{}),
) BarOption {
	return func(s *bState) {
		if t, ok := typeChecker(s.filler); ok {
			cb(t)
		}
	}
}

// BarOptOnCond returns option when condition evaluates to true.
func BarOptOnCond(option BarOption, condition func() bool) BarOption {
	if condition() {
		return option
	}
	return nil
}
