package decor

import (
	"fmt"

	"github.com/vbauerster/mpb/internal"
)

// Percentage returns percentage decorator.
//
//	`wcc` optional WC config
func Percentage(wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		str := fmt.Sprintf("%d %%", internal.Percentage(s.Total, s.Current, 100))
		return wc.FormatMsg(str, widthAccumulator, widthDistributor)
	})
}
