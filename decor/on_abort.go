package decor

var (
	_ Decorator = onAbortWrapper{}
	_ Wrapper   = onAbortWrapper{}
)

// OnAbort wrap decorator.
// Displays provided message on abort event.
// Has no effect if bar.Abort(true) is called.
//
//	`decorator` Decorator to wrap
//	`message` message to display
func OnAbort(decorator Decorator, message string) Decorator {
	if decorator == nil {
		return nil
	}
	return onAbortWrapper{decorator, message}
}

type onAbortWrapper struct {
	Decorator
	msg string
}

func (d onAbortWrapper) Decor(s Statistics) (string, int) {
	if s.Aborted {
		return d.Format(d.msg)
	}
	return d.Decorator.Decor(s)
}

func (d onAbortWrapper) Unwrap() Decorator {
	return d.Decorator
}
