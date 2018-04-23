package decor

import (
	"fmt"
	"math"
	"time"
	"unicode/utf8"
)

const (
	// DidentRight specifies identation direction.
	// |foo   |b     | With DidentRight
	// |   foo|     b| Without DidentRight
	DidentRight = 1 << iota

	// DwidthSync will auto sync max width.
	// Makes sense when there're more than one bar
	DwidthSync

	// DextraSpace adds extra space, makes sense with DwidthSync only.
	// When DidentRight bit set, the space will be added to the right,
	// otherwise to the left.
	DextraSpace

	// DSyncSpace is shortcut for DwidthSync|DextraSpace
	DSyncSpace = DwidthSync | DextraSpace
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

// OnComplete wraps provided decorator `fn` with on complete event `message`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func OnComplete(fn DecoratorFunc, message string, minWidth, conf int) DecoratorFunc {
	msgDecorator := StaticName(message, minWidth, conf)
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		if s.Completed {
			return msgDecorator(s, widthAccumulator, widthDistributor)
		}
		return fn(s, widthAccumulator, widthDistributor)
	}
}

// StaticName is a simple name/message decorator.
// If you set `DwidthSync` bit in `conf` param, `minWidth` is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func StaticName(name string, minWidth, conf int) DecoratorFunc {
	nameFn := func(*Statistics) string {
		return name
	}
	return DynamicName(nameFn, minWidth, conf)
}

// DynamicName is a name/message decorator, with ability to change message via provided `messageFn`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func DynamicName(messageFn func(*Statistics) string, minWidth, conf int) DecoratorFunc {
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
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), name)
	}
}

// CountersNoUnit returns raw counters decorator
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func CountersNoUnit(pairFormat string, minWidth, conf int) DecoratorFunc {
	return Counters(pairFormat, 0, minWidth, conf)
}

// CountersKibiByte returns human friendly byte counters decorator,
// where counters unit is multiple by 1024.
// `pairFormat` must contain two printf compatible verbs, like "%f" or "%d".
// First verb substituted with Current, second one with Total.
// Example: `"%.1f / %.1f" = "1.0MiB / 12.0MiB"` or `"% .1f / % .1f" = "1.0 MiB / 12.0 MiB"`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func CountersKibiByte(pairFormat string, minWidth, conf int) DecoratorFunc {
	return Counters(pairFormat, Unit_KiB, minWidth, conf)
}

// CountersKiloByte returns human friendly byte counters decorator, where
// counters unit is multiple by 1000.
// `pairFormat` must contain two printf compatible verbs, like "%f" or "%d".
// First verb substituted with Current, second one with Total.
// Example: `"%.1f / %.1f" = "1.0MiB / 12.0MiB"` or `"% .1f / % .1f" = "1.0 MiB / 12.0 MiB"`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func CountersKiloByte(pairFormat string, minWidth, conf int) DecoratorFunc {
	return Counters(pairFormat, Unit_kB, minWidth, conf)
}

// Counters provides basic counters decorator.
// `pairFormat` must contain two printf compatible verbs, like "%f" or "%d".
// First verb substituted with Current, second one with Total.
// Example: `"%.1f / %.1f" = "1.0MiB / 12.0MiB"` or `"% .1f / % .1f" = "1.0 MiB / 12.0 MiB"`.
// `unit` is one of decor.Unit_KiB/decor.Unit_kB or just zero if you need raw unitless numbers.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func Counters(pairFormat string, unit Unit, minWidth, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		switch unit {
		case Unit_KiB:
			str = fmt.Sprintf(pairFormat, CounterKiB(s.Current), CounterKiB(s.Total))
		case Unit_kB:
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
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	}
}

// ETA provides exponential-weighted-moving-average ETA decorator.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func ETA(minWidth, conf int) DecoratorFunc {
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
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	}
}

// Elapsed provides elapsed time decorator.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func Elapsed(minWidth, conf int) DecoratorFunc {
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
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	}
}

// Percentage provides percentage decorator.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func Percentage(minWidth, conf int) DecoratorFunc {
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
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	}
}

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
