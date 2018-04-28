package decor

import (
	"fmt"
	"math"
	"time"
	"unicode/utf8"
)

const (
	// DidentRight bit specifies identation direction.
	// |foo   |b     | With DidentRight
	// |   foo|     b| Without DidentRight
	DidentRight = 1 << iota

	// DwidthSync bit enables same column width synchronization.
	// Effective on multiple bars only.
	DwidthSync

	// DextraSpace bit adds extra space, makes sense with DwidthSync only.
	// When DidentRight bit set, the space will be added to the right,
	// otherwise to the left.
	DextraSpace

	// DSyncSpace is shortcut for DwidthSync|DextraSpace
	DSyncSpace = DwidthSync | DextraSpace

	// DSyncSpaceR is shortcut for DwidthSync|DextraSpace|DidentRight
	DSyncSpaceR = DwidthSync | DextraSpace | DidentRight
)

// Statistics represents statistics of the progress bar.
// Cantains: Total, Current, TimeElapsed, TimePerItemEstimate
type Statistics struct {
	ID                  int
	Completed           bool
	Total               int64
	Current             int64
	StartTime           time.Time
	TimeElapsed         time.Duration
	TimePerItemEstimate time.Duration
}

// Eta returns exponential-weighted-moving-average ETA estimator
func (s *Statistics) Eta() time.Duration {
	return time.Duration(s.Total-s.Current) * s.TimePerItemEstimate
}

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string

// OnComplete returns decorator, which wraps provided `fn` decorator, with sole
// purpose to display final on complete message.
//
//	`fn` DecoratorFunc to wrap
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func OnComplete(fn DecoratorFunc, message string, width, conf int) DecoratorFunc {
	msgDecorator := StaticName(message, width, conf)
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		if s.Completed {
			return msgDecorator(s, widthAccumulator, widthDistributor)
		}
		return fn(s, widthAccumulator, widthDistributor)
	}
}

// StaticName returns static name/message decorator.
//
//	`name` string to display
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func StaticName(name string, width, conf int) DecoratorFunc {
	nameFn := func(*Statistics) string {
		return name
	}
	return DynamicName(nameFn, width, conf)
}

// DynamicName returns dynamic name/message decorator.
//
//	`messageFn` callback function to get dynamic string message
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func DynamicName(messageFn func(*Statistics) string, width, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		name := messageFn(s)
		if (conf & DwidthSync) != 0 {
			widthAccumulator <- utf8.RuneCountInString(name)
			max := <-widthDistributor
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), name)
		}
		return fmt.Sprintf(fmt.Sprintf(format, width), name)
	}
}

// CountersNoUnit returns raw counters decorator
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func CountersNoUnit(pairFormat string, width, conf int) DecoratorFunc {
	return counters(pairFormat, 0, width, conf)
}

// CountersKibiByte returns human friendly byte counters decorator, where counters unit is multiple by 1024.
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
//
// pairFormat example:
//
//	"%.1f / %.1f" = "1.0MiB / 12.0MiB" or "% .1f / % .1f" = "1.0 MiB / 12.0 MiB"
func CountersKibiByte(pairFormat string, width, conf int) DecoratorFunc {
	return counters(pairFormat, unitKiB, width, conf)
}

// CountersKiloByte returns human friendly byte counters decorator, where counters unit is multiple by 1000.
//
//	`pairFormat` printf compatible verbs for current and total, like "%f" or "%d"
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
//
// pairFormat example:
//
//	"%.1f / %.1f" = "1.0MB / 12.0MB" or "% .1f / % .1f" = "1.0 MB / 12.0 MB"
func CountersKiloByte(pairFormat string, width, conf int) DecoratorFunc {
	return counters(pairFormat, unitKB, width, conf)
}

func counters(pairFormat string, unit counterUnit, width, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		switch unit {
		case unitKiB:
			str = fmt.Sprintf(pairFormat, CounterKiB(s.Current), CounterKiB(s.Total))
		case unitKB:
			str = fmt.Sprintf(pairFormat, CounterKB(s.Current), CounterKB(s.Total))
		default:
			str = fmt.Sprintf(pairFormat, s.Current, s.Total)
		}
		if (conf & DwidthSync) != 0 {
			widthAccumulator <- utf8.RuneCountInString(str)
			max := <-widthDistributor
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, width), str)
	}
}

// ETA returns exponential-weighted-moving-average ETA decorator.
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func ETA(width, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		str := fmt.Sprint(time.Duration(s.Eta().Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			widthAccumulator <- utf8.RuneCountInString(str)
			max := <-widthDistributor
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, width), str)
	}
}

// Elapsed returns elapsed time decorator.
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func Elapsed(width, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		str := fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			widthAccumulator <- utf8.RuneCountInString(str)
			max := <-widthDistributor
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, width), str)
	}
}

// Percentage returns percentage decorator.
//
//	`width` width reservation to apply, ignored if `DwidthSync` bit is set
//
//	`conf` bit set config, [DidentRight|DwidthSync|DextraSpace]
func Percentage(width, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		str := fmt.Sprintf("%d %%", CalcPercentage(s.Total, s.Current, 100))
		if (conf & DwidthSync) != 0 {
			widthAccumulator <- utf8.RuneCountInString(str)
			max := <-widthDistributor
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, width), str)
	}
}

// CalcPercentage is a helper function, to calculate percentage.
func CalcPercentage(total, current, width int64) (perc int64) {
	if total <= 0 {
		return 0
	}
	if current > total {
		current = total
	}

	num := float64(width) * float64(current) / float64(total)
	ceil := math.Ceil(num)
	diff := ceil - num
	// num = 2.34 will return 2
	// num = 2.44 will return 3
	if math.Max(diff, 0.6) == diff {
		return int64(num)
	}
	return int64(ceil)
}
