package decor

import (
	"fmt"
	"math"
	"time"
	"unicode/utf8"

	"github.com/VividCortex/ewma"
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

// Statistics is a struct, which Decorator interface depends upon.
type Statistics struct {
	ID          int
	Completed   bool
	Total       int64
	Current     int64
	StartTime   time.Time
	TimeElapsed time.Duration
}

// Decorator is an interface with one method:
//
//	Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string
//
// All decorators in this package implement this interface.
type Decorator interface {
	Decor(*Statistics, chan<- int, <-chan int) string
}

// CompleteMessenger is an interface with one method:
//
//	OnComplete(message string, wc ...WC)
//
// Decorators implementing this interface suppose to return provided string on complete event.
type CompleteMessenger interface {
	OnComplete(string, ...WC)
}

// DecoratorFunc is an adapter for Decorator interface
type DecoratorFunc func(*Statistics, chan<- int, <-chan int) string

func (f DecoratorFunc) Decor(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	return f(s, widthAccumulator, widthDistributor)
}

// WC is a struct with two public fields W and C, both of int type.
// W represents width and C represents bit set of width related config.
type WC struct {
	W      int
	C      int
	format string
}

func (wc WC) formatMsg(msg string, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	format := wc.buildFormat()
	if (wc.C & DSyncWidth) != 0 {
		widthAccumulator <- utf8.RuneCountInString(msg)
		max := <-widthDistributor
		if max == 0 {
			max = wc.W
		}
		if (wc.C & DextraSpace) != 0 {
			max++
		}
		return fmt.Sprintf(fmt.Sprintf(format, max), msg)
	}
	return fmt.Sprintf(fmt.Sprintf(format, wc.W), msg)
}

func (wc *WC) buildFormat() string {
	if wc.format != "" {
		return wc.format
	}
	wc.format = "%%"
	if (wc.C & DidentRight) != 0 {
		wc.format += "-"
	}
	wc.format += "%ds"
	return wc.format
}

// Global convenience shortcuts
var (
	WCSyncWidth  = WC{C: DSyncWidth}
	WCSyncWidthR = WC{C: DSyncWidthR}
	WCSyncSpace  = WC{C: DSyncSpace}
	WCSyncSpaceR = WC{C: DSyncSpaceR}
)

// OnComplete returns decorator, which wraps provided decorator, with sole
// purpose to display provided message on complete event.
//
//	`decorator` Decorator to wrap
//
//	`message` message to display on complete event
//
//	`wc` optional WC config
func OnComplete(decorator Decorator, message string, wc ...WC) Decorator {
	if cm, ok := decorator.(CompleteMessenger); ok {
		cm.OnComplete(message, wc...)
		return decorator
	}
	msgDecorator := Name(message, wc...)
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		if s.Completed {
			return msgDecorator.Decor(s, widthAccumulator, widthDistributor)
		}
		return decorator.Decor(s, widthAccumulator, widthDistributor)
	})
}

// StaticName returns name decorator.
//
//	`name` string to display
//
//	`wc` optional WC config
func StaticName(name string, wc ...WC) Decorator {
	return Name(name, wc...)
}

// Name returns name decorator.
//
//	`name` string to display
//
//	`wc` optional WC config
func Name(name string, wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		return wc0.formatMsg(name, widthAccumulator, widthDistributor)
	})
}

// CountersNoUnit returns raw counters decorator
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`wc` optional WC config
func CountersNoUnit(pairFormat string, wc ...WC) Decorator {
	return counters(pairFormat, 0, wc...)
}

// CountersKibiByte returns human friendly byte counters decorator, where counters unit is multiple by 1024.
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`wc` optional WC config
//
// pairFormat example:
//
//	"%.1f / %.1f" = "1.0MiB / 12.0MiB" or "% .1f / % .1f" = "1.0 MiB / 12.0 MiB"
func CountersKibiByte(pairFormat string, wc ...WC) Decorator {
	return counters(pairFormat, unitKiB, wc...)
}

// CountersKiloByte returns human friendly byte counters decorator, where counters unit is multiple by 1000.
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`wc` optional WC config
//
// pairFormat example:
//
//	"%.1f / %.1f" = "1.0MB / 12.0MB" or "% .1f / % .1f" = "1.0 MB / 12.0 MB"
func CountersKiloByte(pairFormat string, wc ...WC) Decorator {
	return counters(pairFormat, unitKB, wc...)
}

func counters(pairFormat string, unit int, wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		switch unit {
		case unitKiB:
			str = fmt.Sprintf(pairFormat, CounterKiB(s.Current), CounterKiB(s.Total))
		case unitKB:
			str = fmt.Sprintf(pairFormat, CounterKB(s.Current), CounterKB(s.Total))
		default:
			str = fmt.Sprintf(pairFormat, s.Current, s.Total)
		}
		return wc0.formatMsg(str, widthAccumulator, widthDistributor)
	})
}

