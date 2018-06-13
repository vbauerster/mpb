package decor

// StaticName returns name decorator.
//
//	`name` string to display
//
//	`wcc` optional WC config
func StaticName(name string, wcc ...WC) Decorator {
	return Name(name, wcc...)
}

// Name returns name decorator.
//
//	`name` string to display
//
//	`wcc` optional WC config
func Name(name string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.BuildFormat()
	return DecoratorFunc(func(s *Statistics, widthAccumulator chan<- int, widthDistributor <-chan int) string {
		return wc.FormatMsg(name, widthAccumulator, widthDistributor)
	})
}
