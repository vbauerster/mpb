package decor

import "fmt"

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
		str := fmt.Sprintf("%d %%", CalcPercentage(s.Total, s.Current, 100))
		return wc.FormatMsg(str, widthAccumulator, widthDistributor)
	})
}

// CalcPercentage is a helper function, to calculate percentage.
func CalcPercentage(total, current, width int64) int64 {
	if total <= 0 {
		return 0
	}
	p := float64(width*current) / float64(total)
	return int64(round(p))
}
