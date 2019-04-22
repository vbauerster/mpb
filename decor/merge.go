package decor

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func Merge(decorator Decorator, wcc ...WC) Decorator {
	if _, ok := decorator.Sync(); !ok {
		return decorator
	}
	var placeHolders []*placeHolderDecorator
	for _, wc := range wcc {
		wc.Init()
		placeHolders = append(placeHolders, &placeHolderDecorator{
			WC:    wc,
			wsync: make(chan int),
		})
	}
	md := &MergeDecorator{
		decorator:    decorator,
		placeHolders: placeHolders,
	}
	md.WC = decorator.SetConfig(md.WC)
	return md
}

type MergeDecorator struct {
	WC
	decorator    Decorator
	placeHolders []*placeHolderDecorator
}

func (d *MergeDecorator) PlaceHolders() []Decorator {
	decorators := make([]Decorator, len(d.placeHolders))
	for i, ph := range d.placeHolders {
		decorators[i] = ph
	}
	return decorators
}

func (d *MergeDecorator) Decor(st *Statistics) string {
	msg := d.decorator.Decor(st)
	msgLen := utf8.RuneCountInString(msg)
	pWidth := msgLen / (len(d.placeHolders) + 1)
	mod := msgLen % (len(d.placeHolders) + 1)
	d.wsync <- pWidth + mod
	for _, ph := range d.placeHolders {
		ph.wsync <- pWidth
	}
	// fmt.Fprintln(os.Stderr, "all sent")
	max := <-d.wsync
	for _, ph := range d.placeHolders {
		max += <-ph.wsync
	}
	if (d.C & DextraSpace) != 0 {
		max++
	}
	return fmt.Sprintf(fmt.Sprintf(d.format, max), msg)
}

type placeHolderDecorator struct {
	WC
	wsync chan int
}

func (d *placeHolderDecorator) Decor(st *Statistics) string {
	go func() {
		width := <-d.wsync
		msg := strings.Repeat(" ", width)
		d.wsync <- utf8.RuneCountInString(d.FormatMsg(msg))
	}()
	return ""
}
