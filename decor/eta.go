package decor

import (
	"fmt"
	"math"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/vbauerster/mpb/internal"
)

// ETA returns exponential-weighted-moving-average ETA decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
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
func ETA(style int, age float64, sb chan time.Time, wcc ...WC) Decorator {
	return MovingAverageETA(style, ewma.NewMovingAverage(age), sb, wcc...)
}

// MovingAverageETA returns ETA decorator, which relies on MovingAverage implementation to calculate average.
// Default ETA decorator relies on ewma implementation. However you're free to provide your own implementation
// or use alternative one, which is provided by decor package:
//
//	decor.MovingAverageETA(decor.ET_STYLE_GO, decor.NewMedian(), sb)
func MovingAverageETA(style int, average MovingAverage, sb chan time.Time, wcc ...WC) Decorator {
	if sb == nil {
		panic("start block channel must not be nil")
	}
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	d := &movingAverageETA{
		style:      style,
		wc:         wc,
		average:    average,
		sbReceiver: sb,
		sbStreamer: make(chan time.Time),
	}
	go d.serve()
	return d
}

type movingAverageETA struct {
	style      int
	wc         WC
	average    ewma.MovingAverage
	sbReceiver chan time.Time
	sbStreamer chan time.Time
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
		v = .0
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

func (s *movingAverageETA) NextAmount(n int) {
	sb := <-s.sbStreamer
	lastBlockTime := time.Since(sb)
	lastItemEstimate := float64(lastBlockTime) / float64(n)
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

func (s *movingAverageETA) Shutdown() {
	close(s.sbReceiver)
}

func (s *movingAverageETA) serve() {
	for now := range s.sbReceiver {
		s.sbStreamer <- now
	}
}
