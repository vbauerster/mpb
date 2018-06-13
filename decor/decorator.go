package decor

import (
	"fmt"
	"unicode/utf8"
)

const (
	// DidentRight bit specifies identation direction.
	// |foo   |b     | With DidentRight
	// |   foo|     b| Without DidentRight
	DidentRight = 1 << iota

	// DextraSpace bit adds extra space, makes sense with DSyncWidth only.
	// When DidentRight bit set, the space will be added to the right,
	// otherwise to the left.
	DextraSpace

	// DSyncWidth bit enables same column width synchronization.
	// Effective with multiple bars only.
	DSyncWidth

	// DSyncWidthR is shortcut for DSyncWidth|DidentRight
	DSyncWidthR = DSyncWidth | DidentRight

	// DSyncSpace is shortcut for DSyncWidth|DextraSpace
	DSyncSpace = DSyncWidth | DextraSpace

	// DSyncSpaceR is shortcut for DSyncWidth|DextraSpace|DidentRight
	DSyncSpaceR = DSyncWidth | DextraSpace | DidentRight
)

const (
	ET_STYLE_GO = iota
	ET_STYLE_HHMMSS
	ET_STYLE_HHMM
	ET_STYLE_MMSS
)

// Statistics is a struct, which gets passed to a Decorator.
type Statistics struct {
	ID        int
	Completed bool
	Total     int64
	Current   int64
}

// Decorator is an interface with one method:
//
//	Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string
//
// All decorators in this package implement this interface.
type Decorator interface {
	Decor(*Statistics, chan<- int, <-chan int) string
}

// OnCompleteMessenger is an interface with one method:
//
//	OnCompleteMessage(message string, wc ...WC)
//
// Decorators implementing this interface suppose to return provided string on complete event.
type OnCompleteMessenger interface {
	OnCompleteMessage(string, ...WC)
}

type AmountReceiver interface {
	NextAmount(int)
}

type ShutdownListener interface {
	Shutdown()
}

// DecoratorFunc is an adapter for Decorator interface
type DecoratorFunc func(*Statistics, chan<- int, <-chan int) string

func (f DecoratorFunc) Decor(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	return f(s, widthAccumulator, widthDistributor)
}

// Global convenience shortcuts
var (
	WCSyncWidth  = WC{C: DSyncWidth}
	WCSyncWidthR = WC{C: DSyncWidthR}
	WCSyncSpace  = WC{C: DSyncSpace}
	WCSyncSpaceR = WC{C: DSyncSpaceR}
)

// WC is a struct with two public fields W and C, both of int type.
// W represents width and C represents bit set of width related config.
type WC struct {
	W      int
	C      int
	format string
}

// FormatMsg formats final message according to WC.W and WC.C.
// Should be called by any Decorator implementation.
func (wc WC) FormatMsg(msg string, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if (wc.C & DSyncWidth) != 0 {
		widthAccumulator <- utf8.RuneCountInString(msg)
		max := <-widthDistributor
		if max == 0 {
			max = wc.W
		}
		if (wc.C & DextraSpace) != 0 {
			max++
		}
		return fmt.Sprintf(fmt.Sprintf(wc.format, max), msg)
	}
	return fmt.Sprintf(fmt.Sprintf(wc.format, wc.W), msg)
}

// BuildFormat builds initial format according to WC.C
func (wc *WC) BuildFormat() {
	wc.format = "%%"
	if (wc.C & DidentRight) != 0 {
		wc.format += "-"
	}
	wc.format += "%ds"
}

// OnComplete returns decorator, which wraps provided decorator, with sole
// purpose to display provided message on complete event.
//
//	`decorator` Decorator to wrap
//
//	`message` message to display on complete event
//
//	`wcc` optional WC config
func OnComplete(decorator Decorator, message string, wcc ...WC) Decorator {
	if cm, ok := decorator.(OnCompleteMessenger); ok {
		cm.OnCompleteMessage(message, wcc...)
		return decorator
	}
	msgDecorator := Name(message, wcc...)
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		if s.Completed {
			return msgDecorator.Decor(s, widthAccumulator, widthDistributor)
		}
		return decorator.Decor(s, widthAccumulator, widthDistributor)
	})
}
