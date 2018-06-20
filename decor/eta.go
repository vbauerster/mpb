package decor

import (
	"fmt"
	"math"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/vbauerster/mpb/internal"
)

// EwmaETA exponential-weighted-moving-average based ETA decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`age` is the previous N samples to average over.
//
//	`wcc` optional WC config
func EwmaETA(style int, age float64, wcc ...WC) Decorator {
	return MovingAverageETA(style, ewma.NewMovingAverage(age), wcc...)
}

// MovingAverageETA decorator relies on MovingAverage implementation to calculate its average.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`average` MovingAverage implementation
//
//	`wcc` optional WC config
func MovingAverageETA(style int, average MovingAverage, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	d := &movingAverageETA{
		style:   style,
		wc:      wc,
		average: average,
	}
	return d
}

type movingAverageETA struct {
	style      int
	wc         WC
	average    ewma.MovingAverage
	onComplete *struct {
		msg string
		wc  WC
	}
}

func (s *movingAverageETA) Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if st.Completed && s.onComplete != nil {
		return s.onComplete.wc.FormatMsg(s.onComplete.msg, widthAccumulator, widthDistributor)
	}

	v := internal.Round(s.average.Value())
	if math.IsInf(v, 0) || math.IsNaN(v) {
		v = 0
	}
	remaining := time.Duration((st.Total - st.Current) * int64(v))
	hours := int64((remaining / time.Hour) % 60)
	minutes := int64((remaining / time.Minute) % 60)
	seconds := int64((remaining / time.Second) % 60)

	var str string
	switch s.style {
	case ET_STYLE_GO:
		str = fmt.Sprint(time.Duration(remaining.Seconds()) * time.Second)
	case ET_STYLE_HHMMSS:
		str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case ET_STYLE_HHMM:
		str = fmt.Sprintf("%02d:%02d", hours, minutes)
	case ET_STYLE_MMSS:
		str = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return s.wc.FormatMsg(str, widthAccumulator, widthDistributor)
}

func (s *movingAverageETA) NextAmount(n int, wdd ...time.Duration) {
	var workDuration time.Duration
	for _, wd := range wdd {
		workDuration = wd
	}
	lastItemEstimate := float64(workDuration) / float64(n)
	s.average.Add(lastItemEstimate)
}

func (s *movingAverageETA) OnCompleteMessage(msg string, wcc ...WC) {
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

// AverageETA decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`wcc` optional WC config
func AverageETA(style int, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	startTime := time.Now()
	return DecoratorFunc(func(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		timeElapsed := time.Since(startTime)
		v := internal.Round(float64(timeElapsed) / float64(st.Current))
		if math.IsInf(v, 0) || math.IsNaN(v) {
			v = 0
		}
		remaining := time.Duration((st.Total - st.Current) * int64(v))
		hours := int64((remaining / time.Hour) % 60)
		minutes := int64((remaining / time.Minute) % 60)
		seconds := int64((remaining / time.Second) % 60)

		switch style {
		case ET_STYLE_GO:
			str = fmt.Sprint(time.Duration(remaining.Seconds()) * time.Second)
		case ET_STYLE_HHMMSS:
			str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		case ET_STYLE_HHMM:
			str = fmt.Sprintf("%02d:%02d", hours, minutes)
		case ET_STYLE_MMSS:
			str = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
		return wc.FormatMsg(str, widthAccumulator, widthDistributor)
	})
}
