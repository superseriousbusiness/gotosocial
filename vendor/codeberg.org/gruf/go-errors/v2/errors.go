package errors

import (
	"fmt"
	"runtime"
)

// New returns a new error created from message.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func New(msg string) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(2, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errormsg{
		cfn: c,
		msg: msg,
		trc: t,
	}
}

// Newf returns a new error created from message format and args.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func Newf(msgf string, args ...interface{}) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(2, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errormsg{
		cfn: c,
		msg: fmt.Sprintf(msgf, args...),
		trc: t,
	}
}

// NewAt returns a new error created, skipping 'skip'
// frames for trace / caller information, from message.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func NewAt(skip int, msg string) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(skip+1, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(skip+1, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errormsg{
		cfn: c,
		msg: msg,
		trc: t,
	}
}

// Wrap will wrap supplied error within a new error created from message.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func Wrap(err error, msg string) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(2, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errorwrap{
		cfn: c,
		msg: msg,
		err: err,
		trc: t,
	}
}

// Wrapf will wrap supplied error within a new error created from message format and args.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func Wrapf(err error, msgf string, args ...interface{}) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(2, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errorwrap{
		cfn: c,
		msg: fmt.Sprintf(msgf, args...),
		err: err,
		trc: t,
	}
}

// WrapAt wraps error within new error created from message,
// skipping 'skip' frames for trace / caller information.
//
// Note this function cannot be inlined, to ensure expected
// and consistent behaviour in setting trace / caller info.
//
//go:noinline
func WrapAt(skip int, err error, msg string) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(skip+1, pcs)
		fn := runtime.FuncForPC(pcs[0])
		c.set(funcName(fn))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, 10)
		n := runtime.Callers(skip+1, pcs)
		iter := runtime.CallersFrames(pcs[:n])
		t.set(gatherFrames(iter, n))
	}
	return &_errorwrap{
		cfn: c,
		msg: msg,
		err: err,
		trc: t,
	}
}

// Stacktrace fetches first stored stacktrace of callers from error chain.
func Stacktrace(err error) Callers {
	if !IncludesStacktrace {
		// compile-time check
		return nil
	}
	if e := As[*_errormsg](err); err != nil {
		return e.trc.value()
	}
	if e := As[*_errorwrap](err); err != nil {
		return e.trc.value()
	}
	return nil
}

type _errormsg struct {
	cfn caller
	msg string
	trc trace
}

func (err *_errormsg) Error() string {
	if IncludesCaller {
		fn := err.cfn.value()
		return fn + " " + err.msg
	} else {
		return err.msg
	}
}

func (err *_errormsg) Is(other error) bool {
	oerr, ok := other.(*_errormsg)
	return ok && oerr.msg == err.msg
}

type _errorwrap struct {
	cfn caller
	msg string
	err error // wrapped
	trc trace
}

func (err *_errorwrap) Error() string {
	if IncludesCaller {
		fn := err.cfn.value()
		return fn + " " + err.msg + ": " + err.err.Error()
	} else {
		return err.msg + ": " + err.err.Error()
	}
}

func (err *_errorwrap) Is(other error) bool {
	oerr, ok := other.(*_errorwrap)
	return ok && oerr.msg == err.msg
}

func (err *_errorwrap) Unwrap() error {
	return err.err
}
