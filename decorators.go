package mpb

import (
	"fmt"
	"time"
	"unicode/utf8"
)

const (
	// DidentRight specifies identation direction.
	// |   foo|     b| Without DidentRight
	// |foo   |b     | With DidentRight
	DidentRight = 1 << iota

	// DwidthSync will auto sync max width
	DwidthSync

	// DextraSpace adds extra space, makes sence with DwidthSync only.
	// When DidentRight bit set, the space will be added to the right,
	// otherwise to the left.
	DextraSpace
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string

// PrependName prepends name argument to the bar.
// The conf argument defines the formatting properties
func (b *Bar) PrependName(name string, minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.PrependFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(name)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), name)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), name)
	})
	return b
}

func (b *Bar) PrependCounters(pairFormat string, unit Units, minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.PrependFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		current := Format(s.Current).To(unit)
		total := Format(s.Total).To(unit)
		str := fmt.Sprintf(pairFormat, current, total)
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) PrependETA(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.PrependFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprint(time.Duration(s.Eta().Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) AppendETA(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.AppendFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprint(time.Duration(s.Eta().Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) PrependElapsed(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.PrependFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) AppendElapsed(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.AppendFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) AppendPercentage(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.AppendFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprintf("%d %%", percentage(s.Total, s.Current, 100))
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}

func (b *Bar) PrependPercentage(minWidth int, conf byte) *Bar {
	format := "%%"
	if (conf & DidentRight) != 0 {
		format += "-"
	}
	format += "%ds"
	b.PrependFunc(func(s *Statistics, myWidth chan<- int, maxWidth <-chan int) string {
		str := fmt.Sprintf("%d %%", percentage(s.Total, s.Current, 100))
		if (conf & DwidthSync) != 0 {
			myWidth <- utf8.RuneCountInString(str)
			max := <-maxWidth
			if (conf & DextraSpace) != 0 {
				max++
			}
			return fmt.Sprintf(fmt.Sprintf(format, max), str)
		}
		return fmt.Sprintf(fmt.Sprintf(format, minWidth), str)
	})
	return b
}
