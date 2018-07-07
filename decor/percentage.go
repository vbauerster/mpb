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
	wc.Init()
	d := &percentageDecorator{
		WC: wc,
	}
	return d
}

type percentageDecorator struct {
	WC
	complete *completeMsg
}

func (d *percentageDecorator) Decor(st *Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(d.complete.msg)
	}
	str := fmt.Sprintf("%d %%", internal.Percentage(st.Total, st.Current, 100))
	return d.FormatMsg(str)
}

func (d *percentageDecorator) OnCompleteMessage(msg string) {
	d.complete = &completeMsg{msg}
}
