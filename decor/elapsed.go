package decor

import (
	"fmt"
	"time"
)

// Elapsed returns elapsed time decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`wcc` optional WC config
func Elapsed(style int, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	startTime := time.Now()
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		var str string
		timeElapsed := time.Since(startTime)
		hours := int64((timeElapsed / time.Hour) % 60)
		minutes := int64((timeElapsed / time.Minute) % 60)
		seconds := int64((timeElapsed / time.Second) % 60)

		switch style {
		case ET_STYLE_GO:
			str = fmt.Sprint(time.Duration(timeElapsed.Seconds()) * time.Second)
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
