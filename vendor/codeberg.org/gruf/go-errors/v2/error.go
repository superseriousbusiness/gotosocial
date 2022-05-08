//go:build !notrace
// +build !notrace

package errors

type errormsg struct {
	msg   string
	wrap  error
	stack Callers
}

func create(msg string, wrap error) *errormsg {
	return &errormsg{
		msg:   msg,
		wrap:  wrap,
		stack: GetCallers(2, 10),
	}
}

func (err *errormsg) Error() string {
	return err.msg
}

func (err *errormsg) Is(target error) bool {
	other, ok := target.(*errormsg)
	return ok && (err.msg == other.msg)
}

func (err *errormsg) Unwrap() error {
	return err.wrap
}

func (err *errormsg) Stacktrace() Callers {
	return err.stack
}
