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

// SpeedNoUnit returns raw I/O operation speed decorator.
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`age` is the previous N samples to average over.
//	 If zero value provided, it defaults to 30.
//
//	`sb` is a start block receive channel. User suppose to send time.Now()
//	 to this channel on each iteration of a start block, right before actual job.
//	 The channel will be auto closed on bar shutdown event, so there is no need
//	 to close from user side.
//
//	`wcc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0" or "% .1f" = "1.0"
func SpeedNoUnit(unitFormat string, age float64, sb chan time.Time, wcc ...WC) Decorator {
	return MovingAverageSpeed(0, unitFormat, ewma.NewMovingAverage(age), sb, wcc...)
}

// SpeedKibiByte returns human friendly I/O operation speed decorator,
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`age` is the previous N samples to average over.
//	 If zero value provided, it defaults to 30.
//
//	`sb` is a start block receive channel. User suppose to send time.Now()
//	 to this channel on each iteration of a start block, right before actual job.
//	 The channel will be auto closed on bar shutdown event, so there is no need
//	 to close from user side.
//
//	`wcc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0MiB/s" or "% .1f" = "1.0 MiB/s"
func SpeedKibiByte(unitFormat string, age float64, sb chan time.Time, wcc ...WC) Decorator {
	return MovingAverageSpeed(UnitKiB, unitFormat, ewma.NewMovingAverage(age), sb, wcc...)
}

// SpeedKiloByte returns human friendly I/O operation speed decorator,
//
//	`unitFormat` printf compatible verb for value, like "%f" or "%d"
//
//	`age` is the previous N samples to average over.
//	 If zero value provided, it defaults to 30.
//
//	`sb` is a start block receive channel. User suppose to send time.Now()
//	 to this channel on each iteration of a start block, right before actual job.
//	 The channel will be auto closed on bar shutdown event, so there is no need
//	 to close from user side.
//
//	`wcc` optional WC config
//
// unitFormat example:
//
//	"%.1f" = "1.0MB/s" or "% .1f" = "1.0 MB/s"
func SpeedKiloByte(unitFormat string, age float64, sb chan time.Time, wcc ...WC) Decorator {
	return MovingAverageSpeed(UnitKB, unitFormat, ewma.NewMovingAverage(age), sb, wcc...)
}

// MovingAverageSpeed returns Speed decorator, which relies on MovingAverage implementation to calculate average.
// Default Speed decorator relies on ewma implementation. However you're free to provide your own implementation
// or use alternative one, which is provided by decor package:
//
//	decor.MovingAverageSpeed(decor.UnitKiB, "% .2f", decor.NewMedian(), sb)
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
