package mpb

import (
	"fmt"
	"strconv"
	"time"
)

type decoratorFuncType uint

const (
	decoratorAppend decoratorFuncType = iota
	decoratorPrepend
)

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(s *Statistics) string

type decorator struct {
	kind decoratorFuncType
	f    DecoratorFunc
}

func (b *Bar) PrependName(name string, padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, name)
	})
	return b
}

func (b *Bar) PrependCounters(unit Units, padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		current := Format(s.Current).To(unit)
		total := Format(s.Total).To(unit)
		str := fmt.Sprintf("%s / %s", current, total)
		return fmt.Sprintf(layout, str)
	})
	return b
}

func (b *Bar) PrependETA(padding int) *Bar {
	layout := "ETA% " + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, time.Duration(s.Eta().Seconds())*time.Second)
	})
	return b
}

func (b *Bar) AppendETA(padding int) *Bar {
	layout := "ETA% " + strconv.Itoa(padding) + "s"
	b.AppendFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, time.Duration(s.Eta().Seconds())*time.Second)
	})
	return b
}

func (b *Bar) PrependElapsed(padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "s"
	b.PrependFunc(func(s *Statistics) string {
		return fmt.Sprintf(layout, time.Duration(s.TimeElapsed.Seconds())*time.Second)
	})
	return b
}

func (b *Bar) AppendElapsed() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		return fmt.Sprint(time.Duration(s.TimeElapsed.Seconds()) * time.Second)
	})
	return b
}

func (b *Bar) AppendPercentage() *Bar {
	b.AppendFunc(func(s *Statistics) string {
		completed := int(100 * float64(s.Current) / float64(s.Total))
		return fmt.Sprintf("%3d %%", completed)
	})
	return b
}

func (b *Bar) PrependPercentage(padding int) *Bar {
	layout := "%" + strconv.Itoa(padding) + "d %%"
	b.PrependFunc(func(s *Statistics) string {
		completed := int(100 * float64(s.Current) / float64(s.Total))
		return fmt.Sprintf(layout, completed)
	})
	return b
}
