package decor

var defaultSpinnerStyle = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner returns spinner decorator.
//
//	`frames` spinner frames, if nil or len==0, default is used
//
//	`wcc` optional WC config
func Spinner(frames []string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	if len(frames) == 0 {
		frames = defaultSpinnerStyle
	}
	d := &spinnerDecorator{
		WC:     wc,
		frames: frames,
	}
	return d
}

type spinnerDecorator struct {
	WC
	frames   []string
	complete *string
}

func (d *spinnerDecorator) Decor(st *Statistics) string {
	if st.Completed && d.complete != nil {
		return d.FormatMsg(*d.complete)
	}
	frame := d.frames[st.Current%int64(len(d.frames))]
	return d.FormatMsg(frame)
}

func (d *spinnerDecorator) OnCompleteMessage(msg string) {
	d.complete = &msg
}
