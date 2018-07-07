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
	wc.Init()
	d := &movingAverageETA{
		WC:      wc,
		style:   style,
		average: average,
	}
	return d
}

type movingAverageETA struct {
	WC
	style    int
	average  ewma.MovingAverage
	complete *completeMsg
}

func (d *movingAverageETA) Decor(st *Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(d.complete.msg)
	}

	v := internal.Round(d.average.Value())
	if math.IsInf(v, 0) || math.IsNaN(v) {
		v = 0
	}
	remaining := time.Duration((st.Total - st.Current) * int64(v))
	hours := int64((remaining / time.Hour) % 60)
	minutes := int64((remaining / time.Minute) % 60)
	seconds := int64((remaining / time.Second) % 60)

	var str string
	switch d.style {
	case ET_STYLE_GO:
		str = fmt.Sprint(time.Duration(remaining.Seconds()) * time.Second)
	case ET_STYLE_HHMMSS:
		str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case ET_STYLE_HHMM:
		str = fmt.Sprintf("%02d:%02d", hours, minutes)
	case ET_STYLE_MMSS:
		str = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return d.FormatMsg(str)
}

func (d *movingAverageETA) NextAmount(n int, wdd ...time.Duration) {
	var workDuration time.Duration
	for _, wd := range wdd {
		workDuration = wd
	}
	lastItemEstimate := float64(workDuration) / float64(n)
	d.average.Add(lastItemEstimate)
}

func (d *movingAverageETA) OnCompleteMessage(msg string) {
	d.complete = &completeMsg{msg}
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
	wc.Init()
	d := &averageETA{
		WC:        wc,
		style:     style,
		startTime: time.Now(),
	}
	return d
}

type averageETA struct {
	WC
	style     int
	startTime time.Time
	complete  *completeMsg
}

func (d *averageETA) Decor(st *Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(d.complete.msg)
	}

	var str string
	timeElapsed := time.Since(d.startTime)
	v := internal.Round(float64(timeElapsed) / float64(st.Current))
	if math.IsInf(v, 0) || math.IsNaN(v) {
		v = 0
	}
	remaining := time.Duration((st.Total - st.Current) * int64(v))
	hours := int64((remaining / time.Hour) % 60)
	minutes := int64((remaining / time.Minute) % 60)
	seconds := int64((remaining / time.Second) % 60)

	switch d.style {
	case ET_STYLE_GO:
		str = fmt.Sprint(time.Duration(remaining.Seconds()) * time.Second)
	case ET_STYLE_HHMMSS:
		str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case ET_STYLE_HHMM:
		str = fmt.Sprintf("%02d:%02d", hours, minutes)
	case ET_STYLE_MMSS:
		str = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return d.FormatMsg(str)
}

func (d *averageETA) OnCompleteMessage(msg string) {
	d.complete = &completeMsg{msg}
}
