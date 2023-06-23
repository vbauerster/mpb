package decor

var (
	_ Decorator = onCompleteWrapper{}
	_ Wrapper   = onCompleteWrapper{}
)

// OnComplete wrap decorator.
// Displays provided message on complete event.
//
//	`decorator` Decorator to wrap
//	`message` message to display
func OnComplete(decorator Decorator, message string) Decorator {
	if decorator == nil {
		return nil
	}
	return onCompleteWrapper{decorator, message}
}

type onCompleteWrapper struct {
	Decorator
	msg string
}

func (d onCompleteWrapper) Decor(s Statistics) (string, int) {
	if s.Completed {
		return d.Format(d.msg)
	}
	return d.Decorator.Decor(s)
}

func (d onCompleteWrapper) Unwrap() Decorator {
	return d.Decorator
}
