package decor

import (
	"fmt"
	"math"
	"sort"
	"sync"
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
// Contains: Total, Current, TimeElapsed, TimePerItemEstimate
type Statistics struct {
	ID                  int
	Completed           bool
	Total               int64
	Current             int64
	StartTime           time.Time
	TimeElapsed         time.Duration
	TimeRemaining       time.Duration
	TimePerItemEstimate time.Duration
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
		str := fmt.Sprint(time.Duration(s.TimeRemaining.Seconds()) * time.Second)
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

// SpeedNoUnit returns raw I/O operation speed decorator.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func SpeedNoUnit(unitFormat string, minWidth, conf int) DecoratorFunc {
	return Speed(unitFormat, 0, minWidth, conf)
}

// SpeedKibiByte returns human friendly I/O operation speed decorator,
// where counters unit is multiple by 1024.
// `unitFormat` must contain one printf compatible verb, like "%f" or "%d".
// Example: `"%.1f" = "1.0MiB/s"` or `"% .1f" = "1.0 MiB/s"`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func SpeedKibiByte(unitFormat string, minWidth, conf int) DecoratorFunc {
	return Speed(unitFormat, unitKiB, minWidth, conf)
}

// SpeedKiloByte returns human friendly I/O operation speed decorator,
// where counters unit is multiple by 1000.
// `unitFormat` must contain one printf compatible verb, like "%f" or "%d".
// Example: `"%.1f" = "1.0MiB/s"` or `"% .1f" = "1.0 MiB/s"`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func SpeedKiloByte(unitFormat string, minWidth, conf int) DecoratorFunc {
	return Speed(unitFormat, unitKB, minWidth, conf)
}

// Speed provides basic I/O operation speed decorator.
// `unitFormat` must contain one printf compatible verb, like "%f" or "%d".
// Example: `"%.1f" = "1.0MiB/s"` or `"% .1f" = "1.0 MiB/s"`.
// If you set `DwidthSync` bit in `conf` param, `minWidth` param is ignored.
// `DwidthSync` is effective with multiple bars only, if set decorator will participate
// in width synchronization process with other decorators in the same column group.
func Speed(unitFormat string, unit, minWidth, conf int) DecoratorFunc {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"

	metricsMu := sync.RWMutex{}
	metrics := make(map[int]*pbSpeedMetrics)

	return func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string

		metricsMu.RLock()
		barMetrics, present := metrics[s.ID]
		metricsMu.RUnlock()
		if !present {
			barMetrics = new(pbSpeedMetrics)
			metricsMu.Lock()
			metrics[s.ID] = barMetrics
			metricsMu.Unlock()
		}

		if !s.Completed {
			// here we use integrated speed instead of instantaneous speed to make value oscillations more smooth
			// also we dropping "invalid" measurements like NaN, +Inf, -Inf
			instantSpeed := float64(s.Current-barMetrics.prevCount) / (s.TimeElapsed - barMetrics.prevDuration).Seconds() // bytes per second
			if !math.IsNaN(instantSpeed) && !math.IsInf(instantSpeed, 1) && !math.IsInf(instantSpeed, -1) {
				barMetrics.PutSpeed(instantSpeed)
			}
			barMetrics.prevCount = s.Current
			barMetrics.prevDuration = s.TimeElapsed
		}
		speed := barMetrics.SpeedMedian()

		switch unit {
		case unitKiB:
			str = fmt.Sprintf(unitFormat, SpeedKiB(speed))
		case unitKB:
			str = fmt.Sprintf(unitFormat, SpeedKB(speed))
		default:
			str = fmt.Sprintf(unitFormat, speed)
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

const speedWindowSize = 3

type pbSpeedMetrics struct {
	prevCount    int64
	prevDuration time.Duration
	speedWindow  [speedWindowSize]float64
}

func (p *pbSpeedMetrics) PutSpeed(speed float64) {
	for i := 0; i < len(p.speedWindow)-1; i++ {
		p.speedWindow[i] = p.speedWindow[i+1]
	}
	p.speedWindow[len(p.speedWindow)-1] = speed
}

func (p *pbSpeedMetrics) SpeedMedian() float64 {
	var temp [speedWindowSize]float64
	copy(temp[:], p.speedWindow[:])
	sort.Float64s(temp[:])
	if l := len(temp) % 2; l != 0 {
		return temp[l]
	} else {
		return (temp[l] + temp[l+1]) / 2
	}
}
