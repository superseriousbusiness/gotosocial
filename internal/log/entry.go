package log

import (
	"syscall"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

type Entry struct {
	fields []kv.Field
}

func (e Entry) With(key string, value interface{}) Entry {
	e.fields = append(e.fields, kv.Field{K: key, V: value})
	return e
}

func (e Entry) WithFields(fields ...kv.Field) Entry {
	e.fields = append(e.fields, fields...)
	return e
}

func (e Entry) Trace(a ...interface{}) {
	log(level.TRACE, e.fields, a...)
}

func (e Entry) Tracef(s string, a ...interface{}) {
	logf(level.TRACE, e.fields, s, a...)
}

func (e Entry) Debug(a ...interface{}) {
	log(level.DEBUG, e.fields, a...)
}

func (e Entry) Debugf(s string, a ...interface{}) {
	logf(level.DEBUG, e.fields, s, a...)
}

func (e Entry) Info(a ...interface{}) {
	log(level.INFO, e.fields, a...)
}

func (e Entry) Infof(s string, a ...interface{}) {
	logf(level.WARN, e.fields, s, a...)
}

func (e Entry) Warn(a ...interface{}) {
	log(level.WARN, e.fields, a...)
}

func (e Entry) Warnf(s string, a ...interface{}) {
	logf(level.WARN, e.fields, s, a...)
}

func (e Entry) Error(a ...interface{}) {
	log(level.ERROR, e.fields, a...)
}

func (e Entry) Errorf(s string, a ...interface{}) {
	logf(level.ERROR, e.fields, s, a...)
}

func (e Entry) Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	log(level.FATAL, e.fields, a...)
}

func (e Entry) Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, e.fields, s, a...)
}