// ETA returns exponential-weighted-moving-average ETA decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`age` is a decay factor alpha for underlying ewma.
//	 General rule of thumb, for the best value:
//	 expected progress time in seconds divided by two.
//	 For example expected progress duration is one hour.
//	 age = 3600 / 2
//
//	`startBlock` is channel, user suppose to send time.Now() on each iteration of block start.
//
//	`wc` optional WC config
func ETA(style int, age float64, startBlock chan time.Time, wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	if age == .0 {
		age = ewma.AVG_METRIC_AGE
	}
	return &EwmaETA{
		MovingAverage: ewma.NewMovingAverage(age),
		StartBlockCh:  startBlock,
		style:         style,
		wc:            wc0,
	}
}

// EwmaETA is a struct, which implements ewma based ETA decorator.
// Normally should not be used directly, use helper func instead:
//
//	decor.ETA(int, float64, chan time.Time, ...decor.WC)
type EwmaETA struct {
	ewma.MovingAverage
	StartBlockCh chan time.Time
	style        int
	wc           WC
	onComplete   *struct {
		msg string
		wc  WC
	}
}

func (s *EwmaETA) Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if st.Completed && s.onComplete != nil {
		return s.onComplete.wc.formatMsg(s.onComplete.msg, widthAccumulator, widthDistributor)
	}

	var str string
	timeRemaining := time.Duration(st.Total-st.Current) * time.Duration(math.Round(s.Value()))
	hours := int64((timeRemaining / time.Hour) % 60)
	minutes := int64((timeRemaining / time.Minute) % 60)
	seconds := int64((timeRemaining / time.Second) % 60)

	switch s.style {
	case ET_STYLE_GO:
		str = fmt.Sprint(time.Duration(timeRemaining.Seconds()) * time.Second)
	case ET_STYLE_HHMMSS:
		str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case ET_STYLE_HHMM:
		str = fmt.Sprintf("%02d:%02d", hours, minutes)
	case ET_STYLE_MMSS:
		str = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return s.wc.formatMsg(str, widthAccumulator, widthDistributor)
}

func (s *EwmaETA) OnComplete(msg string, wc ...WC) {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	s.onComplete = &struct {
		msg string
		wc  WC
	}{msg, wc0}
}

// Elapsed returns elapsed time decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`wc` optional WC config
func Elapsed(style int, wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		hours := int64((s.TimeElapsed / time.Hour) % 60)
		minutes := int64((s.TimeElapsed / time.Minute) % 60)
		seconds := int64((s.TimeElapsed / time.Second) % 60)

		switch style {
		case ET_STYLE_GO:
			str = fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
		case ET_STYLE_HHMMSS:
			str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		case ET_STYLE_HHMM:
			str = fmt.Sprintf("%02d:%02d", hours, minutes)
		case ET_STYLE_MMSS:
			str = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
		return wc0.formatMsg(str, widthAccumulator, widthDistributor)
	})
}

// Percentage returns percentage decorator.
//
//	`wc` optional WC config
func Percentage(wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		str := fmt.Sprintf("%d %%", CalcPercentage(s.Total, s.Current, 100))
		return wc0.formatMsg(str, widthAccumulator, widthDistributor)
	})
}

// CalcPercentage is a helper function, to calculate percentage.
func CalcPercentage(total, current, width int64) int64 {
	if total <= 0 {
		return 0
	}
	if current > total {
		current = total
	}

	p := float64(width) * float64(current) / float64(total)
	return int64(math.Round(p))
}

// SpeedNoUnit returns raw I/O operation speed decorator.
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`wc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0" or "% .1f" = "1.0"
func SpeedNoUnit(unitFormat string, wc ...WC) Decorator {
	return speed(unitFormat, 0, wc...)
}

// SpeedKibiByte returns human friendly I/O operation speed decorator,
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`wc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0MiB/s" or "% .1f" = "1.0 MiB/s"
func SpeedKibiByte(unitFormat string, wc ...WC) Decorator {
	return speed(unitFormat, unitKiB, wc...)
}

// SpeedKiloByte returns human friendly I/O operation speed decorator,
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`wc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0MB/s" or "% .1f" = "1.0 MB/s"
func SpeedKiloByte(unitFormat string, wc ...WC) Decorator {
	return speed(unitFormat, unitKB, wc...)
}

func speed(unitFormat string, unit int, wc ...WC) Decorator {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		speed := float64(s.Current) / s.TimeElapsed.Seconds()
		if math.IsNaN(speed) || math.IsInf(speed, 0) {
			speed = .0
		}

		switch unit {
		case unitKiB:
			str = fmt.Sprintf(unitFormat, SpeedKiB(speed))
		case unitKB:
			str = fmt.Sprintf(unitFormat, SpeedKB(speed))
		default:
			str = fmt.Sprintf(unitFormat, speed)
		}
		return wc0.formatMsg(str, widthAccumulator, widthDistributor)
	})
}
