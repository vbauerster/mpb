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
//	`age` is related to the decay factor alpha by the formula given for the DECAY constant.
//	 It signifies the average age of the samples as time goes to infinity. Basically age is
//	 the previous N samples to average over. If zero value provided, it defaults to 30.
//
//	`startBlock` is a time.Time receive channel. User suppose to send time.Now()
//   to this channel on each iteration of block start, right before actual job.
//   The channel will be closed automatically on bar shutdown event, so there is
//   no need to close from user side.
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
	eta := &ewmaETA{
		mAverage:           ewma.NewMovingAverage(age),
		startBlockReceiver: startBlock,
		startBlockStreamer: make(chan time.Time),
		style:              style,
		wc:                 wc0,
	}
	go eta.serve()
	return eta
}

type ewmaETA struct {
	mAverage           ewma.MovingAverage
	startBlockReceiver chan time.Time
	startBlockStreamer chan time.Time
	style              int
	wc                 WC
	onComplete         *struct {
		msg string
		wc  WC
	}
}

func (s *ewmaETA) Decor(st *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
	if st.Completed && s.onComplete != nil {
		return s.onComplete.wc.formatMsg(s.onComplete.msg, widthAccumulator, widthDistributor)
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

	return s.wc.formatMsg(str, widthAccumulator, widthDistributor)
}

func (s *ewmaETA) NextAmount(n int) {
	sb := <-s.startBlockStreamer
	lastBlockTime := time.Since(sb)
	lastItemEstimate := float64(lastBlockTime) / float64(n)
	s.mAverage.Add(lastItemEstimate)
}

func (s *ewmaETA) OnCompleteMessage(msg string, wc ...WC) {
	var wc0 WC
	if len(wc) > 0 {
		wc0 = wc[0]
	}
	s.onComplete = &struct {
		msg string
		wc  WC
	}{msg, wc0}
}

func (s *ewmaETA) Shutdown() {
	close(s.startBlockReceiver)
}

func (s *ewmaETA) serve() {
	for now := range s.startBlockReceiver {
		s.startBlockStreamer <- now
	}
}
