package decor

import (
	"fmt"
	"time"

	"github.com/VividCortex/ewma"
)

// ETA returns exponential-weighted-moving-average ETA decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`age` is the previous N samples to average over.
//	 If zero value provided, it defaults to 30.
//
//	`sbCh` is a start block receive channel. User suppose to send time.Now()
//	 to this channel on each iteration of a start block, right before actual job.
//	 The channel will be auto closed on bar shutdown event, so there is no need
//	 to close from user side.
//
//	`wcc` optional WC config
func ETA(style int, age float64, sbCh chan time.Time, wcc ...WC) Decorator {
	if sbCh == nil {
		panic("start block channel must not be nil")
	}
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	if age == .0 {
		age = ewma.AVG_METRIC_AGE
	}
	d := &ewmaETA{
		style:      style,
		wc:         wc,
		mAverage:   ewma.NewMovingAverage(age),
		sbReceiver: sbCh,
		sbStreamer: make(chan time.Time),
	}
	go d.serve()
	return d
}

type ewmaETA struct {
	style      int
	wc         WC
	mAverage   ewma.MovingAverage
	sbReceiver chan time.Time
	sbStreamer chan time.Time
	onComplete *struct {
		msg string
		wc  WC
	}
}

func (s *ewmaETA) Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if st.Completed && s.onComplete != nil {
		return s.onComplete.wc.FormatMsg(s.onComplete.msg, widthAccumulator, widthDistributor)
	}

	var str string
	timeRemaining := time.Duration(st.Total-st.Current) * time.Duration(round(s.mAverage.Value()))
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

	return s.wc.FormatMsg(str, widthAccumulator, widthDistributor)
}

func (s *ewmaETA) NextAmount(n int) {
	sb := <-s.sbStreamer
	lastBlockTime := time.Since(sb)
	lastItemEstimate := float64(lastBlockTime) / float64(n)
	s.mAverage.Add(lastItemEstimate)
}

func (s *ewmaETA) OnCompleteMessage(msg string, wcc ...WC) {
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

func (s *ewmaETA) Shutdown() {
	close(s.sbReceiver)
}

func (s *ewmaETA) serve() {
	for now := range s.sbReceiver {
		s.sbStreamer <- now
	}
}
