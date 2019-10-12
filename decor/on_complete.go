package decor

// OnComplete returns decorator, which wraps provided decorator, with
// sole purpose to display provided message on complete event.
//
//	`decorator` Decorator to wrap
//
//	`message` message to display on complete event
func OnComplete(decorator Decorator, message string) Decorator {
	d := &onCompleteWrapper{
		Decorator: decorator,
		wc:        decorator.GetConf(),
		msg:       message,
	}
	return d
}

type onCompleteWrapper struct {
	Decorator
	wc  WC
	msg string
}

func (d *onCompleteWrapper) Decor(st *Statistics) string {
	if st.Completed {
		return d.wc.FormatMsg(d.msg)
	}
	return d.Decorator.Decor(st)
}
