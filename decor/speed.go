package decor

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/VividCortex/ewma"
)

type SpeedKiB float64

func (s SpeedKiB) Format(st fmt.State, verb rune) {
	prec, ok := st.Precision()

	if verb == 'd' || !ok {
		prec = 0
	}
	if verb == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if verb == 's' {
		prec = 1
	}

	var res, unit string
	switch {
	case s >= TiB:
		unit = "TiB/s"
		res = strconv.FormatFloat(float64(s)/TiB, 'f', prec, 64)
	case s >= GiB:
		unit = "GiB/s"
		res = strconv.FormatFloat(float64(s)/GiB, 'f', prec, 64)
	case s >= MiB:
		unit = "MiB/s"
		res = strconv.FormatFloat(float64(s)/MiB, 'f', prec, 64)
	case s >= KiB:
		unit = "KiB/s"
		res = strconv.FormatFloat(float64(s)/KiB, 'f', prec, 64)
	default:
		unit = "b/s"
		res = strconv.FormatInt(int64(s), 10)
	}

	if st.Flag(' ') {
		res += " "
	}
	res += unit

	if w, ok := st.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if st.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	io.WriteString(st, res)
}

type SpeedKB float64

func (s SpeedKB) Format(st fmt.State, verb rune) {
	prec, ok := st.Precision()

	if verb == 'd' || !ok {
		prec = 0
	}
	if verb == 'f' && !ok {
		prec = 6
	}
	// retain old beahavior if s verb used
	if verb == 's' {
		prec = 1
	}

	var res, unit string
	switch {
	case s >= TB:
		unit = "TB/s"
		res = strconv.FormatFloat(float64(s)/TB, 'f', prec, 64)
	case s >= GB:
		unit = "GB/s"
		res = strconv.FormatFloat(float64(s)/GB, 'f', prec, 64)
	case s >= MB:
		unit = "MB/s"
		res = strconv.FormatFloat(float64(s)/MB, 'f', prec, 64)
	case s >= KB:
		unit = "kB/s"
		res = strconv.FormatFloat(float64(s)/KB, 'f', prec, 64)
	default:
		unit = "b/s"
		res = strconv.FormatInt(int64(s), 10)
	}

	if st.Flag(' ') {
		res += " "
	}
	res += unit

	if w, ok := st.Width(); ok {
		if len(res) < w {
			pad := strings.Repeat(" ", w-len(res))
			if st.Flag(int('-')) {
				res += pad
			} else {
				res = pad + res
			}
		}
	}

	io.WriteString(st, res)
}

// EwmaSpeed exponential-weighted-moving-average based speed decorator,
// with dynamic unit measure adjustement.
//
//	`unit` one of [0|UnitKiB|UnitKB] zero for no unit
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`average` MovingAverage implementation
//
//	`sb` is a start block receive channel. It's required by MovingAverage algorithm,
//	 therefore result of time.Now() must be sent to this channel on each iteration
//	 of a start block, right before the actual job. There is no need to close the channel,
//	 as it will be closed automatically on bar completion event.
//
//	`wcc` optional WC config
func EwmaSpeed(unit int, unitFormat string, age float64, sb chan time.Time, wcc ...WC) Decorator {
	return MovingAverageSpeed(unit, unitFormat, ewma.NewMovingAverage(age), sb, wcc...)
}

// MovingAverageSpeed decorator relies on MovingAverage implementation to calculate its average.
//
//	`unit` one of [0|UnitKiB|UnitKB] zero for no unit
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`average` MovingAverage implementation
//
//	`sb` is a start block receive channel. It's required by MovingAverage algorithm,
//	 therefore result of time.Now() must be sent to this channel on each iteration
//	 of a start block, right before the actual job. There is no need to close the channel,
//	 as it will be closed automatically on bar completion event.
//
//	`wcc` optional WC config
func MovingAverageSpeed(unit int, unitFormat string, average MovingAverage, sb chan time.Time, wcc ...WC) Decorator {
	if sb == nil {
		panic("start block channel must not be nil")
	}
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	d := &movingAverageSpeed{
		unit:       unit,
		unitFormat: unitFormat,
		wc:         wc,
		average:    average,
		sbReceiver: sb,
		sbStreamer: make(chan time.Time),
	}
	go d.serve()
	return d
}

type movingAverageSpeed struct {
	unit       int
	unitFormat string
	wc         WC
	average    ewma.MovingAverage
	sbReceiver chan time.Time
	sbStreamer chan time.Time
	onComplete *struct {
		msg string
		wc  WC
	}
}

func (s *movingAverageSpeed) Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if st.Completed && s.onComplete != nil {
		return s.onComplete.wc.FormatMsg(s.onComplete.msg, widthAccumulator, widthDistributor)
	}
	var str string
	speed := s.average.Value()
	switch s.unit {
	case UnitKiB:
		str = fmt.Sprintf(s.unitFormat, SpeedKiB(speed))
	case UnitKB:
		str = fmt.Sprintf(s.unitFormat, SpeedKB(speed))
	default:
		str = fmt.Sprintf(s.unitFormat, speed)
	}
	return s.wc.FormatMsg(str, widthAccumulator, widthDistributor)
}

func (s *movingAverageSpeed) NextAmount(n int) {
	sb := <-s.sbStreamer
	speed := float64(n) / time.Since(sb).Seconds()
	s.average.Add(speed)
}

func (s *movingAverageSpeed) OnCompleteMessage(msg string, wcc ...WC) {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	s.onComplete = &struct {
		msg string
		wc  WC
	}{msg, wc}
}

func (s *movingAverageSpeed) Shutdown() {
	close(s.sbReceiver)
}

func (s *movingAverageSpeed) serve() {
	for now := range s.sbReceiver {
		s.sbStreamer <- now
	}
}

// TotalAverageSpeed decorator with dynamic unit measure adjustement.
//
//	`unit` one of [0|UnitKiB|UnitKB] zero for no unit
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`wcc` optional WC config
func TotalAverageSpeed(unit int, unitFormat string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	startTime := time.Now()
	return DecoratorFunc(func(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		timeElapsed := time.Since(startTime)
		speed := float64(st.Current) / timeElapsed.Seconds()

		switch unit {
		case UnitKiB:
			str = fmt.Sprintf(unitFormat, SpeedKiB(speed))
		case UnitKB:
			str = fmt.Sprintf(unitFormat, SpeedKB(speed))
		default:
			str = fmt.Sprintf(unitFormat, speed)
		}
		return wc.FormatMsg(str, widthAccumulator, widthDistributor)
	})
}
